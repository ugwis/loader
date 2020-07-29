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

func runner(stopCh chan struct{}, wg *sync.WaitGroup, url string, duration time.Duration){
  defer func() { wg.Done() }()
  t := time.NewTicker(duration)
  for {
    select {
    case <-t.C:
      res, err := http.Get(url)
      defer func() {
        // HTTP keepalive requires that all HTTP body read in Go
        io.Copy(ioutil.Discard, res.Body)
        res.Body.Close()
      }()
      if err != nil {
        log.Println(err)
        return
      }
    case <- stopCh:
      log.Println("stop request received")
      return
    default:
    }
  }
  t.Stop()
}


func main() {

  rand.Seed(time.Now().UnixNano())
  duration:=1

  dns = map[string][]string{}
  http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
    hp := strings.Split(addr, ":")
    if _, ok := dns[hp[0]]; ok {
        selector := rand.Intn(len(dns[hp[0]]))
        addr = dns[hp[0]][selector] + ":" + hp[1]
    }
    return dialer.DialContext(ctx, network, addr)
  }
  http.DefaultTransport.(*http.Transport).MaxIdleConns = 35565
  http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 35565

  flag.Parse()
  purl := flag.Arg(0)
  targetRps, err := strconv.Atoi(flag.Arg(1))
  if err != nil {
    log.Println(err)
    return
  }
  u, err := url.Parse(purl)
  if err != nil {
    log.Println(err)
    return
  }

  log.Printf("url: %s\n", purl)
  log.Printf("targetRps: %\n", targetRps)

  // Start Resolver
  res_stop := make(chan bool)
  go resolver(u.Host, 5 * time.Second, res_stop)

  // Start requester
  stopCh := make(chan struct{})
  wg := sync.WaitGroup{}
  for i := 0 ; i< targetRps; i++ {
    log.Printf("%d request/s\n", i)
    wg.Add(1)
    go runner(stopCh, &wg, purl, 1 * time.Second)
    time.Sleep(time.Duration(rand.Intn(1000 * duration)) * time.Millisecond)
  }

  // Stop requester
  time.Sleep(60 * time.Second)
  res_stop <- true
  close(stopCh)
  wg.Wait()
}
