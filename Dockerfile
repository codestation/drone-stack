FROM golang:1.18 AS builder

ARG CI_COMMIT_TAG
ARG CI_COMMIT_BRANCH
ARG CI_COMMIT_SHA
ARG CI_PIPELINE_CREATED_AT
ARG GOPROXY
ENV GOPROXY=${GOPROXY}

WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src/

RUN set -ex; \
    CGO_ENABLED=0 go build -o release/drone-stack \
    -ldflags "-w -s \
    -X main.Version=${CI_COMMIT_TAG:-$CI_COMMIT_BRANCH} \
    -X main.Commit=$(echo "$CI_COMMIT_SHA" | cut -c1-8) \
    -X main.BuildTime=${CI_PIPELINE_CREATED_AT}" \
    ./cmd/drone-stack

FROM docker:20.10-dind
LABEL maintainer="codestation <codestation404@gmail.com>"

COPY --from=builder /go/bin/drone-stack /bin/drone-stack

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-stack"]
