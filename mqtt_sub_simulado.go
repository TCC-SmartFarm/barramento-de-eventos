package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Estrutura de dados organizada
type Telemetria struct {
	SensorID  string    `json:"sensor_id"`
	Valor     float64   `json:"valor"`
	Tipo      string    `json:"tipo"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Simulando um dado de umidade
	dado := Telemetria{
		SensorID:  "ESP32-01",
		Valor:     65.2,
		Tipo:      "umidade",
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(dado)

	// Como o Binding da fila é "sensor.#", essa mensagem VAI CAIR LÁ.
	// routingKey := "sensor.ESP32-01.umidade"
	routingKey := "sensor.%s.%s", dado.SensorID, dado.Tipo

	err = ch.PublishWithContext(context.Background(),
		"telemetria_exchange", // Publicamos na EXCHANGE que criamos
		routingKey,            // A etiqueta de roteamento
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Fatalf("Erro ao publicar: %s", err)
	}

	log.Printf(" [x] Enviado Tópico: %s | Conteúdo: %s", routingKey, body)
}