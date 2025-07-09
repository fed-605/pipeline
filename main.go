package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Create logger to log our actioins with date and time to console
var logger = log.New(os.Stdout, "ACTION: ", log.Ldate|log.Ltime)

// Log our actions with message
func logAction(message string) {
	logger.Printf("[%s] %s", time.Now().Format("2006-01-02 15:04:05.000"), message)
}

// Buffer size
const MybufferSize int = 10

// buffer cleaning interval
const cleaningInterval = 30 * time.Second

// MyBuffer - ring buffer of integers
type MyBuffer struct {
	mu       sync.Mutex
	array    []int
	position int
	size     int
}

// Initializing a new buffer of integers
func bufferInitialization(size int) *MyBuffer {
	logAction("Initializing a new buffer of integers")
	return &MyBuffer{
		mu:       sync.Mutex{},
		array:    make([]int, size),
		position: -1,
		size:     size,
	}
}

// Push the elements in buffer
func (b *MyBuffer) Push(el int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.position == b.size-1 {
		for i := 1; i <= b.size-1; i++ {
			b.array[i-1] = b.array[i]
		}
		b.array[b.position] = el
	} else {
		b.position++
		b.array[b.position] = el
	}
}

// Get elemensts from buffer and than cleans it up
func (b *MyBuffer) Get() []int {
	if b.position < 0 {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	var output []int = b.array[:b.position+1]
	b.position = -1
	return output
}

// This function reads values from the console
// Write "exit" if you want to finish the work with pipeline
func reading(inputChan chan<- int, done <-chan struct{}) {
	logAction("The programm is ready to read from Stdin")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") {
			fmt.Println("The program has completed its work")
			logAction("The first filter has finished its work")
			logAction("The second filter has finished its work")
			logAction("The programm finished reading Stdin")
			logAction("The pipeline has shut down")
			os.Exit(0)
		}
		values := strings.Fields(line)
		for _, val := range values {
			num, err := strconv.Atoi(val)
			if err != nil {
				fmt.Printf("Invalid number: %s (You can only use integers)\n", val)
				logAction("The user entered incorrect data")
				continue
			}
			select {
			case <-done:
				return
			case inputChan <- num:
				logAction("the number was successfully read")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
	close(inputChan)
}

// This filter stage excludes numbers less than 0
func filterStage1(done <-chan struct{}, input <-chan int) <-chan int {
	logAction("The first filter has started its work")
	withoutNegativeChan := make(chan int)
	go func() {
		defer close(withoutNegativeChan)
		for {
			select {
			case <-done:
				return
			case x, ok := <-input:
				if !ok {
					return
				}
				logAction("the number went to the first stage of filtering")
				if x >= 0 {
					select {
					case <-done:
						return
					case withoutNegativeChan <- x:
						logAction("The number has passed the first filter stage")
					}
				} else {
					logAction("The number hasn't passed the first filter stage")
				}
			}
		}
	}()
	return withoutNegativeChan
}

// This filter stage excludes 0 and numbers not divisible by 3
func filterStage2(done <-chan struct{}, input <-chan int) <-chan int {
	logAction("The second filter has started its work")
	outputChan := make(chan int)
	go func() {
		defer close(outputChan)
		for {
			select {
			case <-done:
				return
			case x, ok := <-input:
				if !ok {
					return
				}
				logAction("the number went to the second stage of filtering")
				if x > 0 && x%3 == 0 {
					select {
					case <-done:
						return
					case outputChan <- x:
						logAction("The number has passed the second filter stage")
					}
				} else {
					logAction("The number hasn't passed the second filter stage")
				}
			}
		}
	}()
	return outputChan
}

// a function for transferring data to the buffer
func writeToBuffer(input <-chan int, b *MyBuffer) {
	for num := range input {
		b.Push(num)
		fmt.Printf("Data received: %d\n", num)
		logAction("The number was passed all filter stages")
		logAction("The number was recorded in the buffer")
	}
}

// a function for outputting data to the console at a certain interval
func writeToConsole(done <-chan struct{}, b *MyBuffer, ticker *time.Ticker) {
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			data := b.Get()
			if data != nil {
				fmt.Println("processed data:", data)
				logAction("The data from the buffer is printed to the console")
				logAction("The buffer has been cleared")
			}
		}
	}
}

// !!! The end of work with the pipeline is carried out only by the "exit" command
func main() {
	logAction("The pipeline has started working ")
	done := make(chan struct{})
	defer close(done)
	buffer := bufferInitialization(MybufferSize)
	input := make(chan int)
	go reading(input, done)
	pipeline := filterStage2(done, filterStage1(done, input))
	ticker := time.NewTicker(cleaningInterval)
	defer ticker.Stop()
	go writeToConsole(done, buffer, ticker)
	writeToBuffer(pipeline, buffer)

}
