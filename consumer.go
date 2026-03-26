// esse tipo de processamento paralelo é muito comum em sistemas de consumo,
// onde cada consumidor tem uma função específica (ex: salvar no banco, enviar para o dashboard, atualizar cache, etc)
// mas acho que seria interessante criar um consumer para cada fila, ao invés de um consumer genérico que processa tudo

package main

import (
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

	// um canal para cada processo para não haver interferência
	
	// Lançamos os consumidores em GOROUTINES (processamento paralelo)
	go worker(conn, "fila_influx", saveToInfluxDB)
	go worker(conn, "fila_streaming", streamToFrontend)
	go worker(conn, "fila_cache", updateCache)

	log.Printf(" [*] Sistema de Consumo Triplo iniciado. CTRL+C para sair.")
	
	// Mantém o main rodando para as goroutines não morrerem
	select {} 
}

// worker é uma função genérica que conecta em uma fila e executa uma função de tratamento
func worker(conn *amqp.Connection, queueName string, handler func(Telemetria)) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Erro ao abrir canal para %s: %s", queueName, err)
		return
	}
	defer ch.Close()

	q, _ := ch.QueueDeclare(queueName, true, false, false, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	for d := range msgs {
		var dado Telemetria
		json.Unmarshal(d.Body, &dado)
		
		// Executa a função específica (Influx, Streaming ou Cache)
		handler(dado)
	}
}

// --- TRATAMENTOS DIFERENTES ---

func saveToInfluxDB(d Telemetria) {
	log.Printf("[DATABASE] Salvando histórico: Sensor %s (%s) -> InfluxDB", d.SensorID, d.Tipo)
}

func streamToFrontend(d Telemetria) {
	log.Printf("[DASHBOARD] Enviando Real-time: Sensor %s -> WebSocket", d.SensorID)
}

func updateCache(d Telemetria) {
	log.Printf("[REDIS/CACHE] Atualizando Estado Atual: Sensor %s -> Cache", d.SensorID)
}