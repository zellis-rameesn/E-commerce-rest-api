#!/bin/sh

# Create bucket
awslocal s3 mb s3://ecommerce-uploads

# Create sqs queue
awslocal sqs create-queue --queue-name ecommerce_events

echo "LocalStack initialization complete"