# Using Ubuntu 20.04 as the base image
FROM ubuntu:20.04

# Copy the connectify executable file from the local directory to the /app/connectify path in the image
COPY connectify /app/connectify

# Copy the config directory to the image
COPY config /app/config

# Set the working directory to /app (the default working directory for subsequent commands)
WORKDIR /app

# Specify the command to run the connectify application when the container starts
CMD ["/app/connectify"]