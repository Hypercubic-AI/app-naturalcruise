// Command factor-decision-coverage-smoke drives the unmodified factor MCP
// binary over stdio and proves it consumes Natural decision sidecars.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

type response struct {
	ID     int `json:"id"`
	Result struct {
		Structured struct {
			TotalDecisions int `json:"totalDecisions"`
		} `json:"structuredContent"`
	} `json:"result"`
	Error any `json:"error"`
}

type manifest struct {
	Decisions []json.RawMessage `json:"decisions"`
}

func main() {
	binary := flag.String("binary", "", "factor-mcp binary built from the pinned factor checkout")
	manifestPath := flag.String("manifest", "", "Natural .decisions.json path")
	deadPath := flag.String("dead-branches", "", "Natural .dead-branches.json path")
	program := flag.String("program", "NCINMAPP", "Natural program name")
	factorSHA := flag.String("factor-sha", "", "pinned factor revision reported as evidence")
	flag.Parse()
	if *binary == "" || *manifestPath == "" || *deadPath == "" || *factorSHA == "" {
		fmt.Fprintln(os.Stderr, "binary, manifest, dead-branches, and factor-sha are required")
		os.Exit(2)
	}

	raw, err := os.ReadFile(*manifestPath)
	check(err)
	var sidecar manifest
	check(json.Unmarshal(raw, &sidecar))
	if len(sidecar.Decisions) == 0 {
		fail("Natural decision manifest is empty")
	}

	cmd := exec.Command(*binary)
	stdin, err := cmd.StdinPipe()
	check(err)
	stdout, err := cmd.StdoutPipe()
	check(err)
	cmd.Stderr = os.Stderr
	check(cmd.Start())
	defer func() {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	scanner := bufio.NewScanner(stdout)
	// factor announces tools/list_changed before initialization.
	if !scanner.Scan() {
		fail("factor MCP closed before its startup notification")
	}
	send(stdin, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]string{"name": "natural-decision-smoke", "version": "1"},
		},
	})
	init := receive(scanner)
	if init.ID != 1 || init.Error != nil {
		fail(fmt.Sprintf("factor MCP initialization failed: %+v", init))
	}
	send(stdin, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized", "params": map[string]any{}})
	send(stdin, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "tools/call",
		"params": map[string]any{
			"name": "decision_coverage",
			"arguments": map[string]any{
				"manifestPath":     *manifestPath,
				"deadBranchesPath": *deadPath,
				"program":          *program,
				"traces":           []any{},
			},
		},
	})
	coverage := receive(scanner)
	if coverage.ID != 2 || coverage.Error != nil {
		fail(fmt.Sprintf("factor decision_coverage failed: %+v", coverage))
	}
	if coverage.Result.Structured.TotalDecisions != len(sidecar.Decisions) {
		fail(fmt.Sprintf("factor rendered %d decisions; Natural manifest contains %d", coverage.Result.Structured.TotalDecisions, len(sidecar.Decisions)))
	}

	evidence := map[string]any{
		"consumer":       "factor-mcp@" + *factorSHA,
		"tool":           "decision_coverage",
		"program":        *program,
		"totalDecisions": coverage.Result.Structured.TotalDecisions,
		"manifestCount":  len(sidecar.Decisions),
	}
	encoded, _ := json.Marshal(evidence)
	fmt.Println(string(encoded))
}

func send(stdin interface{ Write([]byte) (int, error) }, value any) {
	raw, err := json.Marshal(value)
	check(err)
	raw = append(raw, '\n')
	_, err = stdin.Write(raw)
	check(err)
}

func receive(scanner *bufio.Scanner) response {
	if !scanner.Scan() {
		message := "factor MCP closed before replying"
		if err := scanner.Err(); err != nil {
			message += ": " + err.Error()
		}
		fail(message)
	}
	var reply response
	check(json.Unmarshal(scanner.Bytes(), &reply))
	return reply
}

func check(err error) {
	if err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
