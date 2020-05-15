package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var m map[string]time.Time

func prepareHCI() {
	prepareCmd := exec.Command("sh", "-c", "hciconfig hci0 down")
	prepareCmd.Run()
	prepareCmd = exec.Command("sh", "-c", "hciconfig hci0 up")
	prepareCmd.Run()
}

func postMac(mac string) {
	if _, ok := m[mac]; ok {
		return
	}
	fmt.Println("detected mac: " + mac)
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

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func cleanUp(t time.Time) {
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
	prepareHCI()
	m = make(map[string]time.Time)
	go doEvery(15*time.Second, cleanUp)
	re := regexp.MustCompile(`(([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2})`)
	cmd := exec.Command("sh", "-c", "stdbuf -oL hcitool lescan --duplicates")
	out, _ := cmd.StdoutPipe()
	cmd.Start()
	r := bufio.NewReader(out)
	go func() {
		err := cmd.Wait()
		fmt.Println("Stopped - what? %v", err)
	}()
	line, _, err := r.ReadLine()
	fmt.Println(err)
	for err == nil {
		mac := re.FindString(string(line))
		if mac != "" {
			postMac(mac)
		}
		line, _, err = r.ReadLine()
	}
}
