#!/bin/bash

echo "🧪 Starting Complete Factory Check-in System Test"

# Start the server in background
echo "🚀 Starting API server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 5

# Test suite
echo "🔬 Running Test Suite..."

echo "Test 1: Health Check"
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.success')
if [ "$HEALTH" = "true" ]; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
fi

echo "Test 2: Employee Check-ins and Check-outs"
EMPLOYEES=("ALICE001" "BOB002" "CHARLIE003" "DIANA004" "EVE005")

for emp in "${EMPLOYEES[@]}"; do
    echo "  Testing employee: $emp"
    
    # Check-in
    CHECKIN=$(curl -s -X POST http://localhost:8080/api/v1/checkin \
        -H "Content-Type: application/json" \
        -d "{\"employee_id\": \"$emp\"}" | jq -r '.event_type')
    
    if [ "$CHECKIN" = "checkin" ]; then
        echo "    ✅ Check-in successful for $emp"
    else
        echo "    ❌ Check-in failed for $emp"
    fi
    
    # Short delay to simulate work time
    sleep 1
    
    # Check-out
    CHECKOUT=$(curl -s -X POST http://localhost:8080/api/v1/checkin \
        -H "Content-Type: application/json" \
        -d "{\"employee_id\": \"$emp\"}" | jq -r '.event_type')
    
    if [ "$CHECKOUT" = "checkout" ]; then
        echo "    ✅ Check-out successful for $emp"
    else
        echo "    ❌ Check-out failed for $emp"
    fi
done

echo "Test 3: Queue Status Check"
sleep 3  # Wait for queue processing
QUEUE_COUNT=$(curl -s http://localhost:8080/api/v1/queue/status | jq -r '.queue.pending_messages')
echo "  Queue has $QUEUE_COUNT pending messages"

echo "Test 4: Employee Status Checks"
for emp in "${EMPLOYEES[@]}"; do
    STATUS=$(curl -s http://localhost:8080/api/v1/employee/$emp/status | jq -r '.status')
    if [ "$STATUS" = "not_checked_in" ]; then
        echo "  ✅ $emp correctly shows as not checked in"
    else
        echo "  ❌ $emp status incorrect: $STATUS"
    fi
done

echo "Test 5: Infrastructure Health"
echo "  🔍 RabbitMQ Status:"
docker exec factory-checkin-rabbitmq rabbitmq-diagnostics status | head -5

echo "  🔍 PostgreSQL Status:"
docker exec factory-checkin-db psql -U checkin_user -d checkin_db -c "SELECT COUNT(*) as checkin_events FROM checkin_events;"
docker exec factory-checkin-db psql -U checkin_user -d checkin_db -c "SELECT COUNT(*) as work_sessions FROM work_sessions;"

echo "🏁 Test Suite Complete!"

# Clean up
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo "✅ All tests completed! Check the logs above for results."
