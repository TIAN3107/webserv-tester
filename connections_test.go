package main

import (
	"fmt"
	"net/http"
	"testing"
	// "github.com/stretchr/testify/assert"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"bufio"
	// "io"
)

//SET THIS to the server adress you have in your config file
var serverAddr = "127.0.0.1:8080"
//SET THIS to the location of your webserv. Normally it expects to be in 3 folders for example tests/some_other/webserv-tester/testishere.go so webserv to folders above
var webservexe = "../../../webserv"
//SET THIS to where your config file is at, same as before, expects 3 folders above noramlly and then in a config folder
var configfile = "../../../config/test.conf"

type HttpRequest struct{
	name string;
	request string;
	expected int;
}

func TestMain(m *testing.M) {
	cmd := exec.Command(webservexe, configfile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("start webserv: %v", err)
	}
	if err := waitReady(serverAddr, 5*time.Second); err != nil {
		cmd.Process.Kill()
		log.Fatalf("server never came up: %v", err)
	}

	//here I run all the tests
	code := m.Run()

	//this is the cleanup
	cmd.Process.Kill()
	cmd.Wait()
	os.Exit(code)
}


func waitReady(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond) // returns net.Conn
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", addr)
}

func TestSmoke(t *testing.T) {
	fmt.Print("TEST SMOKE HERE{{}}\n")
	conn, err := net.DialTimeout("tcp", serverAddr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	// your request-writing and assertion logic here
}

var bigBody = strings.Repeat("A", 2*1024*1024) // 2 MB of "AAAA..."
var bigBodyLen = strconv.Itoa(len(bigBody))

type Result struct {
	StatusCode	int
	Header		http.Header
	// Body		string
}

func SendRequest(request HttpRequest) (Result, error){
	var res Result
	
	conn, err := net.DialTimeout("tcp", serverAddr, time.Second*3)
	if err != nil {
		return res, fmt.Errorf("dialing %w", err)
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(time.Second*5))
	
	if _, err := conn.Write([]byte(request.request)); err != nil{
		return res, fmt.Errorf("writing: %w", err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil{
		return res, fmt.Errorf("Reading response %w", err)
	}
	res.StatusCode = resp.StatusCode
	res.Header = resp.Header
	// res.Body = string(resp.Body)
	return res, err
}
// var strings


func TestHttp(t *testing.T){
	// requests := []HttpRequest{{name:"200", request:"GET / HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", expected:200}, }
	requests := []HttpRequest{
	// --- 2xx: happy path ---
	{"200_GET_root", "GET / HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 200},
	{"200_GET_file", "GET /index.html HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 200},
	{"200_HEAD", "HEAD / HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 200}, // same as GET, no body
	{"200_GET_autoindex", "GET /autoindex_on_dir/ HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 200},
	{"201_POST_create", "POST /upload/test.txt HTTP/1.1\r\nHost: " + serverAddr + "\r\nContent-Length: 11\r\nContent-Type: text/plain\r\n\r\nhello world", 201},
	{"204_DELETE", "DELETE /upload/test.txt HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 204},

	// --- 3xx: redirects ---
	{"301_redirect", "GET /old-page HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 301},

	// --- 4xx: client errors ---
	{"400_malformed_request_line", "GARBAGE THIS IS NOT HTTP\r\n\r\n", 400},
	{"400_missing_host", "GET / HTTP/1.1\r\n\r\n", 400}, // Host is mandatory in 1.1
	{"400_bad_content_length", "POST /upload/test.txt HTTP/1.1\r\nHost: " + serverAddr + "\r\nContent-Length: banana\r\n\r\nhello world", 400},
	{"403_autoindex_off", "GET /no_index_dir/ HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 403},
	{"404_file_not_found", "GET /definitely/not/here.html HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 404},
	{"404_DELETE_missing", "DELETE /upload/ghost.txt HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 404},
	{"405_method_not_allowed", "PUT / HTTP/1.1\r\nHost: " + serverAddr + "\r\nContent-Length: 0\r\n\r\n", 405},
	{"413_body_too_large", "POST /upload/big.txt HTTP/1.1\r\nHost: " + serverAddr + "\r\nContent-Length: " + bigBodyLen + "\r\n\r\n" + bigBody, 413}, // body > client_max_body_size
	{"414_uri_too_long", "GET /" + strings.Repeat("a", 9000) + " HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 414},

	// --- 5xx: server errors ---
	{"500_cgi_crash", "GET /cgi-bin/broken.py HTTP/1.1\r\nHost: " + serverAddr + "\r\n\r\n", 500},
	{"505_bad_version", "GET / HTTP/2.0\r\nHost: " + serverAddr + "\r\n\r\n", 505},
	}
	conn, err := net.DialTimeout("tcp", serverAddr, time.Second*5)
	if (err != nil){
		t.Fatal(err)
	}
	defer conn.Close()

	for _, val := range requests{
		 SendRequest(val)

	}
}

// func SendReques(serverAddr string, request string) (Result, error) {
//     var res Result

//     conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
//     if err != nil {
//         return res, fmt.Errorf("dial: %w", err)
//     }
//     defer conn.Close()

//     conn.SetDeadline(time.Now().Add(5 * time.Second))

//     if _, err := conn.Write([]byte(request)); err != nil {
//         return res, fmt.Errorf("write: %w", err)
//     }

//     resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
//     if err != nil {
//         return res, fmt.Errorf("read response: %w", err)
//     }
//     defer resp.Body.Close()

//     body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // cap at 1MB
//     if err != nil {
//         return res, fmt.Errorf("read body: %w", err)
//     }

//     res.StatusCode = resp.StatusCode
//     res.Header = resp.Header
//     res.Body = string(body)
//     return res, nil
// }












/*
func TestHttp(t *testing.T){
	type TestCase struct{
		// this is as an example
		name string;
		request string;
		expected int;
		err error;
	}
	tests:= []TestCase{
			{name: "Standard HTTP", request: "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n", expected: 200},

		}
	// var Test1 TestCase
	// Test1.name = "Bye Test"
	// fmt.Println("Hello Test ", Test1.name)
	for _, val := range tests{
		fmt.Println("Hello Test ", val)
	}
	// conn, err := net.Dial("tcp", "golang.org:80")
	// if err != nil {
		// handle error
	// }
	// fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	// status, err := bufio.NewReader(conn).ReadString('\n')
	// t.Run(tests[0].name, func (t *testing.T) {
		// fmt.Println("Hello Test ", tests[0].name)
	// })
}*/
