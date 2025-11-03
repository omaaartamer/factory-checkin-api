#!/bin/bash

echo "🧹 Resetting Factory Check-in System for Fresh Testing..."

# Stop any running containers
echo "Stopping containers..."
docker compose down -v 2>/dev/null || true

# Clean up Redis and Postgres data
echo "Cleaning up data volumes..."
docker volume rm factory-checkin-api_postgres_data 2>/dev/null || true
docker volume rm factory-checkin-api_redis_data 2>/dev/null || true

# Start fresh infrastructure
echo "🚀 Starting fresh infrastructure..."
docker compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 15

# Check if services are healthy
echo "🔍 Checking service health..."
docker compose ps

# Test Redis connection
echo "📡 Testing Redis connection..."
docker exec factory-checkin-redis redis-cli ping

# Test PostgreSQL connection
echo "🗄️ Testing PostgreSQL connection..."
docker exec factory-checkin-db pg_isready -U checkin_user -d checkin_db

echo "✅ Infrastructure ready for testing!"
