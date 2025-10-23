# Testing

This document covers useful information if you want to test the
operator or kubernetes. If you are interested to test just the eBPF
program instead, refer the the documentation in [ebpf testing](./EBPF-TESTING.md).

Before running tests, make sure you have followed the
[setup](./SETUP.md) and you have the operator running on your
cluster. In this document we will see a few examples on how to
interact with the operator and what to expect.

Note that you may need high privileges to run the commands from this
document depending on your permissions.

## Trace some packets

Let's deploy the following `EbpfDump` resource to Kubernetes, which
can be found in `config/samples/ebpfdump_v1alpha1.yaml`:

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
``

```bash
kubectl apply -f config/samples/ebpfdump_v1alpha1.yaml 
```

This resource tells the operator that we want to listen on the loopback
interface ("lo"). All the traffic internal to the node will pass
through the `lo` interface, and all the traffic coming from an outside
network will pass through the `eth0` interface.

You can only have one `EpfDump` resource in your cluster, and the
operator will respect only this one. You will see the operator
complaining otherwise.

If you read the logs of an `ebpfdump` operator, you should see a lot
of traffic:

```bash
kubectl get pods -n ebpfdump-operator-system
kubectl logs ebpfdump-operator-manager-wtmsf -n ebpfdump-operator-system
```

```text
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870502,\"remote_ip\":\"127.0.0.1\",\"remote_port\":35626,\"method\":\"GET\",\"status_code\":0,\"path\":\"/readyz\"...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870503,\"remote_ip\":\"127.0.0.1\",\"remote_port\":35626,\"method\":\"\",\"status_code\":200,\"path\":\"\",\"version\":\...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870747,\"remote_ip\":\"10.244.0.1\",\"remote_port\":60096,\"method\":\"GET\",\"status_code\":0,\"path\":\"/health\",...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870747,\"remote_ip\":\"10.244.0.1\",\"remote_port\":60096,\"method\":\"\",\"status_code\":200,\"path\":\"\",\"version\":...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870750,\"remote_ip\":\"10.244.0.1\",\"remote_port\":45266,\"method\":\"GET\",\"status_code\":0,\"path\":\"/health\",...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870750,\"remote_ip\":\"10.244.0.1\",\"remote_port\":45266,\"method\":\"\",\"status_code\":200,\"path\":\"\",\"version\":...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870817,\"remote_ip\":\"172.19.0.3\",\"remote_port\":57488,\"method\":\"GET\",\"status_code\":0,\"path\":\"/readyz\"...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870817,\"remote_ip\":\"172.19.0.3\",\"remote_port\":57488,\"method\":\"GET\",\"status_code\":0,\"path\":\"/readyz\"...
2025-07-23T13:21:10Z    INFO    Traffic received    {"data": "{\"timing_ms\":1753276870817,\"remote_ip\":\"172.19.0.3\",\"remote_port\":57488,\"method\":\"\",\"status_code\":200,\"path\":\"\",\"version\":...
```

You will see a lot of GET requests to `/healthz` and `/readyz` and
their responsens, indeed this is the Kubernetes control pane checking
out on the node's well being.

To further test the operator, let's try to generate some traffic from
outside the pods's network. I will first change the interface to
`eth0` to listen for non-local traffic, then I will create an nginx
pod in the default namespace, and an nginx `NodePort` service so that
we can connect to the pod via `<node-ip>:30007`. Finally, I will make
a GET request to the pod:

```bash
kubectl delete -f config/samples/ebpfdump_v1alpha1.yaml
kubectl apply -f config/samples/ebpfdump_v1alpha1_eth0.yaml
kubectl apply -f hack/k8s-manifests/sample-nginx-pod.yaml
kubectl apply -f hack/k8s-manifests/sample-nginx-service.yaml

kubectl get pods -o=wide  # Check in which node does the nginx-pod live
kubectl get nodes -o=wide # Get the IP of the node where the nginx-pod lives

curl -4 http://<node-ip>:30007/
```

```html
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

And indeed, the `ebpfdump` operator that lives in the same node as my
`nginx-pod` has logged, among other things:

```bash
kubectl logs ebpfdump-operator-manager-wtmsf -n ebpfdump-operator-system
```

```json
2025-07-23T17:26:17Z    INFO    Traffic received    {"data": "{\"direction\":1,\"timing_ms\":1753291577665,\"remote_ip\":\"172.19.0.1\",\"remote_port\":58136,\"method\":\"GET\",\"status_code\":0,\"path\":\"/\",\"version\":\"1.1\",\"headers\":{\"Accept\":[\"*/*\"],\"User-Agent\":[\"curl/8.5.0\"]},\"body\":\"...
```

We know that this is the right one because the `User-Agent` is curl.

Note that there is one `ebpfdump` operator for each node, so each
operator will log only the traffic happening on that node. This is not
really handy to check the logs for each `ebpfdump` pod, we will now
see how you can easily gather all the logs in a single place using
callbacks.

## Callback

You can ask the operator to use a callback by setting the `callback`
filed in the resource:

```yaml
apiVersion: research.dynatrace.com/v1alpha1
kind: EbpfDump
metadata:
  labels:
    app.kubernetes.io/name: ebpfdump-operator
  name: ebpfdump-sample-callback
  namespace: ebpfdump-operator-system
spec:
  interfaces:
   - lo
  callback: "http://callback-service.ebpfdump-operator-system.svc.cluster.local:9376/ingest"

```

Remember that there can only be one `EbpfDump` resource so if you
already created one, you should delete or update It. This project
provides a callback service in `callback/`, you can read more
information in its [README](../callback/README.md) file.

Let's setup the callback server:

```bash
cd callback/
make docker
make deploy
```

Let's now instruct the operator to call this server:

```bash
kubectl apply -f config/samples/ebpfdump_v1alpha1_callback.yaml
```

And we should see all the data arriving centrally to the callback service:

```text
2025-07-23T15:18:49Z    INFO    received request    {"body": "{\"timing_ms\":1753283929360,\"remote_ip\":\"127.0.0.1\",\"remote_port\":34760,\"metho...
2025-07-23T15:18:49Z    INFO    received request    {"body": "{\"timing_ms\":1753283929360,\"remote_ip\":\"127.0.0.1\",\"remote_port\":34760,\"met...
2025-07-23T15:18:49Z    INFO    received request    {"body": "{\"timing_ms\":1753283929360,\"remote_ip\":\"127.0.0.1\",\"remote_port\":34760,\"met...
2025-07-23T15:18:49Z    INFO    received request    {"body": "{\"timing_ms\":1753283929360,\"remote_ip\":\"127.0.0.1\",\"remote_port\":34760,\"met...
...
```

The operator will stop logging in Its pod's output if a callback is
specified.
