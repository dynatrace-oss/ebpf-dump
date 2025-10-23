# ebpfdump

Dump network traffic of a Kubernetes cluster.

This tool logs all HTTP requests over TCP/IPv4 in a Kubernetes
cluster. The tool is implemented as a Kubernetes operator which uses
an eBPF program to intercept and report the traffic.

The operator runs an eBPF _loader_ on each node in the cluster, which
loads/unloads the eBPF program, and reads its dumps. By default, the
data is printed to STDOUT in the operator's container (accessible via
`kubectl logs`); optionally, the operator accepts an HTTP callback
which will instruct the operator to perform a POST request to said
callback with the json-encoded network data.

This implementation only logs HTTP requests over TCP/IPv4 and ignores
all the rest; this can easily be extended to any other protocol. The
operator was scaffolded using [operator sdk](https://github.com/operator-framework/operator-sdk).

## Usage

You can configure the operator by registering a `EbpfDump` resource,
which is a custom Kubernetes resource managed by the operator. The
following is a complete example of a configuration:

```yaml
apiVersion: research.dynatrace.com/v1alpha1
kind: EbpfDump
metadata:
  labels:
    app.kubernetes.io/name: ebpfdump-operator
  name: ebpfdump-sample
  namespace: ebpfdump-operator-system
spec:
  interfaces:
    - lo
  callback: "http://example.com"
```

- `interfaces`: here you can specify a list of interfaces to attach
  to. Usually, the `lo` interface captures all the traffic within one
  Node, and the `eth0` interface intercepts all the traffic coming
  from an outside network. If none specified, the operator will attach
  to all interfaces.
- `callback`: if present and non empty, the operator will make a post
  request to this callback with the traced information encoded in json
  instead of logging to stdout. The json format is specified in the
  `NetworkDump` type in
  [api/v1alpha1/ntworkdump_type.go](./api/v1alpha1/networkdump_type.go)

To learn how you can set up and use the operator, please read the
[setup](./docs/SETUP.md), [testing](./docs/TESTING.md),
[ebpf testing](./docs/EBPF-TESTING.md) and
[deployment](./docs/DEPLOYMENT.md) documents.

To quickly deploy the operator on your cluster, run the following:

```bash
kubectl apply -f https://raw.githubusercontent.com/dynatrace-oss/ebpfdump/main/dist/install-github.yaml
```

## License and Attribution

`ebpfdump` was created by [Giovanni Santini](https://github.com/San7o).

- The controller source code is licensed under [Apache 2.0](./LICENSE.txt)
- The eBPF source code is licensed under [GPL-2.0-only](./ebpf/LICENSE.txt)

---

_**Note:** ebpfdump is not officially supported by Dynatrace._
