package main

import (
	"context"
	"log"
	"time"

	pb "appointment-service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func callCheckDrug(client pb.PharmacyServiceClient, drugCode string, qty int32) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	log.Printf("[RPC Call] Menguji obat: %s (Qty: %d)", drugCode, qty)
	resp, err := client.CheckDrugAvailability(ctx, &pb.CheckDrugRequest{
		DrugCode:       drugCode,
		QuantityNeeded: qty,
	})

	if err != nil {
		st, _ := status.FromError(err)
		log.Printf("[RPC Error Diterima] Code: %s | Message: %s\n", st.Code(), st.Message())
		return
	}

	log.Println("================== RESPON gRPC SUKSES ==================")
	log.Printf("Kode Obat    : %s", resp.GetDrugCode())
	log.Printf("Tersedia     : %t", resp.GetIsAvailable())
	log.Printf("Stok Aktual  : %d", resp.GetCurrentStock())
	log.Printf("Harga Satuan : Rp %.2f", resp.GetUnitPrice())
	log.Printf("Catatan      : %s", resp.GetMessage())
	log.Println("========================================================")
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Tidak dapat membentuk koneksi ke Pharmacy Service: %v", err)
	}
	defer conn.Close()

	client := pb.NewPharmacyServiceClient(conn)

	// Kasus 1: Request Valid (Stok Ada di MongoDB)
	callCheckDrug(client, "MED-AMX-500", 20)

	// Kasus 2: Kode Obat Tidak Ada (codes.NotFound)
	callCheckDrug(client, "MED-OBAT-PALSU", 5)

	// Kasus 3: Melebihi Kuota (codes.ResourceExhausted)
	callCheckDrug(client, "MED-AMX-500", 9999)
}