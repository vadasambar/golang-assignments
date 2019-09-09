package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var urlPattern = "http://www.gutenberg.org/files/15/text/moby-###.txt"

type wordMap map[string]int

func main() {
	filesCount := 135
	wordsCh := make([]chan wordMap, 135)
	completeWordMap := wordMap{}

	for i := 0; i < filesCount; i++ {
		wordsCh[i] = make(chan wordMap)
		go readFile(i, wordsCh[i])
	}

	for _, wordMap := range wordsCh {
		for k, v := range <-wordMap {
			completeWordMap[k] += v
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

func readFile(fileNum int, wordsCh chan wordMap) {
	paddedNum := fmt.Sprintf("%03d", fileNum)
	url := strings.Replace(urlPattern, "###", paddedNum, -1)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	words := wordMap{}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		word = strings.ToLower(word)
		words[word]++
	}

	wordsCh <- words
}
