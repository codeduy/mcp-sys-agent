package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"mcp-sys-agent/internal/dlp"
	"mcp-sys-agent/internal/executor"
	"mcp-sys-agent/internal/mcp"
)

func main() {
	fmt.Fprintf(os.Stderr, "=== Initializing mcp-sys-agent (Ultimate-DLP-6.0-Hybrid) ===\n")

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		var req mcp.RPCMessage
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		if req.ID == nil {
			continue
		}

		switch req.Method {
		case "initialize":
			mcp.SendResponse(req.ID, map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "DevOps-Autonomous-Agent",
					"version": "Ultimate-DLP-6.0-Hybrid",
				},
				"instructions": `You are a professional DevOps Engineer and Linux System Administrator. Your primary mission is to audit and troubleshoot the Server.
                    SUPREME RULES:
                    1. COMMUNICATION: 100% in English.
                    2. PRIVILEGES: You are operating in a Strict READ-ONLY environment. Any commands that alter system state (rm, systemctl, chmod, etc.) are strictly blocked. Do not attempt to bypass these restrictions.
                    3. API QUERIES: If you need to call a local API (e.g., Prometheus/Grafana), ONLY use the 'curl' command directed at localhost or 127.0.0.1 via HTTP GET (Pass parameters directly in the URL; the use of flags like -d, --data, -X POST, or saving to files is absolutely prohibited).
                    4. SEARCH EFFICIENCY: NEVER use the 'cat' command to read large configuration or log files. Prioritize 'tail -n 50', 'head', 'grep -i', 'find', or 'journalctl -n 50' to filter information precisely.
                    5. HANDLING DLP (Data Loss Prevention): If the returned output contains terms like '[BLINDED]', '[ANOMALOUS_STRING]', or '[REDACTED]', automatically assume the local DLP system has masked sensitive data. DO NOT attempt to find alternative commands to re-read those files.
                    6. SELF-CENSORSHIP: Before executing any command, self-evaluate. NEVER execute commands that carry the risk of damaging the system or aim to extract/leak sensitive content (e.g., passwords, tokens). Automatically discard such approaches and find a safer diagnostic method.
                    7. HANDLING FALSE POSITIVES: Because the DLP system is highly aggressive, harmless data might occasionally be masked. If the redacted data is **ABSOLUTELY NECESSARY** for your diagnosis, STOP AND ASK THE USER: "I need the redacted value in field [X], could you manually copy and paste it here?". Never attempt to bypass the rules yourself.
                    8. AGENT OPTIMIZATION FEEDBACK LOOP: During your analysis, if you notice the Agent's filter logic (in main.go) has flaws that hinder debugging, OR if you discover a security vulnerability that could leak sensitive data, PROACTIVELY SUGGEST an optimized solution or code fix. The user will review and update main.go accordingly.
					9. FAIL-FAST ON PERMISSION DENIED: If you encounter access restrictions (e.g., 'Permission denied') or realize you lack the necessary ACLs/permissions to read a critical file/directory, DO NOT waste time trying alternative commands, workarounds, or 'sudo' to bypass it. Immediately HALT your execution loop and explicitly tell the user: "I need ACL/read access to [File/Directory Path] to proceed." This is absolute to avoid infinite retry loops and hitting the 60-second execution timeout.
					10. AUTONOMOUS DEBUGGING LOOP (ReAct Framework): You are strictly forbidden from guessing the root cause or giving a final answer after just one observation. You MUST operate using an autonomous multi-step investigation loop. For every troubleshooting task, output your process clearly:
					- THOUGHT: State your hypothesis based on current information and explain what you need to check next.
					- ACTION: Execute the relevant read-only command via your tools.
					- (Wait for tool execution and OBSERVATION of the result)
					Repeat this [THOUGHT -> ACTION -> OBSERVATION] loop continuously. Dig deeper into logs, check dependent services, verify network ports, or inspect config files. ONLY BREAK THE LOOP and provide a "FINAL DIAGNOSIS" to the user when you have found the definitive Root Cause of the issue.
					11. KNOWLEDGE GROUNDING: When local diagnostics (Bash) confirm a specific error but no obvious solution is found in config files, you MUST use 'search_technical_knowledge' to look for community solutions on StackOverflow, GitHub, or Reddit. Prioritize recent discussions (last 2 years) to ensure compatibility with modern Linux environments.
					12. TIME-BOXED EXECUTION (60s LIMIT):
					- You have a 60-second limit per execution cycle. Do not attempt to solve everything in one single complex command.
					- STRATEGY: Breakdown tasks into smaller, atomic steps (e.g., Step 1: List files; Step 2: Grep logs; Step 3: Analyze config).
					- PROGRESS CHECKPOINTING: If you realize a task is large (e.g., analyzing 1GB of logs), explicitly state: "I am processing the first part of the logs, I will continue in the next step."
					- FAIL-SAFE: If a command is about to timeout, ensure your last Observation summarizes what you have found so far. This allows you to "resume" the diagnostic flow in the next interaction without losing context.
					- FEEDBACK: If you feel the 60s limit is consistently too short for a specific type of task, proactively suggest a more efficient Bash one-liner or a better diagnostic approach to the user.
					13. RATE LIMIT & 429 AVOIDANCE (ANTI-SPAM GUARD):
                    - MAXIMIZE EFFICIENCY: NEVER spam rapid, trivial tool calls. Maximize information gathered per execution by intelligently chaining bash commands. Avoid getting stuck in infinite failure/retry loops.
                    - SEARCH CONSERVATION: Do NOT use 'search_technical_knowledge' for local environment, localhost, or VS Code/plugin connection issues.
                    - 429 FALLBACK PROTOCOL: If any tool returns 'HTTP 429' or 'Too Many Requests', IMMEDIATELY STOP using that tool. Do NOT retry or rephrase the query. Fall back entirely to local Bash diagnostics or explicitly ask the user for instructions.`,
				"capabilities": map[string]interface{}{"tools": map[string]interface{}{}},
			}, nil)

		case "tools/list":
			mcp.SendResponse(req.ID, map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "run_bash_command",
						"description": "Execute a READ-ONLY bash command on the server. Destructive commands are blocked and sensitive data is automatically redacted.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"command": map[string]interface{}{
									"type":        "string",
									"description": "The bash command to run. Example: 'journalctl -xe | tail -n 50' or 'free -m'.",
								},
							},
							"required": []string{"command"},
						},
					},
					{
						"name": "search_technical_knowledge",
						"description": "Search the entire internet for technical solutions, expert blogs, official documentation, and community discussions (StackOverflow, GitHub, Reddit, Medium, Dev.to, etc.). Use this tool to find bug fixes, configuration examples, and proven workarounds when local diagnostics confirm a specific system issue or error code.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type": "string",
									"description": "The technical search query. Focus on error codes, software versions, and specific symptoms for better results from documentation or technical blogs.",
								},
							},
							"required": []string{"query"},
						},
					},
				},
			}, nil)

		case "tools/call":
			var params struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			json.Unmarshal(req.Params, &params)

			var resultText string
			var isError bool

			if params.Name == "run_bash_command" {
				cmdStr := fmt.Sprint(params.Arguments["command"])
				cmdLower := strings.ToLower(cmdStr)

				// Fetch Local LLM Environment Variables Early
				enableLocalLLM := os.Getenv("ENABLE_LOCAL_LLM")
				localEndpoint := os.Getenv("LOCAL_LLM_ENDPOINT")
				if localEndpoint == "" {
					localEndpoint = "http://localhost:11434"
				}

				// ==========================================
				// LAYER 1: STRICT FILTER (COMMAND BLACKLIST)
				// ==========================================
				if executor.IsBlocked(cmdStr) {
					resultText = fmt.Sprintf("⛔ WARNING: Command [%s] is BLOCKED! Agent operates in READ-ONLY mode.", cmdStr)
					mcp.SendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": resultText}},
						"isError": true,
					}, nil)
					continue
				}

				// ==========================================
				// LAYER 1.5: API COMMUNICATION CONTROL
				// ==========================================
				if strings.Contains(cmdLower, "curl ") || strings.Contains(cmdLower, "wget ") {
					if blocked, msg := executor.IsCurlBlocked(cmdStr); blocked {
						mcp.SendResponse(req.ID, map[string]interface{}{
							"content": []map[string]interface{}{{"type": "text", "text": msg}},
							"isError": true,
						}, nil)
						continue
					}
				}

				// ==========================================
				// LAYER 2: CONTEXTUAL BLINDNESS (CONTEXT DLP)
				// ==========================================
				if isBlinded, blindText := dlp.CheckContextualBlindness(cmdStr); isBlinded {
					mcp.SendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": blindText}},
						"isError": false,
					}, nil)
					continue
				}

				// ==========================================
				// LAYER 3: COMMAND EXECUTION (60S TIMEOUT)
				// ==========================================
				result := executor.RunBash(cmdStr)

				if result.Timeout {
					resultText = fmt.Sprintf("⛔ TIMEOUT ERROR (Exceeded 60s): Command [%s] was automatically terminated.\n"+
						"Possible causes: command hang or massive log file. Please use 'tail' or limit the time scope.", cmdStr)
					mcp.SendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": resultText}},
						"isError": true,
					}, nil)
					continue
				}

				rawOutput := result.Output
				var jsonParsed interface{}
				if errParse := json.Unmarshal([]byte(rawOutput), &jsonParsed); errParse == nil {
					if prettyBytes, errIndent := json.MarshalIndent(jsonParsed, "", "  "); errIndent == nil {
						rawOutput = string(prettyBytes)
					}
				}

				// ==========================================
				// LAYER 4: STATIC FILTERING (REGEX + ENTROPY)
				// ==========================================
				cleanOutput := dlp.StaticFilter(rawOutput)

				// ==========================================
				// LAYER 5: AI OUTPUT GUARDRAIL (DLP FILTERING)
				// ==========================================
				if enableLocalLLM == "1" {
					fmt.Fprintf(os.Stderr, "[OUTPUT GUARDRAIL] Local AI DLP activated. Checking endpoint: %s\n", localEndpoint)

					if !dlp.IsOllamaHealthy(localEndpoint) {
						errMsg := fmt.Sprintf("⛔ ZERO-TRUST ERROR: ENABLE_LOCAL_LLM=1 requires an active AI guard. Local AI (%s) is DOWN or unreachable. Execution ABORTED to prevent data leakage.", localEndpoint)
						mcp.SendResponse(req.ID, map[string]interface{}{
							"content": []map[string]interface{}{{"type": "text", "text": errMsg}},
							"isError": true,
						}, nil)
						continue
					}

					fmt.Fprintf(os.Stderr, "[OUTPUT GUARDRAIL] Local AI is responding. Processing text stream...\n")
					aiFilteredOutput, aiErr := dlp.FilterWithLocalLLM(localEndpoint, cleanOutput)

					if aiErr != nil {
						errMsg := fmt.Sprintf("⛔ ZERO-TRUST ERROR: Local AI encountered an error during DLP processing (%v). Execution ABORTED.", aiErr)
						mcp.SendResponse(req.ID, map[string]interface{}{
							"content": []map[string]interface{}{{"type": "text", "text": errMsg}},
							"isError": true,
						}, nil)
						continue
					}

					cleanOutput = aiFilteredOutput
					fmt.Fprintf(os.Stderr, "[OUTPUT GUARDRAIL] Processing successful! Data secured.\n")
				} else {
					fmt.Fprintf(os.Stderr, "[DLP] Bypassing Local AI (ENABLE_LOCAL_LLM=%s). Using static Regex/Entropy only.\n", enableLocalLLM)
				}

				// Process final return result
				if result.Err != nil {
					resultText = fmt.Sprintf("⚠️ ERROR EXECUTING COMMAND [%s]: %v\n---\n%s", cmdStr, result.Err, cleanOutput)
					isError = true
				} else {
					if len(cleanOutput) == 0 {
						resultText = fmt.Sprintf("✅ Command [%s] executed successfully (No output returned).", cmdStr)
					} else {
						resultText = fmt.Sprintf("Result of command [%s]:\n---\n%s", cmdStr, cleanOutput)
					}

					footer := "\n\n[SYSTEM REMINDER: You are in READ-ONLY mode. Strictly adhere to Rule 6 (Self-Censorship) and Rule 7 (Ask User if critical data is redacted).]"
					resultText += footer
				}

			} else if params.Name == "search_technical_knowledge" {
				query := fmt.Sprint(params.Arguments["query"])
				if query == "" || query == "<nil>" {
					resultText = "Search Error: 'query' argument is required and cannot be empty."
					isError = true
				} else {
					res, err := executor.SearchTechnicalKnowledge(query)
					if err != nil {
						resultText = fmt.Sprintf("Search Error: %v", err)
						isError = true
					} else {
						resultText = res
					}
				}
			} else {
				resultText = "Tool does not exist!"
				isError = true
			}

			mcp.SendResponse(req.ID, map[string]interface{}{
				"content": []map[string]interface{}{{"type": "text", "text": resultText}},
				"isError": isError,
			}, nil)
		}
	}
}

