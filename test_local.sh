#!/bin/bash

# Local integration test script for lambdabot
# Tests RUNMODE=stdout functionality with various commands

set -e

echo "Building local binary..."
go build -o lambdabot

echo "Testing echo command..."
result=$(echo '{"command":"echo","args":"test message"}' | RUNMODE=stdout ./lambdabot 2>/dev/null)
if echo "$result" | grep -q '"result": "test message"'; then
    echo "✅ Echo command test passed"
else
    echo "❌ Echo command test failed"
    echo "Got: $result"
    exit 1
fi

echo "Testing unknown command..."
result=$(echo '{"command":"unknown","args":"test"}' | RUNMODE=stdout ./lambdabot 2>/dev/null)
if echo "$result" | grep -q '"result": ""'; then
    echo "✅ Unknown command test passed"
else
    echo "❌ Unknown command test failed"
    echo "Got: $result"
    exit 1
fi

echo "Testing full Command struct..."
result=$(echo '{"user":"testuser","source":"testsource","command":"echo","args":"full test"}' | RUNMODE=stdout ./lambdabot 2>/dev/null)
if echo "$result" | grep -q '"user": "testuser"' && echo "$result" | grep -q '"result": "full test"'; then
    echo "✅ Full Command struct test passed"
else
    echo "❌ Full Command struct test failed"
    echo "Got: $result"
    exit 1
fi

echo "Testing JSON marshaling..."
result=$(echo '{"command":"echo","args":"special chars: àáâãäåæçèé"}' | RUNMODE=stdout ./lambdabot 2>/dev/null)
if echo "$result" | grep -q "àáâãäåæçèé"; then
    echo "✅ JSON marshaling test passed"
else
    echo "❌ JSON marshaling test failed"
    echo "Got: $result"
    exit 1
fi

echo ""
echo "🎉 All local integration tests passed!"
echo "RUNMODE=stdout functionality is working correctly."