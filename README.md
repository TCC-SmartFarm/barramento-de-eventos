# Documentação do Barramento de Eventos
Esta documentação descreve a implementação da camada de mensageria utilizando Go e RabbitMQ (AMQP 0-9-1). A arquitetura é baseada em eventos e utiliza uma Topic Exchange para roteamento inteligente de dados de sensores.
## 1. Infraestrutura (RabbitMQ via Docker)
O RabbitMQ atua como o Message Broker central. Ele é executado via Docker para garantir isolamento e facilidade de deploy.

#### Arquivo: `docker-compose.yml` 

```YAML
services:
  rabbitmq:
    image: rabbitmq:3-management
    container_name: maua-rabbitmq
    ports:
      - "5672:5672"   # Porta de dados AMQP
      - "15672:15672" # Porta do Painel de Gerenciamento HTTP
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin
```


## 2. Configuração da Rede (Config Rabbit)
Este serviço é responsável por declarar a topologia da rede no RabbitMQ: a Exchange, as Filas e as regras de ligação (Bindings) com filtros por tópico.
#### Arquivo: `config_rabbit.go`
- Exchange: `telemetria_exchange` (tipo: `topic`).
- Filas: `fila_influx`, `fila_streaming`, `fila_cache`.
- Roteamento (Binding Keys):
  - `sensor.#`: Captura todos os dados para histórico e tempo real.
  - `sensor.*.umidade`: Filtra apenas dados de umidade para a camada de cache. Isso é so um exemplo de filtragem!!!


## 3. Publisher (mqtt_sub_simulado)
Simula o comportamento do serviço que receberia dados via MQTT e os injeta no barramento RabbitMQ. Ele "carimba" cada mensagem com uma Routing Key específica.

#### Arquivo: `mqtt_sub_simulado.go`

```Go
// Exemplo de chaves de roteamento utilizadas:
// - "sensor.esp32_01.umidade"
// - "sensor.esp32_01.temperatura"
```

O dado é enviado em formato JSON, contendo `sensor_id`, `valor`, `tipo` e `timestamp` (essas chaves vao ser adaptadas).
```JSON
{
  "sensor_id": "string",
  "valor": "float64",
  "tipo": "string",
  "timestamp": "time.Time"
}
```

## 4. Consumidor (Consumer Multi-Fila)
O serviço de consumo utiliza Goroutines para processar múltiplas filas em paralelo. Cada fila possui um Handler específico para tratar o dado de acordo com seu destino final.
#### Arquivo: `consumer.go`
### Estratégia de Processamento:
| Função           | Fila           | Objetivo                                                     |
|------------------|---------------|--------------------------------------------------------------|
| saveToInfluxDB   | fila_influx   | Persistência em banco de séries temporais (Cloud).          |
| streamToFrontend | fila_streaming| Envio de dados em tempo real via WebSockets.                |
| updateCache      | fila_cache    | Atualização do último estado conhecido no Redis/Cache.      |

## 5. Diagrama de Fluxo Lógico
O fluxo de dados segue a hierarquia abaixo:
- Ingestão: `mqtt_sub_simulado` publica na `telemetria_exchange`.
- Roteamento: A Exchange avalia a Routing Key.
- Distribuição:
  - Mensagens de Qualquer Tipo (`sensor.#`) ⮕ `fila_influx` e `fila_streaming`.
  - Mensagens de Umidade (`sensor.*.umidade`) ⮕ fila_cache. *é so um exemplo*
- Consumo: O `consumer.go` retira as mensagens das filas em paralelo e executa as funções de destino.

## Como Executar

### Pré-requisitos
- Docker e Docker Compose instalados.

- Go 1.20 ou superior.

### Passo a passo:

Subir o Broker:
```Bash
docker compose up -d
```

Configurar a Arquitetura:
```Bash
go run config_rabbit.go
```

Iniciar o Consumo:
```Bash
go run consumer.go
```

Simular Entrada de Dados:
```Bash
go run mqtt_sub_simulado.go
```


------------- 
-------------
-------------

#### justificativa da arquitetura de mais de uma fila para o TCC: 
A escolha desta arquitetura baseia-se no princípio do Desacoplamento. Ao utilizar um Broker de Mensagens, garantimos que o sistema seja Resiliente: caso o serviço de banco de dados (saveToInfluxDB) sofra uma interrupção ou latência na nuvem, as mensagens ficarão retidas com segurança na fila_influx (persistência em disco do RabbitMQ), sem interromper o fluxo de dados em tempo real para o dashboard do usuário. Esta separação permite que o sistema seja resiliente.

#### justificativa do por que escolhi o rabbit: 
O RabbitMQ foi selecionado como o mediador central (Message Broker) devido à sua implementação robusta do protocolo AMQP 0-9-1, que oferece garantias de entrega superiores a protocolos de comunicação direta (como HTTP). Em um cenário de Smart Farming, onde a conectividade no campo pode ser instável, o RabbitMQ atua como um pulmão para o sistema: ele utiliza o conceito de Persistência em Fila, onde as mensagens enviadas pelos sensores são armazenadas em disco ou memória até que os consumidores (Bancos de Dados ou Dashboards) confirmem o processamento através de um Acknowledgment (ACK). Essa característica elimina o acoplamento temporal entre os serviços; ou seja, o sensor não precisa que o banco de dados esteja online no exato momento da leitura. Além disso, o motor de roteamento por tópicos (Topic Exchange) permite que o sistema escale horizontalmente, distribuindo a mesma informação para múltiplos destinos de forma seletiva, garantindo que cada componente da arquitetura (Persistência, Tempo Real e Cache) receba apenas os dados estritamente necessários, otimizando o consumo de banda e processamento do servidor.
