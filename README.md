# drone-stack

Drone plugin to deploy stacks in Docker Swarm

## Build

Build the binary with the following commands:

```
sh .drone.sh
```

## Docker

Build the Docker image with the following commands:

```
docker build --rm=true -f docker/Dockerfile -t codestation/drone-stack .
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_NAME=hello-world \
  -e PLUGIN_HOST=tcp://example.com:2376 \
  -e PLUGIN_TLSVERIFY=true \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  codestation/drone-stack
```
