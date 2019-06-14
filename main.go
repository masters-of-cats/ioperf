package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

func main() {
	operation := os.Args[1]

	switch operation {
	case "w":
		size, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
		count, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
		avg, err := runWriteTest(size, count)
		if err != nil {
			panic(err)
		}
		speed := float64(size/1024) / avg.Seconds()
		fmt.Printf("Average write: %s, speed: %f K/s\n", avg.String(), speed)
	}

	fmt.Println("DONE")
}

func runWriteTest(size, count int) (time.Duration, error) {
	random := rand.New(rand.NewSource(42))

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
		_, err = io.CopyN(file, random, int64(size))
		if err != nil {
			return 0, err
		}
		return time.Since(startTime), nil
	}, count)
}

func average(testFunc func() (time.Duration, error), count int) (time.Duration, error) {
	var totalRuntime time.Duration
	for i := 0; i < count; i++ {
		runTime, err := testFunc()
		if err != nil {
			return 0, err
		}
		totalRuntime += runTime
		fmt.Printf("Run %d: %s\n", i, runTime.String())
	}

	return time.Duration(int64(totalRuntime) / int64(count)), nil
}
