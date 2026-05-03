package main

import (
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func main() {
	// conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	var conn *amqp.Connection
	var err error

	maxRetries := 20

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial("amqp://admin:admin@rabbitmq:5672/")
		if err == nil {
			log.Println("Conectado!")
			break
		}

		log.Println("Erro ao conectar:", err)
		log.Printf("Tentando novamente em 2 segundos... (%d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Erro ao abrir canal:", err)
	}
	defer ch.Close()

	// Declarar a EXCHANGE do tipo TOPIC
	err = ch.ExchangeDeclare("telemetria_exchange", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	// mapa para definir qual Fila escuta qual Tópico
	// A chave é o NOME DA FILA e o valor é a ROUTING KEY (o filtro)
	configFilas := map[string]string{
		"fila_influx":    "sensor.#",          // Recebe TUDO (umidade, temperatura, etc)
		"fila_streaming": "sensor.#",          // Recebe TUDO para o dashboard
		"fila_cache":     "sensor.#",  		   // Recebe TUDO
	}

	for nomeFila, filtro := range configFilas {
		// Declarar a Fila
		q, err := ch.QueueDeclare(nomeFila, true, false, false, false, nil)
		if err != nil {
			log.Printf("Erro ao declarar fila %s: %s", nomeFila, err)
			continue
		}

		// Binding com o FILTRO específico
		err = ch.QueueBind(
			q.Name,
			filtro,                // Aqui vem o filtro (ex: "sensor.*.umidade")
			"telemetria_exchange",
			false,
			nil,
		)
		if err != nil {
			log.Printf("Erro no bind da fila %s com filtro %s: %s", nomeFila, filtro, err)
		} else {
			log.Printf("Fila [%s] vinculada com a chave [%s]", nomeFila, filtro)
		}
	}

	log.Println("Infraestrutura configurada com roteamento seletivo!")
}