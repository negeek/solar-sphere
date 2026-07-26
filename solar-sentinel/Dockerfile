# Build context is the repository root — see solar-auth/Dockerfile for why.
#   docker build -f solar-sentinel/Dockerfile -t solar-sentinel .

FROM golang:1-alpine AS builder
WORKDIR /src

COPY solar-sentinel/ ./solar-sentinel/
COPY solar-spectrum/ ./solar-spectrum/

WORKDIR /src/solar-sentinel
RUN CGO_ENABLED=0 go build -o /out/server . && \
    CGO_ENABLED=0 go build -o /out/migrate ./db/v1

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
COPY --from=builder /out/migrate /migrate
EXPOSE 5000
ENTRYPOINT ["/server"]
