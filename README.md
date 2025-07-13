# rinha-2025-api-go

Test command:
```bash
k6 run \
  -e K6_WEB_DASHBOARD=true \
  -e K6_WEB_DASHBOARD_PORT=5665 \
  -e K6_WEB_DASHBOARD_PERIOD=2s \
  -e K6_WEB_DASHBOARD_OPEN=false \
  -e K6_WEB_DASHBOARD_EXPORT='report.html' \
  -e MAX_REQUESTS=80 rinha.js
```