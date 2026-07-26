package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"

	v1 "github.com/negeek/solar-sphere/solar-sentinel/api/v1"
	authmw "github.com/negeek/solar-sphere/solar-sentinel/middlewares/v1"
	repo "github.com/negeek/solar-sphere/solar-sentinel/repository/v1"
	service "github.com/negeek/solar-sphere/solar-sentinel/service/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/env"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
	"github.com/negeek/solar-sphere/solar-spectrum/logging"
	"github.com/negeek/solar-sphere/solar-spectrum/mongoutil"
)

func main() {
	// .env is optional — present for local dev, absent in Docker (where
	// Compose injects env vars directly). Only a malformed file is an error.
	if err := env.Load(".env"); err != nil && !os.IsNotExist(err) {
		panic("solar-sentinel: loading .env: " + err.Error())
	}

	log := logging.New("solar-sentinel")

	ctx := context.Background()
	client, db, err := mongoutil.Connect(ctx, os.Getenv("DATABASE_URL"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Error("connect to mongo", "error", err)
		os.Exit(1)
	}
	log.Info("connected to mongo")

	repository := repo.NewRepository(db)
	deviceService := service.NewDeviceService(repository)
	irradianceService := service.NewIrradianceService(repository)
	handler := v1.NewHandler(deviceService, irradianceService)
	authMiddleware := authmw.Authentication(repository, os.Getenv("VERIFICATION_KEY"))

	router := mux.NewRouter()
	router.Use(httpapi.CORS)
	v1.Routes(router.StrictSlash(true), handler, authMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// A malformed or unexpected MQTT payload must never take the whole
	// service down — log and move on to the next message.
	messageHandler := func(_ mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()

		parts := strings.Split(topic, "/")
		if len(parts) < 4 {
			log.Error("mqtt message on malformed topic", "topic", topic)
			return
		}
		deviceID := parts[3]

		var data map[string]interface{}
		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			log.Error("mqtt message is not valid json", "topic", topic, "error", err)
			return
		}

		if err := irradianceService.Save(context.Background(), deviceID, data); err != nil {
			log.Error("saving irradiance reading", "device_id", deviceID, "error", err)
		}
	}

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(os.Getenv("BROKER_URL"))
	mqttOpts.SetClientID(os.Getenv("MQTT_CLIENT_ID"))
	mqttOpts.SetUsername(os.Getenv("MQTT_USERNAME"))
	mqttOpts.SetPassword(os.Getenv("MQTT_PASSWORD"))
	mqttOpts.SetDefaultPublishHandler(messageHandler)
	mqttOpts.OnConnect = func(mqtt.Client) { log.Info("mqtt connected") }
	mqttOpts.OnConnectionLost = func(_ mqtt.Client, err error) { log.Error("mqtt connection lost", "error", err) }

	mqttClient := mqtt.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error("mqtt connect", "error", token.Error())
	}

	mqttTopic := os.Getenv("MQTT_TOPIC")
	if token := mqttClient.Subscribe(mqttTopic, 0, messageHandler); token.Wait() && token.Error() != nil {
		log.Error("mqtt subscribe", "topic", mqttTopic, "error", token.Error())
	} else {
		log.Info("subscribed to mqtt topic", "topic", mqttTopic)
	}

	go func() {
		log.Info("starting server", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mqttClient.Disconnect(250)
	if err := mongoutil.Disconnect(shutdownCtx, client, 15*time.Second); err != nil {
		log.Error("disconnect mongo", "error", err)
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown", "error", err)
	}

	log.Info("shut down")
}
