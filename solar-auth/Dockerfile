# Build context is the repository root: all four modules are needed to
# resolve github.com/negeek/solar-sphere/solar-spectrum via go.work, since
# it isn't (yet) published as a versioned module.
#   docker build -f solar-auth/Dockerfile -t solar-auth .

FROM golang:1-alpine AS builder
WORKDIR /src

COPY go.work go.work.sum ./
COPY solar-auth/ ./solar-auth/
COPY solar-sentinel/ ./solar-sentinel/
COPY solar-galaxy/ ./solar-galaxy/
COPY solar-spectrum/ ./solar-spectrum/

WORKDIR /src/solar-auth
RUN CGO_ENABLED=0 go build -o /out/server . && \
    CGO_ENABLED=0 go build -o /out/migrate ./db/v1

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
COPY --from=builder /out/migrate /migrate
EXPOSE 3000
ENTRYPOINT ["/server"]
