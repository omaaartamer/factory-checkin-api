#!/bin/bash

echo "🔄 Factory Check-in System - Fresh Test Run"
echo "=========================================="

# Reset everything
./scripts/reset-and-test.sh

# Run complete test suite
./scripts/complete-test.sh

echo "=========================================="
echo "🎉 Fresh test run complete!"
