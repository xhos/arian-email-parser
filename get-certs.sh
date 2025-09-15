#!/usr/bin/env bash

# generates Let's Encrypt SSL certificates via Cloudflare DNS for specified domain

set -euo pipefail

CERT_DIR="${1:-./certs}"

required_vars=(CLOUDFLARE_API_TOKEN LETSENCRYPT_EMAIL DOMAIN)
for var in "${required_vars[@]}"; do
    [[ -z "${!var:-}" ]] && { echo "Missing: $var"; exit 1; }
done

mkdir -p "$CERT_DIR"
export CF_Token="$CLOUDFLARE_API_TOKEN"

if [[ ! -f "$HOME/.acme.sh/account.conf" ]]; then
    acme.sh --register-account -m "$LETSENCRYPT_EMAIL"
fi

acme.sh --issue \
    --dns dns_cf \
    --domain "$DOMAIN" \
    --dnssleep 60 \
    --force

acme.sh --install-cert \
    --domain "$DOMAIN" \
    --cert-file "$CERT_DIR/cert.pem" \
    --key-file "$CERT_DIR/privkey.pem" \
    --fullchain-file "$CERT_DIR/fullchain.pem"

chmod 644 "$CERT_DIR"/*.pem
chmod 600 "$CERT_DIR/privkey.pem"
chown -R 1001:1001 "$CERT_DIR" 2>/dev/null || true