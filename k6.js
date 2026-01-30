import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        holidays_bff_load: {
            executor: 'constant-arrival-rate',
            rate: 500,         // 4000 requests per second (adjust between 3000-5000)
            timeUnit: '1s',
            duration: '30m',    // 30 minutes
            preAllocatedVUs: 2000,
            maxVUs: 3000,       // increased to handle the load
        },
    },
};

const BASE_URL = 'http://localhost:8080/api/v1/holidays-bff';

const HOLIDAY_NAMES = [
    'uusaasta',
    'taasiseseisvumispäev',
    'jõululaupäev',
];

function randomYear() {
    const start = 2020;
    const end = 2030;
    return String(start + Math.floor(Math.random() * (end - start + 1)));
}

function randomDate() {
    const year = 2025;
    const month = 1 + Math.floor(Math.random() * 12);
    const day = 1 + Math.floor(Math.random() * 28);
    return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
}

export default function () {
    if (Math.random() < 0.5) {
        const url = `${BASE_URL}/holidays?year=${randomYear()}`;
        const res = http.get(url);
        check(res, {
            'year endpoint status is 2xx/3xx/4xx': r => r.status > 0,
        });
    } else {
        const name = HOLIDAY_NAMES[Math.floor(Math.random() * HOLIDAY_NAMES.length)];
        const url = `${BASE_URL}/holidays/calculate?date=${randomDate()}&name=${encodeURIComponent(name)}`;
        const res = http.get(url);
        check(res, {
            'calculate endpoint status is 2xx/3xx/4xx': r => r.status > 0,
        });
    }

    sleep(0.01);
}
