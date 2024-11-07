![Run Tests](https://github.com/altlimit/jsonrpc2/actions/workflows/run-tests.yaml/badge.svg)

# restruct

jsonrpc2 is an implementation of jsonrpc2 protocol from a golang struct.

---
* [Install](#install)
* [Examples](#examples)
---

## Install

```sh
go get github.com/altlimit/jsonrpc2
```

## Examples

Refer to cmd/example for a running example.

You are allowed to have any or no parameters. Context is optional and doesn't count in your jsonrpc request.
Returning 2 values with error as last return type will automatically respond proper error response.
Returning more than 2 values with or without errors returns an array of result.
```go

type Calculator struct {
}

func (c *Calculator) Subtract(ctx context.Context, a float64, b float64) float64 {
	return a - b
}

func (c *Calculator) Divide(a float64, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}
	return a / b, nil
}

func main() {
	http.Handle("/rpc", jsonrpc2.NewServer(&Calculator{}))
	port := "8090"
	log.Println("Listening", port)
	http.ListenAndServe(":"+port, nil)
}

```

## License

MIT