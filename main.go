package main

import (
  "context"
  "flag"
  "io"
  "io/ioutil"
  "math/rand"
  "net/http"
  "net/url"
  "net"
  "log"
  "time"
  "strconv"
  "strings"
  "sync"
)

var dns map[string][]string

var dialer = &net.Dialer{
  Timeout:   30 * time.Second,
  KeepAlive: 30 * time.Second,
  DualStack: true,
}

func resolver(addr string, duration time.Duration, stopCh chan bool){
  t := time.NewTicker(duration)
  for {
    select {
    case <-t.C:
      log.Println("net.LookupHost:", addr)
      addrs, err := net.LookupHost(addr)
      if err != nil {
        log.Println("Resolution error ", err)
        return
      }
      prev_len := len(dns[addr])
      dns[addr] = append(dns[addr], addrs...)
      dns[addr] = dns[addr][prev_len:]
    case <- stopCh:
      log.Println("Stop resolver")
      return
    default:
    }
  }
  t.Stop()
}

func runner(runId int64, stopCh chan struct{}, wg *sync.WaitGroup, url string, duration time.Duration){
  before_duration := time.Duration(rand.Float64() * 1000) * time.Millisecond
  log.Println(before_duration)
  time.Sleep(before_duration)

  defer func() { wg.Done() }()
  t := time.NewTicker(duration)
  for {
    select {
    case <-t.C:
      t := http.DefaultTransport.(*http.Transport).Clone()
      t.DisableKeepAlives = true
      client := &http.Client{Transport: t}
      res, err := client.Get(url)
      if err != nil {
        log.Println(err)
        return
      }
      // All HTTP body read
      io.Copy(ioutil.Discard, res.Body)
      res.Body.Close()
    case <- stopCh:
      log.Printf("[Runner %d] stop request received", runId)
      return
    default:
    }
  }
  t.Stop()
}


func main() {
  rand.Seed(time.Now().UnixNano())

  // Flag parser
  flag.Parse()
  purl := flag.Arg(0)

  /// Parse target Rps
  targetRps, err := strconv.ParseInt(flag.Arg(1), 10, 64)
  if err != nil {
    log.Println(err)
    return
  }

  /// Parse time to target Rps
  targetSeconds, err := time.ParseDuration(flag.Arg(2))
  if err != nil {
    log.Println(err)
    return
  }

  /// Parse time exit after target Rps
  exitSeconds, err := time.ParseDuration(flag.Arg(3))
  if err != nil {
    log.Println(err)
    return
  }


  duration := targetSeconds / time.Duration(targetRps)
  log.Printf("duration: %s", duration)
  u, err := url.Parse(purl)
  if err != nil {
    log.Println(err)
    return
  }


  // Create http.DefaultTransport
  dns = map[string][]string{}
  http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
    hp := strings.Split(addr, ":")
    if _, ok := dns[hp[0]]; ok {
        selector := rand.Intn(len(dns[hp[0]]))
        addr = dns[hp[0]][selector] + ":" + hp[1]
    }
    return dialer.DialContext(ctx, network, addr)
  }
  log.Printf("url: %s\n", purl)
  log.Printf("targetRps: %d\n", targetRps)
  log.Printf("targetSeconds: %s\n", targetSeconds)
  log.Printf("exitSeconds: %s\n", exitSeconds)
  log.Printf("duration: %s\n", duration)


  // Start Resolver
  res_stop := make(chan bool)
  go resolver(u.Host, 5 * time.Second, res_stop)

  // Start runner
  stopCh := make(chan struct{})
  wg := sync.WaitGroup{}
  ticker := time.NewTicker(duration)
  for i := int64(1); i< targetRps; {
    select {
    case <-ticker.C:
      log.Printf("%d request/s\n", i)
      wg.Add(1)

      go runner(i, stopCh, &wg, purl, 1 * time.Second)
      i++
    }
  }
  ticker.Stop()

  // Stop runner
  time.Sleep(exitSeconds * time.Second)

  /// Send stop signal to all runner
  res_stop <- true
  close(stopCh)

  /// Check that all runners have stopped 
  wg.Wait()
}
