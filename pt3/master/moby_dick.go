package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
	"sync"
)

var defaultURLPattern = "http://www.gutenberg.org/files/15/text/moby-###.txt"
var defaultFilesCount = 135

type WordMap map[string]int

func main() {
	filesCount, err := strconv.Atoi(os.Getenv("FILES_COUNT"))
	if err != nil {
		filesCount = defaultFilesCount
	}

	urlPattern := os.Getenv("URL_PATTERN")
	if len(urlPattern) < 1 {
		urlPattern = defaultURLPattern
	}

	completeWordMap := WordMap{}
	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	replies := make([][]string, filesCount)
	slavePattern, set := os.LookupEnv("SLAVE_PATTERN")
	if !set {
		log.Fatal("slave pattern not set. Cannot lookup slaves. Exiting..")
	}
	slaves := []string{}

	num := 1
	for {
		slaveHostname := strings.ReplaceAll(slavePattern, "#", strconv.Itoa(num))
		_, err := net.LookupIP(slaveHostname)

		if err != nil {
			break
		}
		slaves = append(slaves, slaveHostname)
		num++
	}

	if len(slaves) < 1 {
		log.Fatal("no slaves present")
	}

	for i := 0; i < filesCount; i++ {
		index := i % len(slaves)
		slaveHost := fmt.Sprintf("%s:8133", slaves[index])

		wg.Add(1)
		go wordCount(mutex, i, replies, wg, slaveHost, urlPattern)

	}

	wg.Wait()
	for _, reply := range replies {
		for _, keyVal := range reply {
			kVPair := strings.Split(keyVal, ":")
			key := kVPair[0]
			val, _ := strconv.Atoi(kVPair[1])
			completeWordMap[key] += val
		}

	}

	f, _ := os.Create("data.txt")
	defer f.Close()
	for k, v := range completeWordMap {
		fmtStr := fmt.Sprintf("%s: %d\n", k, v)
		f.WriteString(fmtStr)
	}

	f.Sync()
}

func wordCount(mutex *sync.Mutex, i int, replies [][]string, wg *sync.WaitGroup, slaveUrl string, urlPattern string) {
	paddedNum := fmt.Sprintf("%03d", i)

	urls := []string{}
	url := strings.Replace(urlPattern, "###", paddedNum, -1)
	urls = append(urls, url)

	conn, err := net.Dial("tcp", slaveUrl)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := jsonrpc.NewClient(conn)
	reply := []string{}
	err = c.Call("WordMap.WordCount", urls, &reply)
	if err != nil {
		panic(err)
	}

	mutex.Lock()
	replies[i] = reply
	mutex.Unlock()

	wg.Done()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
