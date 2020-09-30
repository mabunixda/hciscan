package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/pkg/errors"
)

var (
	device = flag.String("device", "default", "implementation of ble")
	du     = flag.Duration("du", 5*time.Second, "scanning duration")
	dup    = flag.Bool("dup", true, "allow duplicate reported")
)

var m map[string]time.Time

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func cacheCleanUp(t time.Time) {
	if len(m) == 0 {
		return
	}
	dm := make(map[string]int)
	for mac, ts := range m {
		duration := t.Sub(ts)
		if duration.Minutes() < 1 {
			continue
		}
		dm[mac] = 1
	}
	for mac, _ := range dm {
		delete(m, mac)
	}

}
func main() {

	var url = os.Getenv("ENDPOINT")
	if url == "" {
		fmt.Println("No endpoint env ariable set!")
		return
	}
	flag.Parse()

	m = make(map[string]time.Time)
	go doEvery(15*time.Second, cacheCleanUp)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
	d, err := dev.NewDevice(*device)
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)
	for {

		// Scan for specified durantion, or until interrupted by user.
		fmt.Printf("Scanning for %s...\n", *du)
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
		chkErr(ble.Scan(ctx, *dup, advHandler, nil))

		time.Sleep(time.Second * time.Duration(2))

	}
}

func advHandler(a ble.Advertisement) {
	if a.Connectable() {
		fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
	} else {
		fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
	}
	comma := ""
	if len(a.LocalName()) > 0 {
		fmt.Printf(" Name: %s", a.LocalName())
		comma = ","
	}
	if len(a.Services()) > 0 {
		fmt.Printf("%s Svcs: %v", comma, a.Services())
		comma = ","
	}
	if len(a.ManufacturerData()) > 0 {
		fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	}
	fmt.Printf("\n")
	postMac(a.Addr().String())
}

func postMac(mac string) {
	if _, ok := m[mac]; ok {
		return
	}
	m[mac] = time.Now().UTC()
	url := os.Getenv("ENDPOINT")
	var jsonStr = []byte(`{ "mac": "` + mac + `" }`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return
	}
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Fatalf(err.Error())
	}
}
