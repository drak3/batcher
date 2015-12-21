package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

var command = flag.String("command", "", "The command to use")
var tsc0Start = flag.Int("start", 0, "Starting TSC0")
var tsc0End = flag.Int("end", 255, "End TSC0 (inclusive)")
var numKeyCandidates = flag.Int("num_candidates", 0, "Number of candidates to generate")
var template = flag.String("template", "", "Package template file")
var capFile = flag.String("cap_file", "", "Capture file prefix, assumption: file for $tsc0 is at ($cap_file)_$tsc0")
var numProcs = flag.Int("procs", 4, "total number of processes to use")

type Result struct {
	success bool
	tsc0    int
	out     []byte
}

//XXX: have to rewrite generator program so that it gives useful exit code and/or message
func main() {
	flag.Parse()
	wait := make(chan bool, *numProcs)
	results := make(chan Result)

	for tsc0 := *tsc0Start; tsc0 <= *tsc0End; tsc0++ {
		go func(tsc0 int) {
			cmd := createCommand(*command, tsc0, *numKeyCandidates, *template, *capFile)
			wait <- true
			out, err := cmd.Output()
			if err != nil {
				results <- Result{success: false, tsc0: tsc0}
			} else {
				results <- Result{success: true, out: out, tsc0: tsc0}
			}
			<-wait
		}(tsc0)
	}

	fmt.Println("Total Failures:")

	failures := 0

	for tsc0 := *tsc0Start; tsc0 <= *tsc0End; tsc0++ {
		r := <-results
		if r.success {
			fmt.Printf("\nSUCCESS: \n %s\n", string(r.out))
			os.Exit(0)
		} else {
			failures++
			fmt.Printf("\r%d", failures)
		}
	}
	fmt.Println("")
}

func createCommand(name string, tsc0 int, numKeyCandidates int, template string, capFilePref string) *exec.Cmd {
	dataFile := fmt.Sprintf("%s_%d", tsc0)
	return exec.Command(name, strconv.Itoa(tsc0), strconv.Itoa(numKeyCandidates), dataFile, template)
}
