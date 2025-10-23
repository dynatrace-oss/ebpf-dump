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
	"fmt"
)

func uint32ToDotNotation(num uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", num&0xff, (num&0xff00)>>8, (num&0xff0000)>>16, (num&0xff00000)>>24)
}

// Network byte order to host byte order
func ntoh(num uint16) uint16 {
	return ((num & 0xFF) << 8) | (num >> 8)
}

func int8ToByte(in [256]int8) []byte {
	out := make([]byte, len(in))
	for i, d := range in {
		out[i] = byte(d)
	}
	return out
}
