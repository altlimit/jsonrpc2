package jsonrpc2_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/altlimit/jsonrpc2"
)

type Calculator struct {
}

func (c *Calculator) Add(ctx context.Context, a int, b int) int {
	return a + b
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

func TestNewServer(t *testing.T) {
	ctx := context.Background()
	s := jsonrpc2.NewServer(&Calculator{})
	str := func(r any) string {
		b, _ := json.Marshal(r)
		return string(b)
	}

	testCases := map[string]string{
		`{"jsonrpc": "2.0", "method": "Subtract", "params": [5, 2], "id": 1}`:                                                                     `{"jsonrpc":"2.0","result":3,"id":1}`,
		`{"jsonrpc": "2.0", "method": "add", "params": [5, 2], "id": 1}`:                                                                          `{"jsonrpc":"2.0","error":{"code":-32601,"string":"Method not found"},"id":1}`,
		`{"jsonrpc": "2.0", "method": "Divide", "params": [5, ''], "id": 1}`:                                                                      `{"jsonrpc":"2.0","error":{"code":-32700,"string":"Parse error"},"id":null}`,
		`[{"jsonrpc": "2.0", "method": "Divide", "params": [5, 2], "id": 1},{"jsonrpc": "2.0", "method": "Subtract", "params": [5, 2], "id": 2}]`: `[{"jsonrpc":"2.0","result":3,"id":2},{"jsonrpc":"2.0","result":[2.5],"id":1}]`,
		`[1, 2, 3]`: `[{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null},{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null},{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null}]`,
		`[]`:        `{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null}`,
		`{"username": "admin","password": "admin"}`: `{"jsonrpc":"2.0","error":{"code":-32601,"string":"Method not found"},"id":null}`,
	}

	for k, v := range testCases {
		out := str(s.Call(ctx, []byte(k)))
		if out != v {
			t.Errorf("Expected %s got %s", v, out)
		}
	}
}
