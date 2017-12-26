package main

import (
  "github.com/urfave/cli"
  "fmt"
  "os"
  "net/http"
  "time"
  "strconv"
  "net"
  "crypto/tls"
  "math"
  "github.com/cheggaaa/pb"
)

func main() {
  app := cli.NewApp()

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
          Timeout:   30 * time.Second,
          KeepAlive: 30 * time.Second,
        }).Dial,
        TLSClientConfig: &tls.Config{
          InsecureSkipVerify: true,
        },
        TLSHandshakeTimeout:   30 * time.Second,
        ResponseHeaderTimeout: 30 * time.Second,
        ExpectContinueTimeout: 30 * time.Second,
      },
    }

    bar := pb.StartNew(times)

    // FIX: 并发修改数据， 导致数据不正确， 应该上锁
    for i := 0; i < times; i++ {
      start := time.Now()
      go func(i int) {
        if _, err = httpClient.Get(url + "?id=" + strconv.Itoa(i)); err != nil {
          fmt.Println(err)
        }
        end := time.Now()
        diff := math.Abs(float64(end.Nanosecond() - start.Nanosecond()))
        ch <- int(diff)
        count++
        bar.Increment()
        if count == times {
          fmt.Println("end close chanel")
          close(ch)
        }
      }(i)
    }

    for nano := range ch {
      totalNano += nano
    }

    bar.FinishPrint("")

    fmt.Printf("%v times request: \n", times)
    fmt.Printf("Total take %vms\n", totalNano/1000/1000)
    fmt.Printf("Average take %vms\n", totalNano/times/1000/1000)

    return
  }

  app.Run(os.Args)
}
