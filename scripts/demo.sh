#!/bin/sh
set -eu

curl --fail-with-body \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{"work_order_id":"WO-1842","photo":{"object_key":"orders/WO-1842/panel.jpg","caption":"Open electrical panel"},"technicians":[{"id":"tech-b","display_name":"Mina","online":true,"open_jobs":3},{"id":"tech-a","display_name":"Jo","online":true,"open_jobs":1}]}' \
  http://localhost:8080/dispatch
