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

#ifndef _EBPFDUMP_DATA_H_
#define _EBPFDUMP_DATA_H_

#include "vmlinux.h"

#define DATA_BUFF_SIZE (1<<8) // This is the maximum possible size...

enum ip_type_t {
  IPv4 = 0,
  IPv6,
};

enum direction {
  SENT = 0,
  RECEIVED = 0,
};

struct log_data {
  enum direction dir;
  int ifindex;
  enum ip_type_t ip_type;
  union {
    __u32 v4;
    struct in6_addr v6;
  } ip_saddr;
  union {
    __u32 v4;
    struct in6_addr v6;
  } ip_daddr;
  __u8 protocol;
  __be16 tcp_source;
  __be16 tcp_dest;
  __be32 tcp_seq;
  size_t tcp_payload_size;
  char data[DATA_BUFF_SIZE];
};

#endif // _EBPFDUMP_DATA_H_
