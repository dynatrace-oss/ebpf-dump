// Copyright (C) 2025 Dynatrace LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package ebpf

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	v1alpha1 "ebpfdump/api/v1alpha1"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	// Is the eBPF program[s] loaded?
	Loaded bool = false
	// Pointer to the shared ringbuf reader between all loaded programs
	RingbuffReader *ringbuf.Reader = nil
	// eBPF objects
	Objs bpfObjects = bpfObjects{}
	// Map of all loaded XdpPrograms accessed by their interface name
	XdpPrograms map[string]link.Link = make(map[string]link.Link)
	// Contains a list of Interface Numbers that the operator is attached to
	InterfacesNums []int32 = []int32{}
	// Function where the kprobe gets attached to
	KprobedFunc string = "tcp_sendmsg"
	// The Kprobe program
	Kprobe link.Link = nil
	// Function to call after data is received, It is ignored if empty
	Callback string = ""
)

func parseResponse(data bpfLogData, message []byte) (v1alpha1.NetworkDump, error) {
	res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(message)), nil)
	if err != nil {
		return v1alpha1.NetworkDump{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return v1alpha1.NetworkDump{}, err
	}

	ipDotNotation := uint32ToDotNotation(data.IpDaddr.V4)
	return v1alpha1.NetworkDump{
		Direction:  data.Dir,
		TimingMs:   time.Now().UnixNano() / 1e6,
		RemoteIp:   ipDotNotation,
		RemotePort: ntoh(data.TcpDest),
		StatusCode: res.StatusCode,
		Version:    strconv.Itoa(res.ProtoMajor) + "." + strconv.Itoa(res.ProtoMinor),
		Headers:    res.Header,
		Body:       string(body),
	}, nil
}

func parseRequest(data bpfLogData, message []byte) (v1alpha1.NetworkDump, error) {
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(message)))
	if err != nil {
		return v1alpha1.NetworkDump{}, err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return v1alpha1.NetworkDump{}, err
	}

	ipDotNotation := uint32ToDotNotation(data.IpSaddr.V4)
	return v1alpha1.NetworkDump{
		Direction:  data.Dir,
		TimingMs:   time.Now().UnixNano() / 1e6,
		RemoteIp:   ipDotNotation,
		RemotePort: ntoh(data.TcpSource),
		Method:     req.Method,
		Path:       req.URL.String(),
		Version:    strconv.Itoa(req.ProtoMajor) + "." + strconv.Itoa(req.ProtoMinor),
		Headers:    req.Header,
		Body:       string(body),
	}, nil
}

// This function is responsible to output the eBPF data, either by
// logging or by making a POST request if a Callback is set
func dumpData(ctx context.Context, data bpfLogData, message []byte) {

	log := log.FromContext(ctx)

	// Check if we are listening to the data's interface
	found := false
	for _, ifnum := range InterfacesNums {
		if data.Ifindex == ifnum {
			found = true
			break
		}
	}
	if !found && data.Dir != 1 {
		return
	}

	dump, err := parseRequest(data, message)
	if err != nil {
		dump, err = parseResponse(data, message)
		if err != nil {
			return
		}
	}

	jsonDump, err := json.Marshal(dump)
	if err != nil {
		return
	}

	if Callback != "" {
		_, err = http.Post(Callback, "application/json", bufio.NewReader(bytes.NewReader(jsonDump)))
		if err != nil {
			log.Error(err, "error sending request to callback", "callback", Callback)
		}
	} else {
		log.Info("Traffic received", "data", string(jsonDump))
	}

	return
}

