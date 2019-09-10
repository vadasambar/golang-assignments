package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type WordMap map[string]int

func (w *WordMap) WordCount(urls []string, wordCount *[]string) error {
	wordsCh := make([]chan WordMap, len(urls))
	for i, url := range urls {
		wordsCh[i] = make(chan WordMap)
		go readFile(url, wordsCh[i])
	}

	completeWordMap := WordMap{}
	for _, wordMap := range wordsCh {
		for k, v := range <-wordMap {
			completeWordMap[k] += v
		}
	}

	resp := []string{}
	for k, v := range completeWordMap {
		resp = append(resp, fmt.Sprintf("%s:%v", k, v))
	}

	*wordCount = resp
	return nil
}

func main() {

	server := rpc.NewServer()
	wordMap := &WordMap{}
	server.Register(wordMap)

	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	l, e := net.Listen("tcp", ":8133")
	check(e)

	for {
		conn, err := l.Accept()
		check(err)

		go server.ServeCodec(jsonrpc.NewServerCodec(conn))

	}

}

func readFile(url string, wordsCh chan WordMap) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	words := WordMap{}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		words[scanner.Text()]++
	}

	wordsCh <- words
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
