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

package controller

import (
	"context"
	"fmt"
	"net"

	ebpfdumpv1alpha1 "ebpfdump/api/v1alpha1"
	ebpf "ebpfdump/internal/controller/ebpf"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type EbpfDumpReconciler struct {
	client.Client
	UncachedClient client.Reader
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=research.dynatrace.com,resources=ebpfdumps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=research.dynatrace.com,resources=ebpfdumps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=research.dynatrace.com,resources=ebpfdumps/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get

func (r *EbpfDumpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("EbpfDump reconcile triggered.")
	var err error = nil

	ebpfDumpList := &ebpfdumpv1alpha1.EbpfDumpList{}
	err = r.Client.List(ctx, ebpfDumpList)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Failed to get EbpfDump resource: %w", err)
	}

	if len(ebpfDumpList.Items) > 1 {
		// TODO: delete the additional EbpfDump[s] or find another way
		// to make sure that there is only one resource
		return ctrl.Result{}, fmt.Errorf("There can be only one EbpfDump resource!!!")
	}

	if len(ebpfDumpList.Items) == 0 {
		if ebpf.Loaded {
			ebpf.UnloadBpf(ctx)
		}
		log.Info("No EbpfDump resource found, doing nothing.")
		return ctrl.Result{}, nil
	}

	ebpfDump := ebpfDumpList.Items[0]

	if len(ebpfDump.Spec.Interfaces) == 0 {
		// Load all interfaces
		ifaces, err := net.Interfaces()
		if err != nil {
			return ctrl.Result{}, nil
		}
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagRunning != 0 {
				ebpfDump.Spec.Interfaces = append(ebpfDump.Spec.Interfaces, iface.Name)
			}
		}
	}

	log.Info("Loading eBPF program.", "interfaces", ebpfDump.Spec.Interfaces)
	if err := ebpf.LoadBpf(ctx, ebpfDump.Spec.Interfaces); err != nil {
		return ctrl.Result{}, fmt.Errorf("Error loading eBPF program: %w", err)
	}
	go ebpf.LogData(context.Background(), r.UncachedClient)

	ebpf.Callback = ebpfDump.Spec.Callback

	return ctrl.Result{}, nil
}

func (r *EbpfDumpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ebpfdumpv1alpha1.EbpfDump{}).
		Complete(r)
}
