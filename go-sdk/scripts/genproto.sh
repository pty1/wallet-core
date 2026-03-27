#!/usr/bin/env bash
# Generate Go protobuf files for Trust Wallet Core
# Usage: ./genproto.sh

set -e

cd "$(dirname "$0")/.."
mkdir -p pkg/proto

PROTO_DIR="../src/proto"
OUT_DIR="pkg/proto"

# Build M options for all proto files
M_OPTS=""
for f in "$PROTO_DIR"/*.proto; do
    name=$(basename "$f")
    pkg=$(echo "${name%.proto}" | tr '[:upper:]' '[:lower:]')
    M_OPTS="$M_OPTS --go_opt=M$name=github.com/trustwallet/go-wallet-core/pkg/proto/$pkg"
done

# Generate each proto file
for f in "$PROTO_DIR"/*.proto; do
    name=$(basename "$f")
    pkg=$(echo "${name%.proto}" | tr '[:upper:]' '[:lower:]')
    mkdir -p "$OUT_DIR/$pkg"
    protoc -I="$PROTO_DIR" --go_out="$OUT_DIR/$pkg" --go_opt=paths=source_relative $M_OPTS "$f"
    echo "✓ $name"
done

echo ""
echo "Generated $(find "$OUT_DIR" -name '*.pb.go' | wc -l) files"
