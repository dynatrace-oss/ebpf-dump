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
	"flag"
	"io"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	DefaultPort = "8090"
)

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	log := log.FromContext(r.Context())

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "error reading request's body")
	}
	log.Info("received request", "body", string(bytes))
}

func main() {

	var portFlag = flag.String("port", DefaultPort, "port to listen to")

	opts := zap.Options{
		Development: true, // Enables console encoder and disables sampling
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	log := log.FromContext(context.Background())
	http.HandleFunc("/ingest", ingestHandler)

	log.Info("Server listening", "port", *portFlag)

	if err := http.ListenAndServe(":"+*portFlag, nil); err != nil {
		log.Error(err, "server error")
	}

	return
}
