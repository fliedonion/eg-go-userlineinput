package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func inputChannel(input chan string) {
	for {
		in := bufio.NewReader(os.Stdin)
		data, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// input <- data[:len(data)-1]  // Winだと \r が残る
		input <- strings.TrimRight(data, "\r\n")
	}
}

// countdown each second and put log during specified seconds then return.
// This function does not report last (zero).
func countDown(timeoutSec int) {
	i := timeoutSec
	for {
		log.Println(i)
		time.Sleep(1 * time.Second)
		i--
		if i <= 0 {
			break
		}
	}
}

func inputChannelWithTimerMain(timeoutSec int) {
	userIn := make(chan string, 1)
	defer close(userIn)

	go inputChannel(userIn)
	for {
		log.Println("start.")
		go countDown(timeoutSec) // 入力待ちに関係ない、ただ一定時間定期的に出力する独立したもの。

		select {
		case x := <-userIn: // non block (because select block)
			log.Println(x)
			log.Println("this time, user input.")
		case <-time.After(time.Duration(timeoutSec) * time.Second):
			log.Println("this time, timed out.")
		}
	}
}

func inputChannelMain() {
	userIn := make(chan string, 1)
	defer close(userIn)

	go inputChannel(userIn)

	for {
		x := <-userIn // blocking
		log.Println(x)
	}
}

func simpleLineInputWait() {
	for {
		in := bufio.NewReader(os.Stdin)
		data, err := in.ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// fmt.Println(data[:len(data)-1])
		fmt.Println(strings.TrimRight(data, "\r\n"))
	}
}

func simpleLineInputWaitByScanner() {

	// 最新の１行だけ読みたい（２行以上読みたくない）場合にはこれでよいかも
	in := bufio.NewScanner(os.Stdin)
	if in.Scan() {
		data := in.Text()
		fmt.Println(data)
	}
	if err := in.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// or basic infinite loop (ファイル全行読み込みだとこっちが楽かもしれない)
	in2 := bufio.NewScanner(os.Stdin)
	for in2.Scan() {
		// loop until EOF.
		data := in2.Text()
		fmt.Println(data)
	}
	if err := in2.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func chanCapacitySampleProducer(store chan string, numOfInputAtATime int, inDuration time.Duration) {

	count := 1
	for {
		for i := 0; i < numOfInputAtATime; i++ {
			if cap(store) == len(store) {
				log.Println("[INFO] chan is full. input will be blocked until next output.")
			}
			store <- fmt.Sprintf("producer put text %04d times.", count)
			log.Printf("put -> (store) [%d of %d],  buf( %d / %d )\n", i+1, numOfInputAtATime, len(store), cap(store))
			count++
		}

		// wait before every produce
		log.Println("      [PRODUCER] sleeping")
		time.Sleep(inDuration)

		if count > 999 {
			log.Println("[INFO] produce counter reset to 1.")
			count = 1
		}
	}
}

func chanCapacitySampleSubscriber(store chan string, outDuration time.Duration) {
	for {
		if len(store) == 0 {
			log.Println("[INFO] chan is empty. wait for next input")
		}
		x := <-store
		log.Printf("get <- (store), buf( %d / %d )\n", len(store), cap(store))
		log.Printf("    got '%s'\n", x)

		// wait before every output
		log.Println("      [SUBSCRIBER] sleeping")
		time.Sleep(outDuration)
	}
}

func chanCapacitySample(cap int, numOfInputAtATime int, inDuration time.Duration, outDuration time.Duration) {

	store := make(chan string, cap)
	defer close(store)

	go chanCapacitySampleProducer(store, numOfInputAtATime, inDuration)
	go chanCapacitySampleSubscriber(store, outDuration)

	for {
		time.Sleep(1 * time.Second)
	}
}

func chanCapacitySampleInputFasterThanOutput() {
	chanCapacitySample(5, 2, 2*time.Second, 2*time.Second)
}

func chanCapacitySampleOutputFasterThanInput() {
	chanCapacitySample(5, 2, 6*time.Second, 2000*time.Millisecond)
}

func main() {

	var opt string
	if len(os.Args) == 2 {
		opt = os.Args[1]
	}
	switch opt {
	case "simple":
		simpleLineInputWait()
	case "inc":
		inputChannelMain()
	case "incsec":
		inputChannelWithTimerMain(5)
	case "chanin":
		chanCapacitySampleInputFasterThanOutput()
	case "chanout":
		chanCapacitySampleOutputFasterThanInput()
	default:
		inputChannelWithTimerMain(5)
	}

}
