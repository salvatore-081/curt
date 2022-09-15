# Curt

### What is it

Free and open-source self-hosted url shortener service written in Go.

### Quick start

Run in a terminal:

```
docker run -v curt-db:/data --name curt -p 19000:19000 -e PORT=19000 -e LOG_LEVEL=DEBUG -e API_KEY=your_apy_key -e HOST=http://localhost:19000/ salvatoreemilio/curt:latest
```

### Examples

- [compose.yaml](./examples/compose.yaml)

### License

[Apache License 2.0](https://raw.githubusercontent.com/salvatore-081/curt/main/LICENSE)
