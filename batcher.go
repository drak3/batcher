package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var command = flag.String("command", "main.exe", "The command to use")
var tsc0Start = flag.Int("start", 0, "Starting TSC0")
var tsc0End = flag.Int("end", 255, "End TSC0 (inclusive)")
var numKeyCandidates = flag.Int("num_candidates", 100, "Number of candidates to generate")
var template = flag.String("template", "sniff_data/template", "Package template file")
var capFile = flag.String("cap_file", "sniff_data/data_dump_", "Capture file prefix, assumption: file for $tsc0 is at ($cap_file)_$tsc0")
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

	fmt.Printf("Running command \"%s\" from tsc0=%d to tsc0=%d\n", *command, *tsc0Start, *tsc0End)
	fmt.Printf("Generating %d candidates\n", *numKeyCandidates)
	fmt.Printf("Utilizing %d processes at the same time\n", *numProcs)
	fmt.Printf("Reading from template '%s' and capFile '%s'\n", *template, *capFile)

	sampleCmd := createCommand(*tsc0Start)
	fmt.Printf("Sample invocation: \n\t %s \n", strings.Join(sampleCmd.Args, " "))

	for tsc0 := *tsc0Start; tsc0 <= *tsc0End; tsc0++ {
		go func(tsc0 int) {
			cmd := createCommand(tsc0)
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

func createCommand(tsc0 int) *exec.Cmd {
	dataFile := fmt.Sprintf("%s%d", *capFile, tsc0)
	return exec.Command(*command, strconv.Itoa(tsc0), strconv.Itoa(*numKeyCandidates), dataFile, *template)
}
