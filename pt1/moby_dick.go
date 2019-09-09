package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const url string = "http://www.gutenberg.org/files/15/text/moby-000.txt"

func main() {
	resp, err := http.Get(url)
	checkErr(err)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanWords)

	words := make(map[string]int)
	for scanner.Scan() {
		word := scanner.Text()
		word = strings.ToLower(word)
		words[word] += 1
	}

	f, fileError := os.Create("data.txt")
	checkErr(fileError)

	for k, v := range words {
		stringFmt := fmt.Sprintf("%s: %d\n", k, v)
		f.WriteString(stringFmt)
	}
	f.Sync()

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}

}
