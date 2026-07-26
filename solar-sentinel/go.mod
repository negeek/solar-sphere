module github.com/negeek/solar-sphere/solar-sentinel

go 1.26

// See solar-auth/go.mod for why this replace exists.
replace github.com/negeek/solar-sphere/solar-spectrum => ../solar-spectrum

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/gorilla/mux v1.8.1
	github.com/negeek/solar-sphere/solar-spectrum v0.0.0-20240330143710-f79ba390bf71
	go.mongodb.org/mongo-driver v1.17.9
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)
