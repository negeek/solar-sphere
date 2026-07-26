module github.com/negeek/solar-sphere/solar-auth

go 1.26

// solar-spectrum isn't published as a versioned module yet, so this points
// builds at the local copy — needed for Docker builds, which don't bring
// go.work along (see solar-auth/Dockerfile). Local dev via `go build` from
// the repo root still goes through go.work instead, which takes priority
// over this when present.
replace github.com/negeek/solar-sphere/solar-spectrum => ../solar-spectrum

require (
	github.com/gorilla/mux v1.8.1
	github.com/negeek/solar-sphere/solar-spectrum v0.0.0-20240411201025-962f0f2288c9
	go.mongodb.org/mongo-driver v1.17.9
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/montanaflynn/stats v0.12.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)
