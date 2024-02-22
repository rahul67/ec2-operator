# Build binary
FROM --platform=${BUILDPLATFORM:-linux/adm64} golang:1.21-alpine AS builder

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Set work directory
WORKDIR /app
# Copy meta files
COPY go.mod ./
COPY go.sum ./
RUN go mod download
# Copy code
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o ec2-operator

# Runner image
FROM alpine:3.19 as runner
RUN apk add --no-cache aws-cli

WORKDIR /app
COPY --from=builder /app/ec2-operator /app/ec2-operator

CMD ["/app/ec2-operator"]