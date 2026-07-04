#!/usr/bin/env bash
# Relay integration test — run with: bash test/integration/relay_test.sh
# Requires: curl, a running server at http://localhost:8080

set -euo pipefail

BASE="${1:-http://localhost:8080}"
PASS=0
FAIL=0

check() {
  local desc="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "  PASS: $desc"
    PASS=$((PASS+1))
  else
    echo "  FAIL: $desc (expected $expected, got $actual)"
    FAIL=$((FAIL+1))
  fi
}

check_status() {
  local desc="$1" url="$2" expected="$3"
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" "$url")
  check "$desc" "$expected" "$code"
}

check_header() {
  local desc="$1" url="$2" header="$3" expected="$4"
  local val
  val=$(curl -s -I "$url" | grep -i "^$header:" | sed 's/.*: //' | tr -d '\r\n')
  check "$desc" "$expected" "$val"
}

echo "=== KeyChat Integration Tests ==="
echo "Server: $BASE"
echo

echo "--- Health / Root ---"
check_status "GET /" "$BASE/" "200"

echo "--- Security Headers ---"
check_header "CSP present" "$BASE/" "Content-Security-Policy" ""
check_header "X-Content-Type-Options" "$BASE/" "X-Content-Type-Options" "nosniff"
check_header "X-Frame-Options" "$BASE/" "X-Frame-Options" "DENY"

echo "--- Bridge ---"
check_status "GET /fb/send" "$BASE/fb/send" "200"
check_status "GET /fb/receive" "$BASE/fb/receive" "200"
check_status "GET /fb/confirm" "$BASE/fb/confirm" "200"

echo "--- Education Pages ---"
check_status "GET /learn/crypto" "$BASE/learn/crypto" "200"
check_status "GET /learn/nostr" "$BASE/learn/nostr" "200"
check_status "GET /learn/keys" "$BASE/learn/keys" "200"
check_status "GET /learn/messaging" "$BASE/learn/messaging" "200"
check_status "GET /learn/architecture" "$BASE/learn/architecture" "200"

echo "--- Static Files ---"
check_status "GET /static/css/reset.css" "$BASE/static/css/reset.css" "200"
check_status "GET /static/css/main.css" "$BASE/static/css/main.css" "200"

echo "--- PWA ---"
check_status "GET /manifest.json" "$BASE/manifest.json" "200"
check_status "GET /sw.js" "$BASE/sw.js" "200"

echo
echo "=== Results: $PASS passed, $FAIL failed ==="
exit $FAIL
