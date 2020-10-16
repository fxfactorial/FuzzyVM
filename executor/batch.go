// Copyright 2020 Marius van der Wijden
// This file is part of the fuzzy-vm library.
//
// The fuzzy-vm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The fuzzy-vm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the fuzzy-vm library. If not, see <http://www.gnu.org/licenses/>.

package executor

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/korovkin/limiter"
	"github.com/pkg/errors"
)

var (
	batchSize        = 4
	concurrencyLimit = 10
)

// Execute runs all tests in `dirName` and saves crashers in `outDir`
func ExecuteBatch(dirName, outDir string) error {
	infos, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err
	}
	errChan := make(chan error)
	limit := limiter.NewConcurrencyLimiter(concurrencyLimit)
	meter := metrics.GetOrRegisterMeter("ticks", nil)

	for i := 0; i < len(infos)/batchSize; i++ {
		var batch []string
		for k := 0; k < batchSize; k++ {
			name := infos[i].Name()
			// All generated tests end in .json
			if strings.HasSuffix(name, ".json") {
				batch = append(batch, name)
			}
		}
		fmt.Printf("Executing batch: %v of %v, %f per minute \n", i, len(infos)/batchSize, meter.Rate1())
		meter.Mark(1)
		job := func() {
			if err := ExecuteFullBatch(dirName, outDir, batch, true); err != nil {
				err := errors.Wrap(err, fmt.Sprintf("in batch %v:", i))
				fmt.Println(err)
			}
		}
		limit.Execute(job)
	}
	for {
		select {
		case err := <-errChan:
			fmt.Println(err)
		default:
			// All tests sucessfully executed
			return nil
		}
	}
}

func ExecuteFullBatch(dirName, outDir string, filenames []string, doPurge bool) error {
	var (
		testFiles  []string
		testNames  []string
		traceFiles []string
	)
	for _, f := range filenames {
		var (
			testFile  = fmt.Sprintf("%v/%v", dirName, f)
			testName  = strings.TrimRight(f, ".json")
			traceFile = fmt.Sprintf("%v/%v-trace.jsonl", dirName, testName)
		)
		testFiles = append(testFiles, testFile)
		testNames = append(testNames, testName)
		traceFiles = append(traceFiles, traceFile)
	}

	outputs, err := executeTestBatch(testFiles)
	if err != nil {
		return err
	}
	// The outputs are in a weird format, unpack them.
	// Iterate over tests
	for i := range outputs[0] {
		var batch [][]byte
		// Iterate over vms
		for k := range outputs {
			batch = append(batch, outputs[k][i])
		}
		if err := verifyAndPurge(traceFiles[i], testNames[i], outDir, testFiles[i], batch, true); err != nil {
			return err
		}
	}
	return nil
}

// executeTestBatch executes a batch of tests.
// It returns a [][][]byte
// In the first dimension we have the len(vms)
// In the second dimension we have len(tests)
// In the third dimension we have len(outputs)
func executeTestBatch(tests []string) ([][][]byte, error) {
	var buf [][][]byte
	for _, vm := range vms {
		b, err := vm.RunStateTestBatch(tests)
		if err != nil {
			return nil, err
		}
		buf = append(buf, b)
	}
	return buf, nil
}