# Build the manager binary
FROM golang:1.13-alpine as builder

WORKDIR /iris
ENV GOPROXY="http://172.26.1.9:5000"
ENV GOSUMDB="off"
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY pkg pkg
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN  go mod download

RUN go build  -a -o  mybinary  cmd/mybinary/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.12
WORKDIR /
COPY --from=builder /iris/mybinary .
EXPOSE 8080
ENTRYPOINT ["/mybinary"]



