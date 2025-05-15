#!/bin/bash

# Set Go path
export PATH=/usr/local/go/bin:$PATH

# Build the application
echo "Building cfk..."
go build -o cfk ./cmd/cfk

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Run ./cfk to start the application"
else
    echo "Build failed!"
fi