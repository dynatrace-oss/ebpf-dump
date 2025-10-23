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

#ifndef _EBPFDUMP_MAPS_H_
#define _EBPFDUMP_MAPS_H_

#include "vmlinux.h"
#include "log_data.h"
#include <bpf/bpf_helpers.h>

#define MAP_MAX_ENTRIES 1024

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 12);
  __type(value, struct log_data);
} rb SEC(".maps");

#endif // _EBPFDUMP_MAPS_H_
