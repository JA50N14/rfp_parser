package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func (cfg *apiConfig) pdfParser(f *os.File, kpiResults []KpiResult) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	} 

	cmd := exec.Command("pdftotext", "-", "-")
	cmd.Stdin = f

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdoutPipe)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		scanTextWithRegex(line, kpiResults)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		stderrBytes, _ := io.ReadAll(stderrPipe)
		return fmt.Errorf("pdftotext failed: %w, stderr: %s", err, string(stderrBytes))
	}

	return nil
}