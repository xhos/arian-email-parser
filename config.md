# arian-email-parser config

## cli params

| param         | description             | default |
|---------------|-------------------------|---------|
| `--smtp-port` | SMTP server port        | `2525`  |
| `--port`      | gRPC health server port | `50052` |

## env's

| variable      | description                                 | default | required? |
|---------------|---------------------------------------------|---------|-----------|
| `API_KEY`     | API authentication key                      |         | [x]       |
| `ARIAND_URL`  | ariand gRPC URL                             |         | [x]       |
| `SMTP_DOMAIN` | SMTP server domain                          |         | [x]       |
| `LOG_LEVEL`   | logging level: debug, info, warn, error     | `info`  | [ ]       |
| `TLS_CERT`    | TLS certificate file path                   |         | [ ]       |
| `TLS_KEY`     | TLS private key file path                   |         | [ ]       |
| `DEBUG`       | enable debug mode (skips ariand connection) |         | [ ]       |
