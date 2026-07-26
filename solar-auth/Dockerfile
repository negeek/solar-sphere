# Build context is the repository root: solar-spectrum isn't published as a
# versioned module yet, so solar-auth/go.mod has a `replace` pointing at
# ../solar-spectrum — both need to be present, siblings, for that to
# resolve. Sibling services (solar-sentinel, solar-galaxy) aren't needed
# here at all.
#   docker build -f solar-auth/Dockerfile -t solar-auth .

FROM golang:1-alpine AS builder
WORKDIR /src

COPY solar-auth/ ./solar-auth/
COPY solar-spectrum/ ./solar-spectrum/

WORKDIR /src/solar-auth
RUN CGO_ENABLED=0 go build -o /out/server . && \
    CGO_ENABLED=0 go build -o /out/migrate ./db/v1

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
COPY --from=builder /out/migrate /migrate
EXPOSE 3000
ENTRYPOINT ["/server"]
