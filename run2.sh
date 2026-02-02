#!/bin/sh

# Configuration
API_URL="http://localhost:8080/order"
NUM_REQUESTS=100

echo "Starting to send $NUM_REQUESTS orders to $API_URL"
echo "================================================"

i=1
while [ $i -le $NUM_REQUESTS ]; do
    # Format order ID with leading zeros (order-001, order-002, etc.)
    ORDER_ID=$(printf "order-%03d" $((i+100)))
    
    # Simple sequential item selection (sh doesn't support arrays well)
    ITEM_NUM=$((i % 10))
    case $ITEM_NUM in
        0) ITEM="Laptop" ;;
        1) ITEM="Mouse" ;;
        2) ITEM="Keyboard" ;;
        3) ITEM="Monitor" ;;
        4) ITEM="Headphones" ;;
        5) ITEM="Webcam" ;;
        6) ITEM="USB Cable" ;;
        7) ITEM="SSD Drive" ;;
        8) ITEM="RAM" ;;
        9) ITEM="Motherboard" ;;
    esac
    
    # Simple amount calculation
    AMOUNT=$(awk -v i=$i 'BEGIN{print 10 + (i * 5.5)}')
    AMOUNT=$(printf "%.2f" $AMOUNT)
    
    # Send request
    echo "[$i/$NUM_REQUESTS] Sending order: $ORDER_ID - $ITEM - \$$AMOUNT"
    
    RESPONSE=$(curl -s -X POST $API_URL \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"$ORDER_ID\",\"item\":\"$ITEM\",\"amount\":$AMOUNT}")
    
    echo "Response: $RESPONSE"
    echo "---"
    
    i=$((i + 1))
done

echo "================================================"
echo "Completed sending $NUM_REQUESTS orders!"