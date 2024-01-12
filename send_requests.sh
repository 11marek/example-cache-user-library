#!/bin/bash

# Start with an empty JSON object
echo "{" > output.json

# Loop generating random unique identifiers and making HTTP requests
for ((i=1; i<=1000; i++)); do
  # Generate a random unique identifier from 1 to 100
  userID=$(shuf -i 1-100 -n 1)

  # Make an HTTP request with the generated identifier
  response=$(curl -s -X GET "http://localhost:8080/user/$userID")

  # Add a new entry to the JSON object
  echo -n "\"$userID\": $response" >> output.json

  # Add a comma unless it's the last entry
  if [ $i -lt 100 ]; then
    echo "," >> output.json
  else
    echo "" >> output.json
  fi
done

# Close the JSON object
echo "}" >> output.json