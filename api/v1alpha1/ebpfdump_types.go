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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EbpfDumpSpec defines the desired state of EbpfDump
type EbpfDumpSpec struct {
	// The name of the interface to dump
	Interfaces []string `json:"interfaces,omitempty"`
	// The callback to call with the dumped data
	Callback string `json:"callback,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type EbpfDump struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EbpfDumpSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type EbpfDumpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EbpfDump `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EbpfDump{}, &EbpfDumpList{})
}
