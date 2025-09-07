#!/usr/bin/env bash
set -euo pipefail

readonly CERT_DIR="${1:-./certs}"

required_vars=(CLOUDFLARE_API_TOKEN LETSENCRYPT_EMAIL DOMAIN)
for var in "${required_vars[@]}"; do
    [[ -z "${!var:-}" ]] && { echo "🚫 Missing: $var"; exit 1; }
done

echo "🔐 Generating SSL certificates for $DOMAIN using acme.sh"
echo "📁 Output: $CERT_DIR"

mkdir -p "$CERT_DIR"
export CF_Token="$CLOUDFLARE_API_TOKEN"

# Register account if needed
if [[ ! -f "$HOME/.acme.sh/account.conf" ]]; then
    echo "📧 Setting up acme.sh account..."
    acme.sh --register-account -m "$LETSENCRYPT_EMAIL"
fi

echo "🌍 Requesting certificate from Let's Encrypt via Cloudflare DNS..."
acme.sh --issue \
    --dns dns_cf \
    --domain "$DOMAIN" \
    --dnssleep 60 \
    --force

echo "📋 Installing certificates..."
acme.sh --install-cert \
    --domain "$DOMAIN" \
    --cert-file "$CERT_DIR/cert.pem" \
    --key-file "$CERT_DIR/privkey.pem" \
    --fullchain-file "$CERT_DIR/fullchain.pem"

# Set proper permissions
chmod 644 "$CERT_DIR"/*.pem
chmod 600 "$CERT_DIR/privkey.pem"
chown -R 1001:1001 "$CERT_DIR" 2>/dev/null || true

echo "✅ Certificate generation complete!"
