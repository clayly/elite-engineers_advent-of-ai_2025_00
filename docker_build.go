package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

func RunDockerBuild() {
	cmd := exec.Command(
		"docker",
		"build",
		"--progress=plain",
		"--no-cache",
		"-t",
		"mcp_docker_build_"+strconv.FormatUint(rand.Uint64(), 10),
		".",
		"--file",
		"Dockerfile",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start docker container: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Failed to wait for docker container: %v", err)
	}
}
