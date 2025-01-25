# Rate Limiter

## Overview

This is a small and straightforward project designed to showcase how a basic rate limiter works. A rate limiter is an essential tool for controlling the number of requests users or systems can make in a given time frame. This project is great for learning the concept or for lightweight use cases where you need to limit traffic without heavy frameworks.

While it's simple, this project highlights the key concepts of rate limiting and how it can be implemented using Go.

## How It Works

1. Tracks incoming requests per client (identified by IP address).
2. If a client exceeds the allowed requests within the set time window, they receive an HTTP `429 Too Many Requests` error.
3. Requests below the limit are processed successfully with a simple success message.

## Getting Started

### Requirements

- Go 1.22 or later.
- `make` installed on your system.

### How to Run

1. Clone the repository:
   ```bash
   git clone https://github.com/gunawanpras/rate-limiter.git
   cd rate-limiter
   ```

2. Run the program using `make`:
   ```bash
   make all
   ```
   The `make file` will builds and starts the server.

3. The rate limiter will start on port `8000`.

### Example Request

Using `curl`, you can test the rate limiter:

```bash
curl -X GET http://localhost:8000/rate-limiter
```

- Successful requests return `Request allowed`.
- Requests exceeding the limit return `429 Too Many Requests` with the message `Rate limit exceeded`.

### Configuration

Edit `config.yaml` to customize the request limit and time interval:

```yaml
rate_limiter:
  limit: 5         # Number of requests allowed
  interval: 10     # Time window in seconds
```

## Why Use This Project?

- Learn how rate limiting works.
- A starting point for more complex rate-limiting solutions.
- Understand basic concurrency handling in Go.

## Contributions

This is just a simple demonstration, but feel free to fork it, enhance it, or add features like:
- Per-user or API-key-based limits.
- Persistent storage for tracking requests.
- Advanced algorithms like token buckets.

## License

MIT License – do whatever you want but don't forget to share your improvements with the community!