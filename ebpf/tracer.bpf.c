// Copyright (C) 2025 Dynatrace LLC
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License as
// published by the Free Software Foundation; version 2.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public
// License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor,
// Boston, MA 02110-1301, USA.
//
// SPDX-License-Identifier: GPL-2.0-only

//go:build ignore

/*
 *   File: tracer.bpf.c
 *   Original author: Giovanni Santini
 *   License: GPLv2
 */

#include "vmlinux.h"
#include "maps.h"
#include "log_data.h"
#include "license.h"

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

// - Constants -

/*
 *  IEEE 802.3 Ethernet magic constants, you can fins them in the
 *  kernel source at linux/include/uapi/linux/if_ether.h
 */
#define ETH_P_IP   0x0800    /* Internet Protocol packet */
#define ETH_P_IPV6 0x86DD    /* IPv6 over bluebook       */

/*
 *  Some IANA protocol numbers
 */
#define ICMP_PROTO_NUMBER 0x01
#define TCP_PROTO_NUMBER  0x06
#define UDP_PROTO_NUMBER  0x11
#define IPV6_PROTO_NUMBER 0x29  /* IPv6 encapsulation */
// ...

#define NETWORK_INTERFACE_MTU 1500  /* TODO: this depends on the machine and should be generated */
#define MAX_TCP_IPV4_DATA_SIZE (NETWORK_INTERFACE_MTU - sizeof(struct tcphdr) - sizeof(struct iphdr))

// - Macros -

/*
 * This macro converts a 16-bit number from host byte order to network
 * byte order. Apparently there are no kfuncs that do this and you
 * need to implement the macro manually, which I manually copied from
 * libbpf.
 */

#define ___bpf_mvb(x, b, n, m) ((__u##b)(x) << (b-(n+1)*8) >> (b-8) << (m*8))
#define ___bpf_swab16(x) ((__u16)(\
  ___bpf_mvb(x, 16, 0, 1) | \
  ___bpf_mvb(x, 16, 1, 0)))

#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
# define __bpf_ntohs(x) __builtin_bswap16(x)
# define __bpf_constant_ntohs(x)  ___bpf_swab16(x)
#elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
# define __bpf_ntohs(x) (x)
# define __bpf_constant_ntohs(x) (x)
#else
# error "Fix your compiler's __BYTE_ORDER?!"
#endif

#define bpf_ntohs(x) \
  (__builtin_constant_p(x) ? \
   __bpf_constant_ntohs(x) : __bpf_ntohs(x))

// - Functions -

/*
 *  Attempts to parse the ethernet frame. This function simply walks
 *  through the nested layers of payloads (for example, ethernet -> ip
 *  -> tcp -> http) and populates the fields in _out_.
 *
 *  Returns 0 if the frame contains TCP/IP, or 1 otherwise.
 */
static __always_inline int
xdp_parse_frame(struct xdp_md *ctx)
{
  struct log_data out = {0};
  void *data_end = (void*)(long)ctx->data_end;
  void *data     = (void*)(long)ctx->data;

  out.dir = RECEIVED;
  out.ifindex = ctx->ingress_ifindex;
  struct ethhdr *eth = data;

  if ((void*)(eth+1) > data_end) // Do you have something for me?
    return 0; // Not interested to continue

  if (eth->h_proto == bpf_ntohs(ETH_P_IP))   // IPv4
  {

    out.ip_type = IPv4;
    struct iphdr *ip = (void*)(eth + 1);

    if ((void*)(ip + 1) > data_end)    
      return 0;

    if (ip->protocol != TCP_PROTO_NUMBER) // Only interested in TCP
      return 0;
    
    out.protocol = ip->protocol;
    out.ip_saddr.v4 = (__u32)(ip->saddr); // net byte order
    out.ip_daddr.v4 = (__u32)(ip->daddr);
    
    struct tcphdr *tcp = (void*)ip + ip->ihl * 4;

    if ((void*)(tcp + 1) > data_end)    
      return 0;

    int tcp_hdr_len = tcp->doff * 4;
    char* tcp_payload = (void*)tcp + tcp_hdr_len;
    int tcp_data_offset = (void*)tcp_payload - data;
    
    out.tcp_source = tcp->source;
    out.tcp_dest = tcp->dest;
    out.tcp_seq = tcp->seq;
    out.tcp_payload_size = data_end - (void*)&tcp_payload[0];

    __u32 last = 0;
    __u8 byte;
    for (int i = 0; i < MAX_TCP_IPV4_DATA_SIZE; ++i)
    {
      if ((void*)&tcp_payload[0] + i + 1 >  data_end)  // End reached
      {
        bpf_ringbuf_output(&rb, &out, sizeof(struct log_data), 0);
        break;
      }
      if (i != 0 && i % DATA_BUFF_SIZE == 0)
      {
        bpf_ringbuf_output(&rb, &out, sizeof(struct log_data), 0);
        for (int j = 0; j < DATA_BUFF_SIZE; ++j)  // Reset buffer
            out.data[j] = 0;
      }
      if (bpf_xdp_load_bytes(ctx, tcp_data_offset + i, &byte, 1) < 0)
        return 0;
      out.data[i % DATA_BUFF_SIZE] = byte;
      last = i % DATA_BUFF_SIZE;
    }
    
    return 1;
    
  } else if (eth->h_proto == bpf_ntohs(ETH_P_IPV6)) { // IPv6
    return 0; // Not supported yet
  }

  return 0; // We are not interested, thanks
}


