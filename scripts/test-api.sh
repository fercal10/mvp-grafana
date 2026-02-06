#!/bin/bash

set -e

# With microservices: first arg = accounts-api URL, second arg = transfers-api URL (defaults to first if omitted)
BASE_URL="${1:-http://localhost:8080}"
TRANSFERS_URL="${2:-$BASE_URL}"

echo "🧪 Testing Bank API"
echo "   Accounts:  $BASE_URL"
echo "   Transfers: $TRANSFERS_URL"
echo ""

# Health check
echo "1️⃣  Health check..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# Create first account
echo "2️⃣  Creating account ACC001..."
ACC1=$(curl -s -X POST "$BASE_URL/api/accounts" \
  -H "Content-Type: application/json" \
  -d '{
    "account_number": "ACC001",
    "initial_balance": 1000.0
  }')
echo "$ACC1" | jq '.'
ACC1_ID=$(echo "$ACC1" | jq -r '.id')
echo ""

# Create second account
echo "3️⃣  Creating account ACC002..."
ACC2=$(curl -s -X POST "$BASE_URL/api/accounts" \
  -H "Content-Type: application/json" \
  -d '{
    "account_number": "ACC002",
    "initial_balance": 500.0
  }')
echo "$ACC2" | jq '.'
ACC2_ID=$(echo "$ACC2" | jq -r '.id')
echo ""

# List accounts
echo "4️⃣  Listing all accounts..."
curl -s "$BASE_URL/api/accounts" | jq '.'
echo ""

# Make a transfer
echo "5️⃣  Making transfer from ACC001 to ACC002..."
TRANSFER=$(curl -s -X POST "$TRANSFERS_URL/api/transfers" \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_number": "ACC001",
    "to_account_number": "ACC002",
    "amount": 250.0,
    "description": "Test transfer"
  }')
echo "$TRANSFER" | jq '.'
echo ""

# Get account transactions
echo "6️⃣  Getting transactions for ACC001..."
curl -s "$BASE_URL/api/accounts/$ACC1_ID/transactions" | jq '.'
echo ""

echo "7️⃣  Getting transactions for ACC002..."
curl -s "$BASE_URL/api/accounts/$ACC2_ID/transactions" | jq '.'
echo ""

echo "✅ All tests completed!"
