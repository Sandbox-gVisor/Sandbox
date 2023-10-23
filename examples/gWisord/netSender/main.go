package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	for {
		resp, err := http.Get("http://www.boredapi.com/api/activity?key=5881028")
		if err != nil {
			fmt.Println(err)
			continue
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("Response body: %s\n\n", string(bytes))
		time.Sleep(1 * time.Second)
	}
}
