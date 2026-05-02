# Usamos a versão management para ter o painel web
FROM rabbitmq:3-management

# Expomos as portas padrão
EXPOSE 5672 15672

LABEL maintainer="Murilo - SmartFarm Mauá"
LABEL description="RabbitMQ customizado para Barramento de Eventos"