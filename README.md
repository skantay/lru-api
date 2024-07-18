# lru-api

## Usage

```sh
go run ./cmd
```

Available flags, also you can create `.env` with the same variables.

- `HTTP_PORT`: Specifies the HTTP port for the server to listen on. Default is 8080.
- `CACHE_SIZE`: Sets the maximum size of the cache. Default is 10.
- `DEFAULT_CACHE_TTL`: Specifies the default time-to-live (in seconds) for cache entries. Default is 60.
- `LOG_LEVEL`: Sets the logging level (DEBUG, INFO, WARN, ERROR). Default is WARN.