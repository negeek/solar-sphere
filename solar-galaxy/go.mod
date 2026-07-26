module github.com/negeek/solar-sphere/solar-galaxy

go 1.26

// See solar-auth/go.mod for why this replace exists.
replace github.com/negeek/solar-sphere/solar-spectrum => ../solar-spectrum

require (
	github.com/gorilla/mux v1.8.1
	github.com/negeek/solar-sphere/solar-spectrum v0.0.0-00010101000000-000000000000
)
