package consumer

import (
	"encoding/json"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Idempotency Store sederhana (In-Memory) menggunakan Map & Mutex
var (
	processedEvents = make(map[string]bool)
	eventMutex      sync.Mutex
)

type EventEnvelope struct {
	EventID string `json:"event_id"`
	Payload struct {
		AppointmentID string `json:"appointment_id"`
		Prescription  struct {
			PrescriptionID string `json:"prescription_id"`
			Items          []struct {
				DrugCode string `json:"drug_code"`
				Quantity int    `json:"quantity"`
			} `json:"items"`
		} `json:"prescription"`
	} `json:"payload"`
}

func StartPrescriptionConsumer(amqpURI string) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("[Pharmacy Consumer] Gagal koneksi broker: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[Pharmacy Consumer] Gagal membuka channel: %v", err)
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatalf("[Pharmacy Consumer] Gagal set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		"pharmacy.order.q",
		"pharmacy-worker",
		false, // manual ACK
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[Pharmacy Consumer] Gagal subscribe antrean: %v", err)
	}

	log.Println("[Pharmacy Consumer] Menunggu pesan resep pada antrean: pharmacy.order.q")

	for d := range msgs {
		var evt EventEnvelope
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			log.Printf("[Pharmacy Consumer] Format pesan tidak valid, NACK ke DLQ: %v", err)
			d.Nack(false, false)
			continue
		}

		// 1. Validasi Poison Message (Qty <= 0)
		isPoison := false
		for _, item := range evt.Payload.Prescription.Items {
			if item.Quantity <= 0 {
				isPoison = true
				break
			}
		}

		if isPoison {
			log.Printf("[Pharmacy Consumer] Terdeteksi Poison Message (Qty <= 0) pada Event %s! NACK ke DLQ.", evt.EventID)
			d.Nack(false, false)
			continue
		}

		// 2. Pemeriksaan Idempotensi (Langkah 2.3 Modul)
		eventMutex.Lock()
		if processedEvents[evt.EventID] {
			eventMutex.Unlock()
			log.Printf("[Pharmacy Consumer] [DUPLICATE DETECTED] Event ID %s sudah pernah diproses sebelumnya! Mengabaikan payload dan langsung mengirim ACK.", evt.EventID)
			d.Ack(false) // Langsung ACK agar tidak menumpuk di antrean, tapi logika bisnis tidak diulang
			continue
		}
		// Tandai Event ID sudah diproses
		processedEvents[evt.EventID] = true
		eventMutex.Unlock()

		// 3. Eksekusi Logika Bisnis Normal
		log.Println("---------------------------------------------------------")
		log.Printf("[Pharmacy Consumer] Menerima Resep Baru!")
		log.Printf(" Event ID        : %s", evt.EventID)
		log.Printf(" Appointment ID  : %s", evt.Payload.AppointmentID)
		log.Printf(" Prescription ID : %s", evt.Payload.Prescription.PrescriptionID)
		for _, item := range evt.Payload.Prescription.Items {
			log.Printf(" -> Item: %s | Qty: %d", item.DrugCode, item.Quantity)
		}
		log.Println("---------------------------------------------------------")

		d.Ack(false)
		log.Println("[Pharmacy Consumer] Pesan berhasil di-ACK.")
	}
}