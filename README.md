# drone-stack

Drone plugin to deploy stacks in Docker Swarm

## Basic Usage with Drone CI

You can use this plugin to connect to Docker via SSH, TCP or a Unix Socket.

### Connect via SSH

```yml
kind: pipeline
name: default
type: docker

steps:
  - name: deploy
    image: codestation/drone-stack
    settings:
      host: ssh://user@example.com
      stack_name: mystack
      ssh_key:
        from_secret: ssh_key
```

### Connect via TLS

```yml
kind: pipeline
name: default
type: docker

steps:
  - name: deploy
    image: codestation/drone-stack
    settings:
      host: tcp://example.com:2376
      stack_name: mystack
      tlsverify: true
      docker_cert:
        from_secret: docker_cert
      docker_key:
        from_secret: docker_key
      docker_cacert:
        from_secret: docker_cacert
```

#### Use a private registry to pull the image from

```
kind: pipeline
name: default
type: docker

steps:
  - name: deploy
    image: codestation/drone-stack
    settings:
      host: tcp://example.com:2376
      stack_name: mystack
      tlsverify: true
+     registry: registry.example.com
+     docker_username:
+       from_secret: docker_username
+     docker_password:
+       from_secret: docker_password
      docker_cert:
        from_secret: docker_cert
      docker_key:
        from_secret: docker_key
      docker_cacert:
        from_secret: docker_cacert
```

The `tls`, `tlsverify`, `docker_cert`, `docker_key` and `docker_cacert` combinations are the same of the client modes supported on the docker binary. Check [here](https://docs.docker.com/engine/security/https/#client-modes) for more details.

## Load certificates as drone secrets

Using the drone cli, go to the directory where your docker certificates or ssh keys are located then run the following commands:

### For a single repo

```bash
drone secret add myuser/myapp --name docker_cert --data @cert.pem
drone secret add myuser/myapp --name docker_key --data @key.pem
drone secret add myuser/myapp --name docker_cacert --data @ca.pem
# for for ssh
drone secret add myuser/myapp --name ssh_key --data @id_rsa
```

### For a whole organization

```bash
drone orgsecret add myuser docker_cert @cert.pem
drone orgsecret add myuser docker_key @key.pem
drone orgsecret add myuser docker_cacert @ca.pem
# or for ssh
drone orgsecret add myuser ssh_key @id_rsa
```

You can use the files or encode the secrets using base64.

## Secret Reference

* `docker_username` - authenticates with this username
* `docker_password` - authenticates with this password
* `docker_cert` - client certificate
* `docker_key` - client key
* `docker_cacert` - CA certificate
* `ssh_key` - SSH private key

## Parameter Reference

* `compose` - compose file to be used, defaults to docker-compose.yml
* `host` - remote docker swarm host:port, can use `SSH` or `TLS`
* `prune` - prune services that are no longer referenced
* `stack_name` - name of the stack to deploy
* `tls` - use TLS. Implied by `tlsverify`
* `tlsverify` - use TLS and verify the remote host
* `registry` - authenticates to this registry
* `ssh_key` - SSH private key
