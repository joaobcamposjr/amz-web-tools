#!/bin/bash

echo "🔍 Verifying all API endpoints are using correct backend..."

BASE_URL="http://52.206.225.24:8080/api/v1"

echo "📋 Testing all backend endpoints:"
echo ""

# Dashboard
echo "🏠 Dashboard:"
curl -s "$BASE_URL/dashboard/stats" | head -c 100 && echo "..." || echo "❌ Dashboard failed"
echo ""

# Auth
echo "🔐 Auth:"
curl -s -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" -d '{"email":"test","password":"test"}' | head -c 100 && echo "..." || echo "❌ Auth failed"
echo ""

# Car Plate
echo "🚗 Car Plate:"
curl -s "$BASE_URL/car-plate/history" | head -c 100 && echo "..." || echo "❌ Car Plate failed"
echo ""

# Users
echo "👥 Users:"
curl -s "$BASE_URL/users" | head -c 100 && echo "..." || echo "❌ Users failed"
echo ""

# DePara
echo "🔄 DePara:"
curl -s "$BASE_URL/depara" | head -c 100 && echo "..." || echo "❌ DePara failed"
echo ""

# Stock
echo "📦 Stock:"
curl -s "$BASE_URL/stock" | head -c 100 && echo "..." || echo "❌ Stock failed"
echo ""

# Profile
echo "👤 Profile:"
curl -s "$BASE_URL/profile" | head -c 100 && echo "..." || echo "❌ Profile failed"
echo ""

echo "✅ All endpoints should be accessible at port 8080!"
echo "🎯 Frontend will now call ALL endpoints at: $BASE_URL"


