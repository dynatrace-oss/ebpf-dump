# Ebpf Testing

In this document we discuss how to test the eBPF program. We assume
that you have the required dependencies, if not please check the
[setup](./SETUP.md) document so that you are able to compile the
program.

The code of the eBPF program lives in the `ebpf/` directory of the
project. A go loader is implemented in `test/ebpf-local/` which will
load the ebpf program on your machine. To help you during development,
the Makefile at the root directory comes with the following helpful
make commands:

- `test-build-ebpf`: build the ebpf-local program
- `test-run-ebpf`: run the ebpf-local program
- `test-ebpf`: build and run the ebpf-local program

You often need some kind of server to send network traffic to. You can
use the `callback/` server which will echo the requests he receives
and is accessible via `http://localhost:8090/callback`.

To run both the eBPF program and the callback service, you can create
a tmux session with:

```bash
make test-tmux
```

This will create a session with three panes split horizontally: one
with the eBPF program, one with the callback server, and one with a
bash shell.

## Example

Run the eBPF program:

```bash
make test-run-ebpf INTERFACE=eth0
```

Generate some traffic:

```bash
curl -4 http://kernel.org
```

Output, with both the request (User-Agent curl) and response (unicode,
encoding html):

```text
2025-07-23T19:47:02+02:00       INFO    Traffic received        {"data": "{\"direction\":0,\"timing_ms\":1753292822887,\"remote_ip\":\"10.106.3.0\",\"remote_port\":50684,\"method\":\"GET\",\"status_code\":0,\"path\":\"/\",\"version\":\"1.1\",\"headers\":{\"Accept\":[\"*/*\"],\"User-Agent\":[\"curl/8.5.0\"]},\"body\":\"\"}"}
2025-07-23T19:47:03+02:00       INFO    Traffic received        {"data": "{\"direction\":0,\"timing_ms\":1753292823016,\"remote_ip\":\"10.106.3.0\",\"remote_port\":50684,\"method\":\"\",\"status_code\":301,\"path\":\"\",\"version\":\"1.1\",\"headers\":{\"Connection\":[\"keep-alive\"],\"Content-Length\":[\"162\"],\"Content-Type\":[\"text/html\"],\"Date\":[\"Wed, 23 Jul 2025 17:47:02 GMT\"],\"Location\":[\"https://kernel.org/\"],\"Server\":[\"nginx\"]},\"body\":\"\\u003chtml\\u003e\\r\\n\\u003chead\\u003e\\u003ctitle\\u003e301 Moved Permanently\\u003c/title\\u003e\\u003c/head\\u003e\\r\\n\\u003cbody\\u003e\\r\\n\\u003ccenter\\u003e\\u003ch1\\u003e301 Moved Permanently\\u003c/h1\\u003e\\u003c/center\\u003e\\r\\n\\u003chr\\u003e\\u003ccenter\\u003enginx\\u003c/center\\u003e\\r\\n\\u003c/body\\u003e\\r\\n\\u003c/html\\u003e\\r\\n\"}"}
```
