#!/bin/bash

# Get names of running Docker containers
containers=($(docker ps --format "{{.Names}}"))

# Check if there are no running containers
if [ ${#containers[@]} -eq 0 ]; then
    echo "No running containers found"
    exit 1
fi

# Display selection menu
echo "Which container would you like to attach to?"
PS3="Please select an option : "

select container in "${containers[@]}"; do
    if [[ -n "$container" ]]; then
        docker container exec -it $container /bin/bash
        break
    else
        echo "Invalid selection. Please try again."
    fi
done
