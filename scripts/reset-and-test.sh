#!/bin/bash

echo "🧹 Resetting Factory Check-in System for Fresh Testing..."

# Stop any running containers and remove orphans
echo "Stopping containers..."
docker compose down -v --remove-orphans 2>/dev/null || true

# Clean up data volumes
echo "Cleaning up data volumes..."
docker volume rm factory-checkin-api_postgres_data 2>/dev/null || true
docker volume rm factory-checkin-api_rabbitmq_data 2>/dev/null || true
docker volume rm factory-checkin-api_redis_data 2>/dev/null || true

# Start fresh infrastructure
echo "🚀 Starting fresh infrastructure..."
docker compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 20

# Check if services are healthy
echo "🔍 Checking service health..."
docker compose ps

# Test RabbitMQ connection
echo "📡 Testing RabbitMQ connection..."
docker exec factory-checkin-rabbitmq rabbitmq-diagnostics ping

# Test PostgreSQL connection
echo "🗄️ Testing PostgreSQL connection..."
docker exec factory-checkin-db pg_isready -U checkin_user -d checkin_db

echo "✅ Infrastructure ready for testing!"
