#!/bin/bash

TEST_MESSAGE="TEST_ECHO_MESSAGE"
SERVER_PORT=12345

RESPONSE=$(docker run --rm --network tp0_testing_net alpine:latest \
  sh -c "echo $TEST_MESSAGE | nc -w 5 server $SERVER_PORT")

if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
  echo "action: test_echo_server | result: success"
  exit 0
else
  echo "action: test_echo_server | result: fail"
  exit 1
fi
