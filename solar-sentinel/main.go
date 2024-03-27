package main

import (
    mqtt "github.com/eclipse/paho.mqtt.golang"
    "log"
    "time"
	"os"
    irr"github.com/negeek/solar-sphere/solar-sentinel/api/v1"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
    device_id:= strings.Split(msg.Topic, "/")[3]
    log.Println("Device id: ", device_id)
    err:= irr.SaveSolarIrrdianceData(device_id, msg.Payload())

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
        log.Error(token.Error())
    }
    SuscribeToTopic(client, os.Getenv("MQTT_TOPIC"), 0, nil)


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
	// disconnect broker and db
	db.Disconnect(dbctx,dbcancel)
    log.Println("Disconnect from broker")
    client.Disconnect(250)
    log.Println("Shutting down")
    server.Shutdown(ctx)
	os.Exit(0)
}

func SuscribeToTopic(client mqtt.Client, topic string, qos byte, msgH mqtt.MessageHandler) {
    token := client.Subscribe(topic, qos, msgH)
    if token.Wait() && token.Error() != nil{
		log.Error(token.Error())
	}
  log.Printf("Subscribed to topic: %s", topic)
}
