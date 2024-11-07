package jsonrpc2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

var (
	typeContext = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeError   = reflect.TypeOf((*error)(nil)).Elem()
)

type (
	Server struct {
		handler interface{}
		methods map[string]*method

		Log func(level string, msg string)
	}

	method struct {
		name    string
		source  reflect.Value
		params  []reflect.Type
		returns []reflect.Type
	}
)

func NewServer(h any) *Server {
	s := &Server{
		handler: h,
		methods: make(map[string]*method),
		Log: func(level, msg string) {
			log.Println(level, msg)
		},
	}
	s.mustCompile()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.Log("error", err.Error())
	}

	out := s.Call(r.Context(), body)

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	if err := enc.Encode(out); err != nil {
		s.Log("error", err.Error())
	}
}

func (s *Server) Call(ctx context.Context, payload []byte) any {
	// handle batch requests if starts with [ and ends with ]
	if payload[0] == 91 && payload[len(payload)-1] == 93 {
		if len(payload) == 2 {
			return &Response{
				Version: "2.0",
				Error: &Error{
					Code:    -32600,
					Message: "Invalid Request",
				},
			}
		}
		var requests []json.RawMessage
		if err := json.Unmarshal(payload, &requests); err != nil {
			return &Response{
				Version: "2.0",
				Error: &Error{
					Code:    -32700,
					Message: "Parse error",
					Err:     err,
				},
			}
		}
		var (
			wg    sync.WaitGroup
			resps []Response
			mu    sync.Mutex
		)
		wg.Add(len(requests))
		for _, req := range requests {
			go func(r []byte) {
				defer wg.Done()
				resp := s.request(ctx, r)
				if resp != nil {
					mu.Lock()
					resps = append(resps, *resp)
					mu.Unlock()
				}
			}(req)
		}
		wg.Wait()
		if len(resps) == 0 {
			return nil
		}
		return resps
	}

	return s.request(ctx, payload)
}

func (s *Server) request(ctx context.Context, payload []byte) (resp *Response) {
	req := &Request{}
	resp = &Response{Version: "2.0"}

	defer func() {
		if resp.Error != nil && resp.Error.Err != nil {
			s.Log("error", resp.Error.Error())
		}
		if req.ID == nil {
			resp = nil
		}
	}()

	if err := json.Unmarshal(payload, req); err != nil {
		req.ID = -1
		resp.ID = nil
		if strings.Contains(err.Error(), "cannot unmarshal") {
			resp.Error = &Error{
				Code:    -32600,
				Message: "Invalid Request",
				Err:     err,
			}
		} else {
			resp.Error = &Error{
				Code:    -32700,
				Message: "Parse error",
				Err:     err,
			}

		}
		return
	}
	resp.ID = req.ID
	m, ok := s.methods[req.Method]
	if !ok {
		req.ID = -1
		resp.Error = &Error{
			Code:    -32601,
			Message: "Method not found",
		}
		return
	}

	var (
		argTypes   []reflect.Type
		argIndexes []int
	)

	args := make([]reflect.Value, len(m.params))
	for k, v := range m.params {
		switch v {
		case typeContext:
			args[k] = reflect.ValueOf(ctx)
		default:
			argTypes = append(argTypes, v)
			argIndexes = append(argIndexes, k)
		}
	}

	pv := reflect.ValueOf(req.Params)
	if pv.Kind() != reflect.Slice || pv.Len() != len(argIndexes) {
		resp.Error = &Error{
			Code:    -32602,
			Message: "Invalid params",
		}
		return
	}
	if len(argIndexes) > 0 {
		for k, i := range argIndexes {
			t := argTypes[k]
			val := reflect.New(t)
			if err := json.Unmarshal(req.Params[k], val.Interface()); err != nil {
				resp.Error = &Error{
					Code:    -32602,
					Message: "Invalid params",
					Err:     fmt.Errorf("param %d must be %s (%v)", i, t, err),
				}
				return
			}
			args[i] = val.Elem()
		}
	}
	out := m.source.Call(args)
	ot := len(out)

	if ot == 0 {
		return
	}

	lt := len(m.returns)
	if m.returns[lt-1] == typeError {
		errVal := out[lt-1]
		if !errVal.IsNil() {
			err := errVal.Interface().(error)
			if sErr, ok := err.(ServerError); ok {
				resp.Error = &Error{
					Code:    sErr.Code,
					Message: "Server error",
					Data:    sErr.Data,
				}
			} else {
				resp.Error = &Error{
					Code:    -32603,
					Message: "Internal error",
					Err:     errVal.Interface().(error),
				}
			}
			return
		}
		out = out[:lt-1]
	}
	if ot == 1 {
		resp.Result = out[0].Interface()
		return
	}
	var vals []interface{}
	for _, v := range out {
		vals = append(vals, v.Interface())
	}
	resp.Result = vals
	return
}

func (s *Server) mustCompile() {

	tv := reflect.TypeOf(s.handler)
	vv := reflect.ValueOf(s.handler)

	tvt := vv.NumMethod()

	for i := 0; i < tvt; i++ {
		m := tv.Method(i)

		mm := &method{
			name:   m.Name,
			source: vv.Method(i),
		}
		mm.mustParse()
		s.methods[m.Name] = mm
	}
}

func (m *method) mustParse() {
	if m.source.IsValid() {
		mt := m.source.Type()
		if mt.Kind() != reflect.Func {
			panic("method must be of type func")
		}
		for i := 0; i < mt.NumOut(); i++ {
			m.returns = append(m.returns, mt.Out(i))
		}
		for i := 0; i < mt.NumIn(); i++ {
			m.params = append(m.params, mt.In(i))
		}
	}
}
