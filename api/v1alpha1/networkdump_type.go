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

package v1alpha1

type NetworkDump struct {
	Direction  uint32              `json:"direction"`
	TimingMs   int64               `json:"timing_ms"`
	RemoteIp   string              `json:"remote_ip"`
	RemotePort uint16              `json:"remote_port"`
	Method     string              `json:"method"`
	StatusCode int                 `json:"status_code"`
	Path       string              `json:"path"`
	Version    string              `json:"version"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}