static __always_inline void
kprobe_parse_sock(struct tcp_sock *tp, struct log_data *out)
{
  if (tp == NULL || out == NULL) return;
  
  out->dir = SENT;
  out->protocol = TCP_PROTO_NUMBER;
  
  unsigned char ipv6only =
    BPF_CORE_READ_BITFIELD_PROBED(tp, inet_conn.icsk_inet.sk.__sk_common.skc_ipv6only);
  if (ipv6only == true)  // IPv6
  {
    out->ip_type = IPv6;
    return; // Not supported yet
  }

  out->ip_type = IPv4;
  out->tcp_seq = BPF_CORE_READ(tp, snd_nxt);
  out->ifindex = BPF_CORE_READ((struct sock*)tp, sk_dst_cache, dev, ifindex);

  out->ip_saddr.v4 = BPF_CORE_READ(tp, inet_conn.icsk_inet.inet_saddr);
  out->ip_daddr.v4 = BPF_CORE_READ(tp, inet_conn.icsk_inet.sk.__sk_common.skc_daddr);

  out->tcp_source = BPF_CORE_READ(tp, inet_conn.icsk_inet.inet_sport);
  out->tcp_dest = BPF_CORE_READ(tp, inet_conn.icsk_inet.sk.__sk_common.skc_dport);
  
  return;
}


static __always_inline void
kprobe_dump_msg(struct msghdr *msg, size_t size, struct log_data *out)
{
  if (msg == NULL || size == 0 || out == NULL) return;
  
  __u8 byte;
  __kernel_size_t iov_len = BPF_CORE_READ(msg, msg_iter.__ubuf_iovec.iov_len);
  void* iov_base = BPF_CORE_READ(msg, msg_iter.__ubuf_iovec.iov_base);
  out->tcp_payload_size = iov_len;

  // TODO: use bpf_loop for higher bound
  for (int i = 0; i < MAX_TCP_IPV4_DATA_SIZE; ++i)
  {  
    if (i > size || i > iov_len)
    {
      bpf_ringbuf_output(&rb, out, sizeof(struct log_data), 0);
      break;
    }
    if (i != 0 && i % DATA_BUFF_SIZE == 0)
    {
      bpf_ringbuf_output(&rb, out, sizeof(struct log_data), 0);
      for (int j = 0; j < DATA_BUFF_SIZE; ++j)  // Reset buffer
        out->data[j] = 0;
    }
    if (bpf_probe_read_user(&byte, sizeof(byte), iov_base + i) < 0)
      return;
    out->data[i % DATA_BUFF_SIZE] = byte;
  }

  return;
}

/*
 *  - BPF_PROG_TYPE_XDP -
 *
 *  XDP programs can attach to network devices and are called for every
 *  incoming (ingress) packet received by that network device.
 */
SEC("xdp")
int tcp_dump(struct xdp_md *ctx)
{
  if (!xdp_parse_frame(ctx))
  {
    return XDP_PASS; // Move on
  }
  
  return XDP_PASS; // Options are: PASS, DROP, ABORTED, TX, REDIRECT
}


SEC("kprobe/tcp_sendmsg")
int kprobe_tcp_sendmsg(struct pt_regs* ctx)
{
  // int tcp_sendmsg(struct sock *sk, struct msghdr *msg, size_t size);
  struct tcp_sock *tp = (struct tcp_sock *)PT_REGS_PARM1(ctx);
  struct msghdr *msg = (struct msghdr *)PT_REGS_PARM2(ctx);
  size_t size = (size_t)PT_REGS_PARM3(ctx);

  struct log_data out = {0};
  kprobe_parse_sock(tp, &out);
  if (out.ip_type == IPv6) return 0;  // Not suported yet
  kprobe_dump_msg(msg, size, &out);
  
  return 0; // Return value does not mean anything
}
