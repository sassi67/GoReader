package lib

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const cTestDir string = "tmpDir"
const cFileContent string = `
Go by Example: Worker Pools

In this example we’ll look at how to implement a worker pool using goroutines and channels.
Here’s the worker, of which we’ll run several concurrent instances.
These workers will receive work on the jobs channel and send the corresponding results on results. 
We’ll sleep a second per job to simulate an expensive task.
In order to use our pool of workers we need to send them work and collect their results.
We make 2 channels for this.
This starts up 3 workers, initially blocked because there are no jobs yet.
Here we send 5 jobs and then close that channel to indicate that’s all the work we have.
Finally we collect all the results of the work.
This also ensures that the worker goroutines have finished.
An alternative way to wait for multiple goroutines is to use a WaitGroup.
Our running program shows the 5 jobs being executed by various workers.
The program only takes about 2 seconds despite doing about 5 seconds of total work because there are 3 workers operating concurrently.
`

func clearTestDir() {
	d, err := os.Open(cTestDir)
	if err != nil {
		panic(err)
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(cTestDir, name))
		if err != nil {
			panic(err)
		}
	}
}

func beforeClass(t *testing.T) func(t *testing.T) {
	t.Logf("Setup test case")
	if _, err := os.Stat(cTestDir); os.IsNotExist(err) {
		if os.Mkdir(cTestDir, 0777) != nil {
			panic(err)
		}
	} else {
		clearTestDir()
	}
	// after class...
	return func(t *testing.T) {
		t.Logf("Teardown test case")
		clearTestDir()
		_ = os.Remove(cTestDir)
	}
}

func beforeEach(t *testing.T, num int) func(t *testing.T, num int) {
	t.Logf("Setup test case with %d files", num)

	f, err := os.OpenFile(cTestDir+"/f"+strconv.Itoa(num),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for j := 1; j <= num; j++ {
		if _, err := f.WriteString(cFileContent); err != nil {
			panic(err)
		}
	}

	// after each...
	return func(t *testing.T, num int) {
		t.Logf("Teardown test case with %d files", num)
	}
}

func TestFileReader(t *testing.T) {

	var ch chan uint64
	var done chan struct{}
	var sum uint64
	wg := sync.WaitGroup{}

	counterFunc := func(count int) {
		defer wg.Done()
		counterDone := 0
		for {
			select {
			case <-done:
				counterDone++
				if counterDone == count {
					close(ch)
					return
				}
			case v := <-ch:
				sum = sum + v
			}
		}
	}

	expected := 0
	tearDownTestCase := beforeClass(t)
	defer tearDownTestCase(t)

	for i := 1; i <= 50; i++ {
		expected += i
		// init the test
		ch = make(chan uint64)
		done = make(chan struct{})
		sum = 0
		teardownSubTestCase := beforeEach(t, i)
		defer teardownSubTestCase(t, i)
		// start the collector
		wg.Add(1)
		go counterFunc(i)
		// run a FileReader for each file in parallel
		for j := 1; j <= i; j++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				filReader := NewFileReader(cTestDir+"/f"+strconv.Itoa(index), ch)
				assert.NotNil(t, filReader)
				filReader.StringSearch("Go by Example: Worker Pools")
				done <- struct{}{}
			}(j)
		}

		wg.Wait()
		assert.Equal(t, uint64(expected), sum)
	}
}
