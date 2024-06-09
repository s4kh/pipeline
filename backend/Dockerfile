# syntax=docker/dockerfile:1

FROM golang:1.22 as builder

ARG CGO_ENABLED=0
# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY . .

RUN go build -o /backend ./cmd/server/.

FROM scratch
COPY --from=builder /backend /backend
COPY .env .

ENTRYPOINT ["/backend"]
