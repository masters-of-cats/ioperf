package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

func main() {
	concurrentOps, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	operation := os.Args[2]

	switch operation {
	case "w":
		size, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
		count, err := strconv.Atoi(os.Args[4])
		if err != nil {
			panic(err)
		}
		avg, err := runWriteTest(concurrentOps, size, count)
		if err != nil {
			panic(err)
		}
		speed := float64(size/1024) / avg.Seconds()
		fmt.Printf("Average write: %s, speed: %f K/s\n", avg.String(), speed)
	}

	fmt.Println("DONE")
}

func runWriteTest(concurrentOps, size, count int) (time.Duration, error) {
	random := rand.New(rand.NewSource(42))
	data := make([]byte, size)
	_, err := random.Read(data)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Running write test with file size %d and file count %d\n\n", size, count)
	return average(func() (time.Duration, error) {
		u, err := uuid.NewV4()
		if err != nil {
			return 0, err
		}
		file, err := os.OpenFile("myfile-"+u.String(), os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return 0, err
		}
		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()

		startTime := time.Now()
		_, err = file.Write(data)
		if err != nil {
			return 0, err
		}
		return time.Since(startTime), nil
	}, concurrentOps, count)
}

func average(testFunc func() (time.Duration, error), concurrentOps, count int) (time.Duration, error) {
	durationChan := make(chan time.Duration)
	errChan := make(chan error)

	runWg := sync.WaitGroup{}
	runWg.Add(concurrentOps)

	for i := 0; i < concurrentOps; i++ {
		go func() {
			defer runWg.Done()

			for i := 0; i < count; i++ {
				runTime, err := testFunc()
				if err != nil {
					errChan <- err
					return
				}

				durationChan <- runTime
				fmt.Printf("%s\n", runTime.String())
			}
		}()
	}

	collectWg := sync.WaitGroup{}
	collectWg.Add(1)
	var totalRuntime time.Duration
	var operationErr error
	go func() {
		defer collectWg.Done()

		for {
			select {
			case d, ok := <-durationChan:
				if !ok {
					return
				}
				totalRuntime += d
			case err, ok := <-errChan:
				if !ok {
					return
				}
				operationErr = err
			}
		}
	}()

	runWg.Wait()
	close(durationChan)
	close(errChan)

	collectWg.Wait()

	return time.Duration(int64(totalRuntime) / (int64(count) * int64(concurrentOps))), operationErr
}
