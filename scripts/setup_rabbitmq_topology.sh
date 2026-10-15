#!/bin/sh
set -e

echo "[DevOps] Mengonfigurasi Exchange, Queue, Binding & DLQ di MedCore Broker..."

# 1. Deklarasi Topic Exchange Utama
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare exchange name=medcore.clinical.events type=topic durable=true

# 2. Deklarasi Dead Letter Exchange (DLX) & Dead Letter Queue (DLQ)
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare exchange name=medcore.deadletter.exchange type=fanout durable=true
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare queue name=pharmacy.dlq durable=true
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare binding source=medcore.deadletter.exchange destination=pharmacy.dlq

# 3. Deklarasi Main Queues (pharmacy dikaitkan ke DLX)
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare queue name=pharmacy.order.q durable=true arguments='{"x-dead-letter-exchange":"medcore.deadletter.exchange"}'
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare queue name=billing.invoice.q durable=true
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare queue name=notification.all.q durable=true

# 4. Binding Antrean ke medcore.clinical.events sesuai Topologi
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare binding source=medcore.clinical.events destination=pharmacy.order.q routing_key="clinical.*.completed"
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare binding source=medcore.clinical.events destination=billing.invoice.q routing_key="clinical.*.completed"
docker exec medcore-message-broker rabbitmqadmin -u admin_medcore -p SecuredPassRabbit2026! -V /medcore_vhost declare binding source=medcore.clinical.events destination=notification.all.q routing_key="clinical.#"

echo "[DevOps] Seluruh topologi AMQP berhasil dibuat."
