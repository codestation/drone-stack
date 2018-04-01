# docker build --rm -f docker/Dockerfile -t registry.megpoid.xyz/drone-stack .

FROM docker:18.03.0-ce-dind

ADD release/linux/amd64/drone-stack /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-stack"]
