# Callback Service

This directory contains the callback container. This container hosts
the `/ingest` HTTP endpoint at default port `8090` and logs the body
of all the requests made to It. The `ebpfdump` operator can use this
service to aggreate all logs from all nodes in a single palce.

On kubernetes, It registers a service at port `9376`, which can be
accessed by other pods via the url
`http://callback-service.ebpfdump-operator-system.svc.cluster.local:9376/ingest`.

The commands to build, containerize and deploy the application in a
kubernetes cluster are the same as the `ebpfdump` operator expect that
there is no code generation commands. Here is a quick summary:

```bash
make build            # Build locally
make docker           # Build docker image
make deploy           # Deploy on a kubernetes cluster
make undeploy         # Undeploy the application
```

By default, the image will be pushed to
`localhost:5001/callback:latest`, you can change this by appending
`IMG=...` to the commands above.
