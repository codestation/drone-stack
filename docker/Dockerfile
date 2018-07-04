FROM golang:1.10 AS builder

ARG BUILD_NUMBER=0
ARG COMMIT_SHA
ARG SOURCE_COMMIT
ENV BUILD_COMMIT=${COMMIT_SHA:-${SOURCE_COMMIT:-unknown}}

COPY . $GOPATH/src/megpoid.xyz/go/drone-stack/
WORKDIR $GOPATH/src/megpoid.xyz/go/drone-stack/

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN CGO_ENABLED=0 go install -ldflags "-w -s -X main.build=${BUILD_NUMBER} -X main.commit=$(expr substr BUILD_COMMIT_SHORT 1 8)" -a -tags netgo ./...

FROM docker:18.03.0-ce-dind
LABEL maintainer="codestation <codestation404@gmail.com>"

COPY --from=builder /go/bin/drone-stack /bin/drone-stack

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-stack"]
