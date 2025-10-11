package request

import (
	"bytes"
	"fmt"
	"io"

	"github.com/Samudra-G/http-golang/internal/headers"
)

type parserState string
const (
	StateInit parserState = "init"
	StateHeaders parserState = "headers"
	StateDone parserState = "done"
	StateError parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method 		  string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state 		parserState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func parseRequestLine (b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx+len(SEPARATOR) // do NOT include \r\n

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method: string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion: string(httpParts[1]),
	}
	
	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	outer:
	for {
		currentData := data[read:]

		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			} 
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer 
			}

			read += n

			if done {
				r.state = StateDone
			}

		case StateDone:
			break outer
		
		default:
			panic("We wrote poor code . . .")
		}
	}
	return read, nil 
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	//Important: io.ReadAll reads all at once, TCP can send packets slowly over time, hence we need a loop
	buf := make([]byte, 1024) // Note: buffer size could get exceeded, adjust accordingly
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil
}