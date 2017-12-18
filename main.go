package main

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"net/http"
	"time"
	"strconv"
	"net"
)

func main() {
	app := cli.NewApp()

	fmt.Println("main")

	app.Name = "bing"
	app.Usage = "bing" + " <url>"
	app.Version = "0.0.1"
	app.Description = "A CLI tool to test concurrent"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "times, t",
			Value: 1000,
			Usage: "How many times you want to test",
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		var (
			url       = c.Args().Get(0)
			ch        = make(chan int, 100)
			count     = 0 // 已发送请求次数
			times     = c.Int("times")
			totalNano = 0
		)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		for i := 0; i < times; i++ {
			start := time.Now()
			go func(i int) {
				if _, err = httpClient.Get(url + "?id=" + strconv.Itoa(i)); err != nil {
					fmt.Println(err)
				}
				end := time.Now()
				diff := end.Nanosecond() - start.Nanosecond()
				ch <- diff
				count++
				if count == times {
					close(ch)
				}
			}(i)
		}

		for nano := range ch {
			totalNano += nano
			fmt.Printf("%v request sended\n", count)
		}

		fmt.Printf("%v request: \n", times)
		fmt.Printf("Total take %vms\n", totalNano/1000/1000)
		fmt.Printf("Average take %vms\n", totalNano/times/1000/1000)

		return
	}

	app.Run(os.Args)
}
