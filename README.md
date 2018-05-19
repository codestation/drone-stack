# drone-stack

Drone plugin to deploy stacks in Docker Swarm

## Basic Usage with Drone CI

```yml
pipeline:
  deploy:
    image: codestation/drone-stack
    host: tcp://example.com:2376
    stack_name: mystack
    tlsverify: true
    secrets: [ docker_cert, docker_key, docker_cacert ]
```

### Use a private registry to pull the image from

```yml
pipeline:
  deploy:
    image: codestation/drone-stack
    host: tcp://example.com:2376
    stack_name: mystack
    tlsverify: true
+   registry: registry.example.com
-   secrets: [ docker_cert, docker_key, docker_cacert ]
+   secrets: [ docker_username, docker_password, docker_cert, docker_key, docker_cacert ]
```

The `tls`, `tlsverify`, `docker_cert`, `docker_key` and `docker_cacert` combinations are the same of the client modes supported on the docker binary. Check [here](https://docs.docker.com/engine/security/https/#client-modes) for more details.

## Basic Usage using a Docker Container

Execute from the working directory:

```bash
docker run --rm \
  -e PLUGIN_HOST=tcp://example.com:2376 \
  -e PLUGIN_TLSVERIFY=true \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  codestation/drone-stack
```

## Load certificates as drone secrets

Using the drone cli, go to the directory where your docker certificates are located then run the following commands:

```bash
drone secret create myuser/myapp --name docker_cert --value @cert.pem
drone secret create myuser/myapp --name docker_key --value @key.pem
drone secret create myuser/myapp --name docker_cacert --value @ca.pem
```

## Secret Reference

* `docker_username` - authenticates with this username
* `docker_password` - authenticates with this password
* `docker_cert` - client certificate
* `docker_key` - client key
* `docker_cacert` - CA certificate

## Parameter Reference

* `compose` - compose file to be used, defaults to docker-compose.yml
* `host` - remote docker swarm host:port
* `prune` - prune services that are no longer referenced
* `stack_name` - name of the stack to deploy
* `tls` - use TLS. Implied by `tlsverify`
* `tlsverify` - use TLS and verify the remote host
* `registry` - authenticates to this registry
