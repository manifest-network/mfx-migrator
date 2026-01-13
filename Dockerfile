# Start from the official Golang image to build our application.
FROM golang:1.25 AS builder

# Set the current working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum to download the dependencies.
# This is done before copying the source code to cache the dependencies layer.
COPY go.mod ./
COPY go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed.
RUN go mod download

# Copy the source code into the container.
COPY . .

# Build the Go app as a static binary.
# -o specifies the output file, in this case, the executable name.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mfx-migrator .

# Start from a Debian Slim image to keep the final image size down.
FROM debian:bookworm-slim

# Install cron to schedule the job.
RUN apt-get update && apt-get install -y cron jq && rm -rf /var/lib/apt/lists/*

# The application configuration file should be stored in /mfx-migrator
VOLUME /mfx-migrator

# The job files should be stored in /jobs
VOLUME /jobs

# Copy the pre-built binary file and script from the previous stage.
COPY --from=builder /app/mfx-migrator /usr/local/bin/mfx-migrator
COPY --from=builder /app/scripts/claim_and_migrate.sh /etc/cron.d/claim_and_migrate.sh
RUN chmod +x /etc/cron.d/claim_and_migrate.sh

# Add a cron job to run the script every minute.
RUN echo "* * * * * root /usr/bin/flock -n /tmp/migrator.lock /etc/cron.d/claim_and_migrate.sh >> /var/log/cron.log 2>&1" >> /etc/crontab

# Create a log file to store the cron job output.
RUN touch /var/log/cron.log

# Command to run the executable.
CMD cron && tail -f /var/log/cron.log
