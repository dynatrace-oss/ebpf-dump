# Setup

This document contains useful information for building and running the
operator in a local cluster. Additional documentation on how to deploy
the operator from the GitHub container registry (GHCR) can be found in
[deployment](./DEPLOYMENT.md).

## Setup a local cluster

This repository provides the script `hack/registry-cluster.sh` which
will create a local cluster using
[kind](https://github.com/kubernetes-sigs/kind) with one control node
and one worker node. Additionally, It sets up a local docker registry
to push the operator's image during development and It provides
useful Make commands to interact with the local cluster.

Run the following command to create the cluster (this needs to be run
only once):

```bash
make create-cluster-local
```

To remove it, you can run `make delete-cluster-local`.

## Generate files

The operator uses generators to create various config files such as
RBAC policies, CRD manifests and the eBPF program. Those need to be
regenerated each time they are updated.

To generate the RBAC policies, run:

```bash
make generate
```

To create the CRD manifests, run:

```bash
make manifests
```

To generate the eBPF program, first you need to have the following
dependencies in your system:

- Linux kernel version 5.7 or later, with ebpf support enabled
- LLVM 11 or later (clang and llvm-strip)
- libbpf headers
- Linux kernel headers
- a recent go compiler

On Ubuntu, these are the packages you need:

```bash
apt-get install make clang llvm libbpf-dev golang linux-headers-$(uname -r)
```

Once you have the dependencies, run:

```bash
make generate-ebpf
```

## Build the docker container

Note that you may need sudo privileges for the following commands
based on your system.

The operator can be compiled with `make build`, while this is useful
to check if the code compiles or not, you cannot directly test the
operator in Kubernetes since it is required to create a container. You
can build a local docker image with:

```bash
make docker-build
```

You should now push the image to a docker repository. If you generated
the cluster with `registry-cluster.sh` you already have a local
registry available, and you can push the container with:

```bash
make docker-push
```

To do both of the above in a single command, for convenience, run:

```bash
make docker
```

## Deploy the operator

Finally, you can deploy the operator to the test cluster with:

```bash
make deploy
```

To undeploy:

```bash
make undeploy
```

Usually, if you made some changes to the application and you want to
deploy it again, you usually need to kill the existing operator so
that the images can be updated to the new version. You can easily do
this with:

```bash
make kill-pods
```

You can now proceed to do some [testing](./TESTING.md).
