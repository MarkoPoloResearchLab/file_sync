package sync

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

type mergeInputs struct {
	BaseBytes  []byte
	SideABytes []byte
	SideBBytes []byte
	Diff3Path  string
}

func mergeThreeWay(inputs mergeInputs) ([]byte, bool) {
	if inputs.Diff3Path != "" {
		merged, ok := mergeWithDiff3(inputs.Diff3Path, inputs.BaseBytes, inputs.SideABytes, inputs.SideBBytes)
		if ok {
			return merged, true
		}
	}
	return mergeWithMarkers(inputs.SideABytes, inputs.SideBBytes), false
}

func mergeWithMarkers(sideA []byte, sideB []byte) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("<<<<<<< SIDE_A\n")
	buffer.Write(sideA)
	if len(sideA) > 0 && sideA[len(sideA)-1] != '\n' {
		buffer.WriteByte('\n')
	}
	buffer.WriteString("=======\n")
	buffer.Write(sideB)
	if len(sideB) > 0 && sideB[len(sideB)-1] != '\n' {
		buffer.WriteByte('\n')
	}
	buffer.WriteString(">>>>>>> SIDE_B\n")
	return buffer.Bytes()
}

func mergeWithDiff3(diff3Path string, base []byte, sideA []byte, sideB []byte) ([]byte, bool) {
	if diff3Path == "" {
		return nil, false
	}
	tempDir, tempErr := os.MkdirTemp("", "filez-sync-merge-*")
	if tempErr != nil {
		return nil, false
	}
	defer os.RemoveAll(tempDir)

	basePath := filepath.Join(tempDir, "base")
	aPath := filepath.Join(tempDir, "a")
	bPath := filepath.Join(tempDir, "b")

	if err := os.WriteFile(basePath, base, 0o644); err != nil {
		return nil, false
	}
	if err := os.WriteFile(aPath, sideA, 0o644); err != nil {
		return nil, false
	}
	if err := os.WriteFile(bPath, sideB, 0o644); err != nil {
		return nil, false
	}

	cmd := exec.Command(diff3Path, "-m", aPath, basePath, bPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return out, true
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return out, true
	}
	return nil, false
}
