package main

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222,nats://localhost:4223,nats://localhost:4224",
		nats.Name("NATS Simple Service Endpoint"),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Println("disconnected from NATS server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Println("reconnected to NATS server")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("connection to NATS server closed")
		}),
	)
	if err != nil {
		log.Println("Error connecting to NATS server:", err)
		return
	}
	defer nc.Drain()

	log.Println("Connected to:", nc.ConnectedUrl())

	service, err := micro.AddService(nc, micro.Config{
		Name:        "test",
		Version:     "1.0.0",
		Description: "This the simple service endpoint example",
	})
	if err != nil {
		log.Println("Error creating service endpoint:", err)
		return
	}
	log.Println("Service endpoint 'test' created", service.Info().Name, service.Info().ID)

	group1 := service.AddGroup("group1")

	group1.AddEndpoint("get", micro.HandlerFunc(getHandler))
	group1.AddEndpoint("post", micro.HandlerFunc(postHandler))

	group2 := service.AddGroup("group2")

	group2.AddEndpoint("get", micro.HandlerFunc(getHandler))
	group2.AddEndpoint("post", micro.HandlerFunc(postHandler))

	resp, err := nc.Request("group1.get", []byte(""), nats.DefaultTimeout)
	if err != nil {
		log.Println("Error requesting get service endpoint:", err)
		return
	}
	log.Println("Response from service endpoint:", string(resp.Data))

	resp, err = nc.Request("group2.post", []byte("Sample POST data"), nats.DefaultTimeout)
	if err != nil {
		log.Println("Error requesting post service endpoint:", err)
		return
	}
	log.Println("Response from service endpoint:", string(resp.Data))

}

func getHandler(req micro.Request) {
	var response []byte
	switch req.Subject() {

	case "group1.get":
		log.Println("Received GET request for group1")
		response = []byte("Response from group1 GET endpoint")
	case "group2.get":
		log.Println("Received GET request for group2")
		response = []byte("Response from group2 GET endpoint")
	default:
		log.Println("Unknown GET request subject:", req.Subject())
		response = []byte("Unknown GET request")
	}
	req.Respond(response)
}

func postHandler(req micro.Request) {
	var response []byte
	switch req.Subject() {

	case "group1.post":
		log.Println("Received POST request for group1 with data:", string(req.Data()))
		response = []byte("Response from group1 POST endpoint")
	case "group2.post":
		log.Println("Received POST request for group2 with data:", string(req.Data()))
		response = []byte("Response from group2 POST endpoint")
	default:
		log.Println("Unknown POST request subject:", string(req.Data()))
		response = []byte("Unknown POST request")
	}
	req.Respond(response)
}
