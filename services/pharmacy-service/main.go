package main

import (
	"pharmacy-service/consumer"
)

func main() {
	// Gunakan %2F untuk encode karakter slash '/' pada vhost
	amqpURI := "amqp://admin_medcore:SecuredPassRabbit2026!@localhost:5672/%2Fmedcore_vhost"
	consumer.StartPrescriptionConsumer(amqpURI)
}