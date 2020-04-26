package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
)

// Check and panic on error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Retrieve bytes from `filename`
func getBytes(filename string) []byte {
	f, err := ioutil.ReadFile(filename)
	check(err)
	return f
}

// Create new file with `data`
func createNew(data []byte) {
	err := ioutil.WriteFile("mutated.jpg", data, 0644)
	check(err)
}

// Randomly flip 1% of the bits in `data`
func mutateBits(data []byte) []byte {
	var byteIndexes []int
	numBytes := int(float64(len(data)-4) * .01)
	for len(byteIndexes) < numBytes {
		// For random ints in range = rand.Intn(Max - Min) + Min
		byteIndexes = append(byteIndexes, rand.Intn((len(data)-4)-4)+4)
	}
	//fmt.Println("Indexes chosen: ", byteIndexes)

	// Randomly change the bytes at the location of the chosen indexes
	for _, index := range byteIndexes {
		//oldbytes := data[index]
		newbytes := byte(rand.Intn(0xFF))
		data[index] = newbytes
		//fmt.Printf("Changed %x to %x\n", oldbytes, newbytes)
	}
	return data
}

// Mutate `data` in special ways to cause overflows, etc.
func mutateMagic(data []byte) []byte {
	// Gynvael's magic numbers https://www.youtube.com/watch?v=BrDujogxYSk&
	magicVals := [][]int{
		{1, 255},
		{1, 255},
		{1, 127},
		{1, 0},
		{2, 255},
		{2, 0},
		{4, 255},
		{4, 0},
		{4, 128},
		{4, 64},
		{4, 127},
	}

	pickedMagic := magicVals[rand.Intn(len(magicVals))]
	index := rand.Intn(len(data) - 8)

	// Hardcode byte overwrites for tuples beginning with (1, )
	if pickedMagic[0] == 1 {
		switch pickedMagic[1] {
		case 255:
			data[index] = 255
		case 127:
			data[index+1] = 127
		case 0:
			data[index] = 0
		}
		// Hardcode byte overwrites for tuples beginning with (2, )
	} else if pickedMagic[0] == 2 {
		switch pickedMagic[1] {
		case 255:
			data[index] = 255
			data[index+1] = 255
		case 0:
			data[index] = 0
			data[index+1] = 0
		}
		// Hardcode byte overwrites for tuples beginning with (4, )
	} else if pickedMagic[0] == 4 {
		switch pickedMagic[1] {
		case 255:
			data[index] = 255
			data[index+1] = 255
			data[index+2] = 255
			data[index+3] = 255
		case 0:
			data[index] = 0
			data[index+1] = 0
			data[index+2] = 0
			data[index+3] = 0
		case 128:
			data[index] = 128
			data[index+1] = 0
			data[index+2] = 0
			data[index+3] = 0
		case 64:
			data[index] = 64
			data[index+1] = 0
			data[index+2] = 0
			data[index+3] = 0
		case 127:
			data[index] = 127
			data[index+1] = 255
			data[index+2] = 255
			data[index+3] = 255
		}
	}
	return data
}

// Select mutator at random to mutate `data`
func mutate(data []byte) []byte {
	mutators := []func([]byte) []byte{mutateBits, mutateMagic}
	return mutators[rand.Intn(len(mutators))](data)
}

// Run exif program and keep track of segfaults
func exif(counter int, data []byte) {
	// Run command, capture output
	var output bytes.Buffer
	//exifCommand := "cmd.exe"
	//exifArgs := []string{"/c", ".\\exif\\bin\\exif_win32.exe", ".\\mutated.jpg", "-verbose"}
	exifCommand := ".\\exif\\bin\\exif_win32.exe"
	exifArgs := []string{".\\mutated.jpg", "-verbose"}
	cmd := exec.Command(exifCommand, exifArgs[0], exifArgs[1])
	cmd.Stderr = &output
	cmd.Stdout = &output
	err := cmd.Start()
	check(err)

	// Write any crashes to file
	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Println(exitError)
			if strings.Contains(exitError.String(), "3221225477") {
				err = ioutil.WriteFile(fmt.Sprintf(".\\crashes\\crash.%d.jpg", counter), data, 0644)
				check(err)
			}
		}
	}

	// Print counter as status updates
	if counter%100 == 0 {
		fmt.Println(counter)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go exif_fuzz.go <valid_jpg>")
		os.Exit(1)
	}

	// Create mutated file
	filename := os.Args[1]
	for counter := 0; counter < 100000; counter++ {
		data := getBytes(filename)
		mutated := mutate(data)
		createNew(mutated)
		exif(counter, mutated)
	}
}
