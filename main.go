package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") {
			fmt.Println("The program has completed its work")
			os.Exit(0)
		}
		values := strings.Fields(line)
		for _, val := range values {
			num, err := strconv.Atoi(val)
			if err != nil {
				fmt.Printf("Invalid number: %s (You can only use integers)\n", val)
				continue
			}
			select {
			case <-done:
				return
			case inputChan <- num:
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
				if x >= 0 {
					select {
					case <-done:
						return
					case withoutNegativeChan <- x:
					}
				}
			}
		}
	}()
	return withoutNegativeChan
}

// This filter stage excludes 0 and numbers not divisible by 3
func filterStage2(done <-chan struct{}, input <-chan int) <-chan int {
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
				if x > 0 && x%3 == 0 {
					select {
					case <-done:
						return
					case outputChan <- x:
					}
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

			}
		}
	}
}
func main() {
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
