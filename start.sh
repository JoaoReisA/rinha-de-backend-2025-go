#!/bin/bash
docker build -t rinha-backend-payments:latest .
cd payment-processor
docker-compose down
docker-compose up -d
cd ..
docker-compose down
docker-compose up -d
