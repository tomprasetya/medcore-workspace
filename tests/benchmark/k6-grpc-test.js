import grpc from 'k6/net/grpc';
import { check } from 'k6';

const client = new grpc.Client();
client.load(['../../proto'], 'pharmacy.proto');

export const options = {
  stages: [
    { duration: '5s', target: 50 },
    { duration: '10s', target: 100 },
    { duration: '5s', target: 0 },
  ],
};

export default function () {
  // Hubungkan sekali jika belum terkoneksi
  client.connect('127.0.0.1:50051', { plaintext: true });

  const response = client.invoke('medcore.pharmacy.v1.PharmacyService/CheckDrugAvailability', {
    drug_code: 'MED-AMX-500',
    quantity_needed: 10,
  });

  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
    'has valid stock': (r) => r && r.message && (r.message.isAvailable === true || r.message.is_available === true),
  });
}

export function teardown() {
  client.close();
}