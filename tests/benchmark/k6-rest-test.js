import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '5s', target: 50 },
    { duration: '10s', target: 200 },
    { duration: '5s', target: 0 },
  ],
};

export default function () {
  const res = http.get('http://localhost:8081/api/v1/drugs/check?drug_code=MED-AMX-500');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has valid json': (r) => r.json().drug_code === 'MED-AMX-500',
  });
}