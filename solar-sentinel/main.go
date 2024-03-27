package main

import (
	"os"
	"log"
	"time"
	"context"
	"syscall"
	"strings"
	"encoding/json"
	"net/http"
    "os/signal"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
    mqtt "github.com/eclipse/paho.mqtt.golang"
    irr"github.com/negeek/solar-sphere/solar-sentinel/api/v1"
	v1routes "github.com/negeek/solar-sphere/solar-sentinel/api/v1"
	"github.com/negeek/solar-sphere/solar-sentinel/db"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var payloadMap map[string]interface{}
	topic:=msg.Topic()
	payload:=msg.Payload()
	log.Printf("Received message: %s from topic: %s\n", payload, topic)
    device_id:= strings.Split(topic, "/")[3]
	err := json.Unmarshal(payload, &payloadMap)
	if err != nil {
		log.Fatal("Invalid msg format")
	}
    err = irr.SaveSolarIrrdianceData(device_id, payloadMap)
	if err != nil {
		log.Fatal("Unable to save data")
	}

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    log.Println("Mqtt Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    log.Printf("Mqtt Connection lost: %v", err)
}

func main() {
	appEnv:=os.Getenv("APP_ENV")
	if appEnv=="dev"{
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

    //custom servermutiplexer
	router := mux.NewRouter()
	v1routes.Routes(router.StrictSlash(true))

    //custom server
	server:=&http.Server{
		Addr: ":5000",
		Handler: router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 *  time.Second,
	}

    // DB connection
	dbUrl:= os.Getenv("DATABASE_URL")
	dbName:=os.Getenv("DB_NAME")
	log.Println("Connecting to db")
	dbctx, dbcancel, err:= db.Connect(dbUrl,dbName)
	if err != nil {
		log.Fatal(err)
	}
	
    // Connect to broker and suscribe to topic
    opts := mqtt.NewClientOptions()
    opts.AddBroker(os.Getenv("BROKER_URL"))
    opts.SetClientID(os.Getenv("MQTT_CLIENT_ID"))
    opts.SetUsername(os.Getenv("MQTT_USERNAME"))
    opts.SetPassword(os.Getenv("MQTT_PASSWORD"))
    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Println(token.Error())
    }
    suscribeToTopic(client, os.Getenv("MQTT_TOPIC"), 0, nil)


    go func() {
		log.Println("Start server")
		if err:= server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

    c := make(chan os.Signal, 1)
	// accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL will not be caught.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// disconnect broker and db
	db.Disconnect(dbctx,dbcancel)
    log.Println("Disconnect from broker")

    client.Disconnect(250)
    log.Println("Shutting down")

    server.Shutdown(ctx)
	os.Exit(0)
}

func suscribeToTopic(client mqtt.Client, topic string, qos byte, msgH mqtt.MessageHandler) {
    token := client.Subscribe(topic, qos, msgH)
    if token.Wait() && token.Error() != nil{
		log.Println(token.Error())
	}
  log.Printf("Subscribed to topic: %s", topic)
}
