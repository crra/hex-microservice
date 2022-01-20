ARG BUILDER_IMAGE=golang:1.18-rc-alpine
ARG DISTROLESS_IMAGE=gcr.io/distroless/static
############################
# STEP 1 build executable binary
############################
FROM ${BUILDER_IMAGE} as builder

# NOTE: if docker images are built inside wsl2 behind a corporate proxy:
# 1. copy you MITM proxy into the "certificates" folder
# 2. start cntlm (or any other NTLM proxy)
# 3. provide the proxy via `host.docker.internal`
#    --add-host=host.docker.internal:host-gateway 
#    --build-arg HTTPS_PROXY=http://host.docker.internal:3128
#    --build-arg HTTP_PROXY=http://host.docker.internal:3128 .
COPY certificates/*.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /src/app

# use modules
COPY go.mod .

RUN go mod download
RUN go mod verify

ARG VERSION=develop
ARG NAME=shortener
ARG MAIN=./cmd/service/

ENV VERSION=${VERSION}
ENV NAME=${NAME}
ENV MAIN=${MAIN}

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
      -ldflags="-w -s \
      -X main.name="${NAME}" \
      -X main.version="${VERSION}" \
      -extldflags '-static'" -a \
      -buildvcs=false \
      -buildinfo=false \
      -o /go/bin/app ${MAIN}

############################
# STEP 2 build a small image
############################
# using static nonroot image
# user:group is nobody:nobody, uid:gid = 65534:65534
FROM ${DISTROLESS_IMAGE}

WORKDIR /results

COPY --from=builder /go/bin/app /

EXPOSE 8000
ENTRYPOINT ["/app"]