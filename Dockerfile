############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache 'git=~2'

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Fetch dependencies.
# Using go get.
COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/main .

############################
# STEP 2 build a small image
############################
FROM gcr.io/distroless/static

WORKDIR /

# Copy our static executable.
COPY --from=builder /go/main /go/main
COPY views /go/views
COPY static /go/static

ENV PORT 8080
ENV GIN_MODE release
EXPOSE 8080

WORKDIR /go

# Run the Go Gin binary.
ENTRYPOINT ["/go/main"]