func LoadBpf(ctx context.Context, interfaces []string) error {

	log := log.FromContext(ctx)

	// Remove resource limits for kernels <5.11.
	err := rlimit.RemoveMemlock()
	if err != nil {
		return fmt.Errorf("Error removing memlock: %w", err)
	}

	if err = loadBpfObjects(&Objs, nil); err != nil {
		return fmt.Errorf("Error loading eBPF objects: %w", err)
	}

	// Close unused XDP programs
	for ifName, xdpProgram := range XdpPrograms {
		found := false
		for _, ifNameWanted := range interfaces {
			if ifName == ifNameWanted {
				found = true
				break
			}
		}
		if !found {
			if err := xdpProgram.Close(); err != nil {
				return fmt.Errorf("Error closing XDP program: %w", err)
			}
			delete(XdpPrograms, ifName)
		}
	}

	InterfacesNums = []int32{}
	for _, ifName := range interfaces {

		if _, ok := XdpPrograms[ifName]; ok {
			continue
		}

		iface, err := net.InterfaceByName(ifName)
		if err != nil {
			return fmt.Errorf("Error getting the interface number for interface %s:  %w", ifName, err)
		}
		InterfacesNums = append(InterfacesNums, int32(iface.Index))

		// Check if interface is up and running
		if iface.Flags&net.FlagUp == 0 {
			return fmt.Errorf("Error: interface %s is not up", ifName)
		}
		if iface.Flags&net.FlagRunning == 0 {
			return fmt.Errorf("Error: interface %s is not running", ifName)
		}

		XdpPrograms[ifName], err = link.AttachXDP(link.XDPOptions{
			Program:   Objs.TcpDump,
			Interface: iface.Index,
		})
		if err != nil {
			return fmt.Errorf("Could not attach XDP program: %w", err)
		}

		log.Info("Loaded XDP program", "ifName", ifName)
	}

	Kprobe, err = link.Kprobe(KprobedFunc, Objs.KprobeTcpSendmsg, nil)
	if err != nil {
		return fmt.Errorf("Could not load Kprobe: %w", err)
	}

	// Open a ringbuf reader from userspace RINGBUF map described in the
	// eBPF C program.
	RingbuffReader, err = ringbuf.NewReader(Objs.Rb)
	if err != nil {
		return fmt.Errorf("Error opening ringbuf reader: %w", err)
	}
	Loaded = true
	return nil
}

func UnloadBpf(ctx context.Context) {
	log := log.FromContext(ctx)
	for key, xdpProgram := range XdpPrograms {
		if err := xdpProgram.Close(); err != nil {
			log.Error(err, "Failed to close ebpf program")
			return
		}
		delete(XdpPrograms, key)
	}

	if Kprobe != nil {
		Kprobe.Close()
	}
	if RingbuffReader != nil {
		RingbuffReader.Close()
	}
	Objs.Close()
	Loaded = false

	log.Info("Ebpf program closed")
	return
}

func LogData(ctx context.Context, cli client.Reader) {

	log := log.FromContext(ctx)

	if RingbuffReader == nil {
		log.Error(nil, "Logger error: ringbuffer not inizialized")
		return
	}

	var data bpfLogData
	var prevData bpfLogData
	messageBuffer := []byte{}

	// Here we read the eBPF ring buffer to receive information about
	// the network traffic. TCP payloads may be too long to be sent from
	// the eBPF program to userspace as a single buffer, so they may be
	// split into multiple buffers and they need to be reconstructed. In
	// order to do this, we use the full message size and the sequence
	// number: if the message we received is less than the message size,
	// then we need to wait for more as long as the sequence number is
	// the same.
	for {
		record, err := RingbuffReader.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				log.Info("Received signal, exiting..")
				return
			}
			log.Error(err, "Error reading from reader")
			continue
		}

		// Parse the ringbuf data entry into a bpfLogData structure.
		err = binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &data)
		if err != nil {
			log.Error(err, "Error parsing ringbuf data")
			continue
		}

		if data.IpType == 0 { // IpV4
			message := int8ToByte(data.Data)

			if uint64(len(message)) < data.TcpPayloadSize { // Did we get the full message?
				if prevData.TcpSeq == data.TcpSeq { // Same sequence?
					messageBuffer = append(messageBuffer, message...)
					if uint64(len(messageBuffer)) >= data.TcpPayloadSize {
						dumpData(ctx, data, messageBuffer)
						messageBuffer = []byte{}
					}
				} else { // New sequence, do not send anything yet
					messageBuffer = message
				}
			} else { // Send the data
				dumpData(ctx, data, message)
				messageBuffer = []byte{}
			}

			prevData = data

		} else if data.IpType == 1 { // IPv6
			log.Info("IPv6 not supported")
		} else {
			log.Info("Layer 3 protocol is not IPv4 or IPv6")
		}
	}
}
