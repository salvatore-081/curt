# Curt

### What is it

Free and open-source self-hosted url shortener service written in Go.

### Quick start

Run in a terminal:

```docker
docker run -v curt-db:/data --name curt -p 19000:19000 -e PORT=19000 -e LOG_LEVEL=DEBUG -e API_KEY=your_api_key -e HOST=http://localhost:19000/ salvatoreemilio/curt:latest
```

Or use a [docker-compose](./examples/compose.yaml) version 
### Examples

- With an API Client send a **POST** request with this body
    ```JSON
    {
        "url":"url_to_shorten"
    }
    ```
    remember to insert an API key in the header if you configured it in the env variable
- Response
    ```JSON
    {
        "curt": "generated_key",
        "url": "http:://localhost:19000/generated_key"
    }
    ```

### License

[Apache License 2.0](https://raw.githubusercontent.com/salvatore-081/curt/main/LICENSE)
