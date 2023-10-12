package mercury

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

/*
 * Execute an operation using the Mercury Client Adapter Node.js process.
 *
 * Usage:
 *		content, data, err := mercury.Request(operation, target)
 *
 * operation string:
 *		"request" makes a zCap-authorized HTTP request
 *		"download" uses zCaps to download and decrypt an EDV document
 * target string:
 *		URL of the HTTP API or EDV resource
 * content []byte:
 * 		Portion of sdtout buffer returned by Mercury Client Adapter.
 *		Includes log messages and EDV document `content` (where applicable)
 * data []byte:
 *		Portion of stdout buffer returned by Mercury Client Adapter.
 *		Includes HTTP Response or EDV stream data
 * err error:
 *		Exit code from os/exec operation
 *		Note: Mercury Client Node.js error information is output to the console
 */
func Request(operation string, target string) ([]byte, []byte, error) {
	// Declare variables
	var qcontent []byte
	var qdata []byte

	// Collect results from StdOut
	type output struct {
		result []byte
		stderr []byte
		err    error
	}
	ch := make(chan output)

	// Run Mercury Client Adapter Node.js process
	go func() {
		if operation == "download" {
			// Requires version 3.0.0 of the client adapter
			operation = "get -d"
		}
		cmd := exec.Command("mercury-client", operation, target)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		ch <- output{result: stdout.Bytes(), stderr: stderr.Bytes(), err: err}
	}()

	// Handle Timeouts and Errors
	select {
	case <-time.After(60 * time.Second):
		fmt.Println("Command timed out")
	case x := <-ch:
		if x.err != nil {
			// Catch errors from node.js
			fmt.Println("Error executing command:", string(x.stderr))
			return nil, nil, x.err
		}

		var qindex int
		var offset int
		switch operation {
		case "download":
			// Find the start of the Stream data (Uint8 array)
			qindex = findAllOccurrences(x.result, []string{`EDV Stream:`})[`EDV Stream:`][0]
			offset = len([]byte(`EDV Stream:`)) + 1
		case "request":
			// Find the start of the Stream data (Uint8 array)
			qindex = findAllOccurrences(x.result, []string{`Response:`})[`Response:`][0]
			offset = len([]byte(`Response:`)) + 1
		default:
			qindex = 0
			offset = 0
		}

		// Preceding console logs and EDV Document `content``
		qcontent = x.result[0 : qindex-1]

		// HTTP Response OR EDV Document Stream data(Uint8 array)
		qdata = x.result[qindex+offset : len(x.result)-1]
	}
	return qcontent, qdata, nil
}

/*
 * Helper Function
 * Finds the location(s) of a particular string in binary UTF-8 data
 */
func findAllOccurrences(data []byte, searches []string) map[string][]int {
	results := make(map[string][]int)
	for _, search := range searches {
		searchData := data
		term := []byte(search)
		for x, d := bytes.Index(searchData, term), 0; x > -1; x, d = bytes.Index(searchData, term), d+x+1 {
			results[search] = append(results[search], x+d)
			searchData = searchData[x+1:]
		}
	}
	return results
}
