#!/bin/bash

set -e

echo "Testing health..."
curl -f http://api-test:7000/api/v1/health

echo "Testing metrics..."
curl -f http://api-test:7000/metrics

echo "All tests passed"