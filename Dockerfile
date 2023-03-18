FROM golang:1.20 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -o controller ./cmd

FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /
COPY --from=builder /workspace/controller .
USER nonroot:nonroot

ENTRYPOINT ["/controller"]