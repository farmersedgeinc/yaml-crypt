package fixtures

import (
	"path/filepath"
	"runtime"
	"os"
	"errors"
	"io"
	"bytes"
)

var Strings = []string{
	"1ZQheycERGjpTeXgJrzpjmxDxaixVJVcstgKyiwUshRx7AwZAsHtRwGVFgDtQlyVRiMMw618Hr4kbty66v1NN7acAnApEBAz9BfwfQ5kz87aGcaPSeck3F8obdyRPHmA",
	"xQr3HS4TmD87DPLMU17gUicZ",
	"lWBsiwTmlpA0H5k2nZjz64UM",
	"test",
	"{\"message\": \"weird, but could happen\"}",
	"\n % cat /dev/urandom | strings -n 16\n3 ,V\";u=)gg	H(>{d\nhV9+pprKG	>|=)IyO\nPvico@rXJ{L/g&g\n b83x)))n5GU+oI*a)\n!__dL`;5IX]/)ro1Pa4\nh;{a3U8\\fRYI}07Ivr]\nCy%JJY]kBMl!)tm	\nWo:Y	@1|5<FPFF	t }\n",
	"🤔🤔🤔",
}

func TestDataDir() (string, error) {
	// get path to this source file
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("Unable to find test dir: runtime.Caller failed")
	}
	root, err := filepath.Abs(here)
	if err != nil {
		return "", err
	}
	for i := 0; i < 3; i++ {
		root = filepath.Dir(root)
	}
	return filepath.Join(root, "testdata"), nil
}

func cp(src, dst string) error {
	// open source
	srcHandle, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcHandle.Close()
	// open destination
	dstHandle, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstHandle.Close()

	// copy
	_, err = io.Copy(dstHandle, srcHandle)
	if err != nil {
		return err
	}
	return dstHandle.Sync()
}

func compareFiles(paths ...string) (bool, error) {
	var err error
	n := len(paths)
	files := make([]*os.File, n)
	for i, path := range paths {
		files[i], err = os.Open(path)
		defer files[i].Close()
		if err != nil {
			return false, err
		}
	}
	errs := make([]error, n)
	bufs := make([][]byte, n)
	for i := range bufs {
		bufs[i] = make([]byte, 1024)
	}
	for {
		// read a chunk from all files
		for i, file := range files {
			_, errs[i] = file.Read(bufs[i])
			if errs[i] != nil && errs[i] != io.EOF {
				return false, err
			}
		}
		// if some but not all files reached EOF, the files differ
		nEofs := 0
		for _, err := range errs {
			if err == io.EOF {
				nEofs += 1
			}
		}
		if !(nEofs == 0 || nEofs == n) {
			return false, nil
		}
		// if any buffers differ from the first, the files differ
		for _, buf := range bufs[1:] {
			if !bytes.Equal(buf, bufs[0]) {
				return false, nil
			}
		}
		// if at this point we have an EOF, we have finished comparing and found no differences
		if nEofs == n {
			return true, nil
		}
	}
}
