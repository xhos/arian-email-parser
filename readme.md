# arian-parser
If your bank is shitty, and doesn't provide a decent way to visualize your transactions, this is for you.

arian-parser is designed to parse bank transactions from emails. It integrates with Mailpit to fetch emails, extracts relevant transaction information using bank-specific parsers, and stores the structured data into a PostgreSQL database. 

What you do with the data is up to you, but it is designed to be used with [arian](https://github.com/xhos/arian) to visualize your transactions, and provide insights into your spending habits.

## Setup
Intended for use with docker compose, instructions are to be added later.

## Development

I recommend to use `devenv` for a consistent development environment. It installs all necessary dependencies and provides helpers.

## Project Structure

*   `cmd/main.go`: The main entry point for the application.
*   `internal/`: Contains the core logic of the application.
    *   `db/`: Database interaction logic (PostgreSQL).
    *   `email/`: Email parsing logic.
        *   `all/`: Imports all specific bank parsers.
        *   `rbc/`: Example parser for RBC emails (contains `deposit.go`, `purchase.go`, `withdrawal.go`).
    *   `ingest/`: Orchestrates the process of fetching emails and saving transactions.
    *   `mailpit/`: Client for interacting with the Mailpit API.
    *   `parser/`: Defines common interfaces and helpers for email parsers.
*   `devenv.nix`: Configuration for the `devenv` development environment.

## Adding New Parsers

It is quite easy to add new parsers for different banks, as long as you know a bit of Go, or willing to spend some time prompting it into existence.

1.  Create a new package under `internal/email/` (e.g., `internal/email/yourbank`).
2.  Implement the `parser.Parser` interface from `internal/parser/types.go`.
3.  Register your new parser in an `init()` function within your new package (e.g., `parser.Register(&yourBankParser{})`).
4.  Add a blank import for your new parser package in `internal/email/all/all.go`.
5.  Write tests for your new parser, including test data (e.g., JSON email fixtures).

The added parser should work without any other changes.

Pull requests for new parsers are welcome, as it's not feasible for me to cover banks I don't use myself.