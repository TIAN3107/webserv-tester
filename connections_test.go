package main

import (
	"fmt"
	"testing"
	// "github.com/stretchr/testify/assert"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

//SET THIS to the server adress you have in your config file
var serverAddr = "127.0.0.1:8080"
//SET THIS to the location of your webserv. Normally it expects to be in 3 folders for example tests/some_other/webserv-tester/testishere.go so webserv to folders above
var webservexe = "../../../webserv"
//SET THIS to where your config file is at, same as before, expects 3 folders above noramlly and then in a config folder
var configfile = "../../../config/test.conf"

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

	code := m.Run()

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
	fmt.Print("TEST SMOKE HER{{}}")
	conn, err := net.DialTimeout("tcp", serverAddr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	// your request-writing and assertion logic here
}

func TestHttp(t *testing.T){
	
}














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
