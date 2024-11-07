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

func (c *Calculator) Update(ctx context.Context, a int, b int) {

}

func (c *Calculator) NotifyHello() {}

func (c *Calculator) GetData() (string, int) {
	return "hello", 5
}

func (c *Calculator) Sum(ctx context.Context, a []int) int {
	var total int
	for _, v := range a {
		total += v
	}
	return total
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
		`{"jsonrpc": "2.0", "method": "Subtract", "params": [42, 23], "id": 1}`:                           `{"jsonrpc":"2.0","result":19,"id":1}`,
		`{"jsonrpc": "2.0", "method": "Subtract", "params": [23, 42], "id": 2}`:                           `{"jsonrpc":"2.0","result":-19,"id":2}`,
		`{"jsonrpc": "2.0", "method": "Update", "params": [1,2,3,4,5]}`:                                   `null`,
		`{"jsonrpc": "2.0", "method": "foobar", "id": "1"}`:                                               `{"jsonrpc":"2.0","error":{"code":-32601,"string":"Method not found"},"id":"1"}`,
		`{"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]`:                                    `{"jsonrpc":"2.0","error":{"code":-32700,"string":"Parse error"},"id":null}`,
		`{"jsonrpc": "2.0", "method": 1, "params": "bar"}`:                                                `{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null}`,
		`[{"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},{"jsonrpc": "2.0", "method"]`: `{"jsonrpc":"2.0","error":{"code":-32700,"string":"Parse error"},"id":null}`,
		`[]`:      `{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null}`,
		`[1]`:     `[{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null}]`,
		`[1,2,3]`: `[{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null},{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null},{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null}]`,
		`[
			{"jsonrpc": "2.0", "method": "Sum", "params": [[1,2,4]], "id": "1"},
			{"jsonrpc": "2.0", "method": "NotifyHello", "params": [7]},
			{"jsonrpc": "2.0", "method": "Subtract", "params": [42,23], "id": "2"},
			{"foo": "boo"},
			{"jsonrpc": "2.0", "method": "foo.get", "params": {"name": "myself"}, "id": "5"},
			{"jsonrpc": "2.0", "method": "GetData", "id": "9"}
		]`: `[{"jsonrpc":"2.0","result":["hello",5],"id":"9"},{"jsonrpc":"2.0","error":{"code":-32601,"string":"Method not found"},"id":null},{"jsonrpc":"2.0","result":19,"id":"2"},{"jsonrpc":"2.0","error":{"code":-32600,"string":"Invalid Request"},"id":null},{"jsonrpc":"2.0","result":7,"id":"1"}]`,
		`[
			{"jsonrpc": "2.0", "method": "NotifyHello", "params": [1,2,4]},
			{"jsonrpc": "2.0", "method": "Update", "params": [7]}
		]`: `null`,
	}

	for k, v := range testCases {
		out := str(s.Call(ctx, []byte(k)))
		if out != v {
			t.Errorf("Expected %s got %s", v, out)
		}
	}
}
