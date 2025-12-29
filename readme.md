# arian-email-parser

arian-parser is designed to parse bank transactions from emails. it just runs an smtp server, so you are expected to set up forwarding rules in the email client. the parser extracts relevant transaction information using bank-specific parsers, and sends over the data to [ariand](https://github.com/xhos/ariand). however the structure of the parser itself is largely decoupled from arian so if you want to send data elsewhere i recon you can do that too with a bit of effort.

## why?

my bank does not have an api or any clean way for accessing my transactions (rbc im looking at you). sometimes they offer csv exports, but those for some reason don't have half the transactions i need, and it's not automatic anyways. so here we are, parsing emails to get the data we need. like cavemen.

## ⚙️ config

### cli params

| param          | description            | default  |
|----------------|------------------------|----------|
| `--smtp-port`  | smtp server port       | `2525`   |
| `--port`       | grpc health port       | `50052`  |

### environment variables

| variable      | description                     | default  | required?  |
|---------------|---------------------------------|----------|------------|
| `API_KEY`     | authentication key for ariand   |          | [x]        |
| `ARIAND_URL`  | ariand grpc service url         |          | [x]        |
| `DOMAIN`      | email domain to serve           |          | [x]        |
| `LOG_LEVEL`   | log level (debug, info, warn)   | `info`   | [ ]        |
| `TLS_CERT`    | tls certificate file path       |          | [ ]        |
| `TLS_KEY`     | tls private key file path       |          | [ ]        |
| `SAVE_EML`    | save incoming emails as .eml    | `false`  | [ ]        |

- the `LOG_LEVEL=debug` env makes it so the incoming email contents are logged, which is useful for accepting forwarding rules

## setup

intended for use with docker compose, instructions are to be added later. #TODO

most email providers, when you set up forwarding, require you to confirm it by clicking a link in the email. you can see the contents of the emails by either setting `SAVE_EML` or simply setting `LOG_LEVEL` to `debug`, that will print all the incoming email contents to logs. this is, of course, intended for one-time forwarding setup, not constant use.

## development

- I highly recommend to use `devenv` for a consistent development environment. it installs all necessary dependencies and provides helpers.
- `SAVE_EML` is useful for developing new parsers, it saves incoming emails as `.eml` files in the `emails/` directory for later inspection.

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

### adding new parsers

it is quite easy to add new parsers for different banks, as long as you know a bit of go/regex, or willing to spend some time prompting it into existence.

1. create a new package under `internal/email/` (e.g., `internal/email/yourbank`).
2. implement the `parser.Parser` interface from `internal/parser/types.go`.
3. register your new parser in an `init()` function within your new package (e.g., `parser.Register(&yourBankParser{})`).
4. add a blank import for your new parser package in `internal/email/all/all.go`.
5. write tests for your new parser, include test data (email objects can be obtained in debug mode).

the added parser should work without any other changes.

contributions are highly welcome, as it's not feasible for me to cover banks I don't use myself.

## 🌱 ecosystem


```definition
arian (n.) /ˈarjan/ [Welsh] Silver; money; wealth.  
```

- [ariand](https://github.com/xhos/ariand) - main backend service
- [arian-web](https://github.com/xhos/arian-web) - frontend web application
- [arian-mobile](https://github.com/xhos/arian-mobile) - mobile appplication
- [arian-protos](https://github.com/xhos/arian-protos) - shared protobuf definitions
- [arian-receipts](https://github.com/xhos/arian-receipts) - receipt parsing microservice
- [arian-email-parser](https://github.com/xhos/arian-email-parser) - email parsing service
