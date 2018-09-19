package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/ants"
	"log"
	"os/exec"
	"strings"
	"sync"
)

var poolSize int

func init() {
	flag.IntVar(&poolSize, "pool", 20000, "Pool Size")
	flag.Parse()
}

func ping(ip string) {
	cmd := exec.Command("ping", ip, "-c", "1", "-W", "5")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(ip + " is blocked")
	} else {
		if strings.Contains(string(output), "100.0% packet loss") {
			fmt.Println(ip + " is blocked")
		} else {
			fmt.Println(ip + " is live")
		}
	}
}

func main() {
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Release()

	var wg sync.WaitGroup

	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			ip := "192.168." + fmt.Sprintf("%d", i) + "." + fmt.Sprintf("%d", j)
			wg.Add(1)
			pool.Submit(func() error {
				ping(ip)
				wg.Done()
				return nil
			})

			for k := 0; k < 256; k = k + 2 {
				ip1 := "10." + fmt.Sprintf("%d", i) + "." + fmt.Sprintf("%d", j) + "." + fmt.Sprintf("%d", k)
				wg.Add(1)
				pool.Submit(func() error {
					ping(ip1)
					wg.Done()
					return nil
				})
				ip2 := "10." + fmt.Sprintf("%d", i) + "." + fmt.Sprintf("%d", j) + "." + fmt.Sprintf("%d", k+1)
				wg.Add(1)
				pool.Submit(func() error {
					ping(ip2)
					wg.Done()
					return nil
				})
			}
		}
	}

	wg.Wait()
}
