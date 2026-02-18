package utils

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
)

func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

func ReadLine(filename, prefix string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if strings.HasPrefix(line, prefix) {
			return line, nil
		}
	}

	return "", nil
}

func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer file.Close()

	var ret []string
	reader := bufio.NewReader(file)

	for i := uint(0); i < uint(n)+offset || n < 0; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && line != "" {
				ret = append(ret, strings.Trim(line, "\n"))
			}
			break
		}

		if i < offset {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}

// ReadLinesOffsetNWithContext reads lines with offset and count, respecting ctx cancellation.
func ReadLinesOffsetNWithContext(ctx context.Context, filename string, offset uint, n int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer file.Close()

	var ret []string
	reader := bufio.NewReader(file)

	for i := uint(0); i < uint(n)+offset || n < 0; i++ {
		if err := ctx.Err(); err != nil {
			return ret, err
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && line != "" {
				ret = append(ret, strings.Trim(line, "\n"))
			}
			break
		}

		if i < offset {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}
