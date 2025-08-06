# arian-email-parser

If your bank is shitty, and doesn't provide a decent way to visualize your transactions, this is for you.

arian-parser is designed to parse bank transactions from emails. It integrates with Mailpit to fetch emails, extracts relevant transaction information using bank-specific parsers, and stores the structured data into a PostgreSQL database.

What you do with the data is up to you, but it is designed to be used with [arian](https://github.com/xhos/arian) to visualize your transactions, and provide insights into your spending habits.

## setup

Intended for use with docker compose, instructions are to be added later.

## development

I recommend to use `devenv` for a consistent development environment. It installs all necessary dependencies and provides helpers.

### project structure

- `cmd/main.go`: The main entry point for the application.
- `internal/`: Contains the core logic of the application.
  - `db/`: Database interaction logic (PostgreSQL).
  - `email/`: Email parsing logic.
    - `all/`: Imports all specific bank parsers.
    - `rbc/`: Example parser for RBC emails (contains `deposit.go`, `purchase.go`, `withdrawal.go`).
  - `smtp/`: smtp server implementation to receive emails.
  - `parser/`: Defines common interfaces and helpers for email parsers.
- `devenv.nix`: Configuration for the `devenv` development environment.

### debug mode

```shell
docker run -d \
  --name arian-parser-debug \
  -p 25:25 \
  -p 587:587 \
  -p 2525:2525 \
  -p 8080:8080 \
  -v /path/to/your/certs:/certs \
  -v /path/to/debug_emails:/debug_emails \
  -e DEBUG=1 \
  -e SMTP_ADDR=:25 \
  -e SMTP_DOMAIN=yourdomain.com \
  -e TLS_CERT=/certs/cert.pem \
  -e TLS_KEY=/certs/key.pem \
  -e HTTP_ADDR=:8080 \
  ghcr.io/xhos/arian-email-parser:latest
```

### adding new parsers

it is quite easy to add new parsers for different banks, as long as you know a bit of go/regex, or willing to spend some time prompting it into existence.

1. create a new package under `internal/email/` (e.g., `internal/email/yourbank`).
2. implement the `parser.Parser` interface from `internal/parser/types.go`.
3. register your new parser in an `init()` function within your new package (e.g., `parser.Register(&yourBankParser{})`).
4. add a blank import for your new parser package in `internal/email/all/all.go`.
5. write tests for your new parser, include test data (e.g., JSON email fixtures).

the added parser should work without any other changes.

Pull requests for new parsers are welcome, as it's not feasible for me to cover banks I don't use myself.

## arian ecosystem
