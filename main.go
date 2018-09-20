package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/ants"
	"log"
	"os/exec"
	"strings"
	"sync"
	"os"
	"os/signal"
	"syscall"
)

var poolSize int

var pool *ants.Pool

var wg sync.WaitGroup

func init() {
	flag.IntVar(&poolSize, "pool", 20000, "Pool Size")
	flag.Parse()

	var err error
	pool, err = ants.NewPool(poolSize)
	if err != nil {
		log.Fatalln(err)
	}
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

func addTask(task func(args []interface{}) error, args ...interface{}) {
	wg.Add(1)
	pool.Submit(func() error {
		err := task(args)
		wg.Done()
		return err
	})
}

func main() {
	defer pool.Release()

	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT)
		<-ch
		pool.Release()
		log.Println("Stopped.")
		os.Exit(1)
	}()

	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			addTask(func(args []interface{}) error {
				ping(args[0].(string))
				return nil
			}, "192.168." + fmt.Sprintf("%d", i) + "." + fmt.Sprintf("%d", j))

			for k := 0; k < 256; k = k + 2 {
				addTask(func(args []interface{}) error {
					ping(args[0].(string))
					return nil
				}, "10." + fmt.Sprintf("%d", i) + "." + fmt.Sprintf("%d", j) + "." + fmt.Sprintf("%d", k))
				addTask(func(args []interface{}) error {
					ping(args[0].(string))
					return nil
				}, "10." + fmt.Sprintf("%d", i) + "." + fmt.Sprintf("%d", j) + "." + fmt.Sprintf("%d", k+1))
			}
		}
	}

	wg.Wait()
}
