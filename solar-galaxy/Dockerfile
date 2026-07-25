# Build context is the repository root — see solar-auth/Dockerfile for why.
#   docker build -f solar-galaxy/Dockerfile -t solar-galaxy .

FROM golang:1-alpine AS builder
WORKDIR /src

COPY go.work go.work.sum ./
COPY solar-auth/ ./solar-auth/
COPY solar-sentinel/ ./solar-sentinel/
COPY solar-galaxy/ ./solar-galaxy/
COPY solar-spectrum/ ./solar-spectrum/

WORKDIR /src/solar-galaxy
RUN CGO_ENABLED=0 go build -o /out/server .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
