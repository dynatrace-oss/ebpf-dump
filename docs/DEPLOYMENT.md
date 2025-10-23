# Deployment

In this document you will find instructions on how to deploy the
operator. For convenience, this project uses different environments to
manage building and deploying. By default, there are two different
environments: Local and GitHub. You can easily add a custom
environment by creating a file called `.env-<ENV-NAME>` where
`<ENV-NAME>` is a name of your choice. This file will be included in
the Makefile before running any command, so you can change the
variables used by the Makefile from the env file without changing the
Makefile.

For example, the `IMG` variable tells the Makefile where to push
images and tells the operator where to pull them. You can add an entry
to your custom environment `.env-custom` like so:

```text
IMG=registry/my-beautiful-name:latest
```

To select which environment to use, append `ENV=<ENV-NAME>` after your
make commands, for example:

```bash
make deploy ENV=github
```

The above commands deploys the operator using the `github` environment
defined in `.env-github`. The default environment is `local`, in this
case you can omit the `ENV` from the Makefile command to use the local
environment.

## Docker images

To build the docker image, you can use the command `make docker-build`.
To push the image you can use `make docker-push`. To do both, you can
simply use `make docker`.

For example, to build and push the images to GitHub:

```bash
make docker ENV=github
```

Note that to build the image you need to have all the necessary
dependencies, please refer to the [setup](./SETUP.md) document for
instructions.

All the commands in this document can be easily found via `make help`.

## Deploy

To deploy the operator images from a registry:

```bash
make deploy ENV=local
```

You can also build an installer with the following command:

```bash
make build-installer ENV=local
```

The file `dist/install-local.yaml` will be created, which can be used
to deploy the operator with:

```bash
kubectl apply -f dist/install-local.yaml
```
