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

package main

import (
	"context"
	"os"

	controller "ebpfdump/internal/controller"
	ebpf "ebpfdump/internal/controller/ebpf"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	InterfacesNames = []string{"lo"}
	Port            = "8090"
)

func main() {

	if len(os.Args) > 1 {
		InterfacesNames = []string{}
		for i, arg := range os.Args {
			if i == 0 {
				continue
			}
			InterfacesNames = append(InterfacesNames, arg)
		}
	}

	// Log format
	opts := zap.Options{
		Development: true, // Enables console encoder and disables sampling
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := context.Background()
	log := log.FromContext(ctx)

	reconciler := controller.EbpfDumpReconciler{}

	if err := ebpf.LoadBpf(ctx, InterfacesNames); err != nil {
		log.Error(err, "Error loading eBPF program")
		return
	}
	defer ebpf.UnloadBpf(ctx)

	log.Info("Logging data...")
	ebpf.LogData(ctx, reconciler.UncachedClient)

	return
}
