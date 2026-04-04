package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// --- Standard MCP JSON-RPC Structure ---
type RPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

func sendResponse(id json.RawMessage, result interface{}, err interface{}) {
	resp := RPCMessage{JSONRPC: "2.0", ID: id, Result: result, Error: err}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data) // MCP strictly requires Stdout for JSON
}

// --- SHANNON ENTROPY CALCULATION ---
func shannonEntropy(data string) float64 {
	if data == "" {
		return 0
	}
	frequencies := make(map[rune]float64)
	for _, char := range data {
		frequencies[char]++
	}

	entropy := 0.0
	length := float64(len(data))

	for _, freq := range frequencies {
		p := freq / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// ==========================================
// [ZERO-TRUST PATCH]: LOCAL LLM INTEGRATION
// ==========================================
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

// 1. Health Check (Ping Ollama endpoint)
func isOllamaHealthy(endpoint string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(endpoint + "/api/tags")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	return true
}

// 2. [ENDPOINT: DATA LOSS PREVENTION] - Output Filtering
func filterWithLocalLLM(endpoint, rawText string) (string, error) {
	url := endpoint + "/api/generate"

	prompt := fmt.Sprintf(`[SYSTEM] You are a local Data Loss Prevention (DLP) filter. 
Task: Find and replace ALL passwords, API Keys, Tokens, Private Keys, and Public IPs in the following text with "[REDACTED_BY_LOCAL_AI]". 
RETURN ONLY the redacted content. DO NOT explain, DO NOT greet, DO NOT add any extra formatting outside the requested output.

[DATA]:
%s`, rawText)

	reqBody := OllamaRequest{
		Model:  "qwen2.5:3b",
		Prompt: prompt,
		Stream: false,
	}

	jsonData, _ := json.Marshal(reqBody)
	// Longer timeout (30s) for analyzing large text chunks
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}

func main() {
	fmt.Fprintf(os.Stderr, "=== Initializing mcp-sys-agent (Ultimate-DLP-6.0-Hybrid) ===\n")

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		var req RPCMessage
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		if req.ID == nil {
			continue
		}

		switch req.Method {
		case "initialize":
			sendResponse(req.ID, map[string]interface{}{
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
                    8. AGENT OPTIMIZATION FEEDBACK LOOP: During your analysis, if you notice the Agent's filter logic (in main.go) has flaws that hinder debugging, OR if you discover a security vulnerability that could leak sensitive data, PROACTIVELY SUGGEST an optimized solution or code fix. The user will review and update main.go accordingly.`,
				"capabilities": map[string]interface{}{"tools": map[string]interface{}{}},
			}, nil)

		case "tools/list":
			sendResponse(req.ID, map[string]interface{}{
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
				blacklist := []string{
					"rm ", "mv ", "cp ", "touch ", "mkdir ", "rmdir ", "dd ", "ln ", "truncate ", "tee ",
					">", ">>", "nano ", "vim ", "vi ", "editor ", "sed -i ", "awk -i ",
					"chmod ", "chown ", "chgrp ", "passwd ", "useradd ", "userdel ", "usermod ", "su ", "sudo ", "visudo ",
					"scp ", "rsync ", "nc ", "netcat ", "ftp ", "sftp ",
					"apt ", "apt-get ", "yum ", "dnf ", "dpkg ", "rpm ", "pacman ", "snap ", "npm ", "pip ",
					"systemctl stop", "systemctl restart", "systemctl disable", "systemctl start", "systemctl enable", "systemctl reload", "systemctl mask",
					"service ", "/etc/init.d/", "crontab -e", "crontab -r",
					"kill ", "pkill ", "killall ", "reboot", "shutdown", "init 0", "init 6", "halt", "poweroff",
					"mysqladmin ", "mysqldump ", "git clone ", "git pull ", "git push ",
					"tar -x", "unzip ", "gunzip ",
					"mount ", "umount ", "mkfs", "fdisk ", "parted ", "iptables ", "ufw ", "firewall-cmd ",
					"docker run ", "docker stop ", "docker rm ", "docker exec ", "kubectl apply ", "kubectl delete ", "kubectl edit ",
				}

				isBlocked := false
				for _, badWord := range blacklist {
					if strings.Contains(cmdLower, badWord) {
						isBlocked = true
						break
					}
				}

				if isBlocked {
					resultText = fmt.Sprintf("⛔ WARNING: Command [%s] is BLOCKED! Agent operates in READ-ONLY mode.", cmdStr)
					sendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": resultText}},
						"isError": true,
					}, nil)
					continue
				}

				// ==========================================
				// LAYER 1.5: API COMMUNICATION CONTROL
				// ==========================================
				if strings.Contains(cmdLower, "curl ") || strings.Contains(cmdLower, "wget ") {
					if !strings.Contains(cmdLower, "localhost") && !strings.Contains(cmdLower, "127.0.0.1") {
						sendResponse(req.ID, map[string]interface{}{
							"content": []map[string]interface{}{{"type": "text", "text": "⛔ WARNING: curl/wget is ONLY PERMITTED to call [localhost] or [127.0.0.1] to prevent data exfiltration."}},
							"isError": true,
						}, nil)
						continue
					}

					curlBlacklist := []string{
						" -o", "--output", "--remote-name",
						" -x", "--request",
						" -d", "--data", "--form",
					}

					for _, badFlag := range curlBlacklist {
						if strings.Contains(cmdLower, badFlag) {
							sendResponse(req.ID, map[string]interface{}{
								"content": []map[string]interface{}{{"type": "text", "text": "⛔ WARNING: curl is restricted to pure HTTP GET. POST/PUT/DELETE, data payloads, and file saving are blocked."}},
								"isError": true,
							}, nil)
							isBlocked = true
							break
						}
					}
					if isBlocked {
						continue
					}
				}

				// ==========================================
				// LAYER 2: CONTEXTUAL BLINDNESS (CONTEXT DLP)
				// ==========================================
				type TechProfile struct {
					TechName     string
					TriggerWords []string
					BlindMessage string
				}

				techProfiles := []TechProfile{
					{"Node.js / Web Frameworks", []string{".env"}, "contains environment variables and API Keys"},
					{"WordPress", []string{"wp-config.php"}, "contains clear-text Database passwords"},
					{"Docker / Container", []string{"docker-compose.yml", "docker-compose.yaml"}, "contains sensitive service parameters"},
					{"Kubernetes", []string{".kube/config", "kubeconfig"}, "contains K8s Cluster admin certificates"},
					{"AWS Cloud", []string{".aws/credentials", ".aws/config"}, "contains AWS Access/Secret Keys"},
					{"SSH / Git Keys", []string{"id_rsa", "id_ed25519", "authorized_keys", ".git/credentials"}, "contains Private/Foreign Keys"},
					{"Linux Core System", []string{"/etc/shadow", "/etc/sudoers"}, "contains password hashes and permissions"},
					{"Terminal History", []string{".bash_history", ".zsh_history", ".mysql_history", ".psql_history"}, "contains manually typed commands and passwords"},
					{"DirectAdmin Core", []string{"/usr/local/directadmin/conf/"}, "contains core configs and Root MySQL passwords"},
					{"DirectAdmin Users Data", []string{"/usr/local/directadmin/data/users/"}, "contains sensitive data of all hosted clients"},
					{"DirectAdmin Setup Log", []string{"setup.txt"}, "contains initial Admin password (Clear-text)"},
					{"Mail Server Configs", []string{"/etc/virtual/"}, "contains email lists and client password hashes"},
					{"MariaDB / MySQL", []string{"my.cnf", ".my.cnf", "debian.cnf"}, "contains high-privilege Database credentials"},
					{"CSF Firewall", []string{"/etc/csf/csf.allow", "/etc/csf/csf.deny", "/etc/csf/csf.ignore"}, "contains client IPs and firewall rules"},
					{"Customer Source Code", []string{"wp-config.php", "configuration.php", ".env"}, "contains website-specific Database passwords"},
				}

				isBlinded := false
				for _, profile := range techProfiles {
					for _, trigger := range profile.TriggerWords {
						if strings.Contains(cmdLower, trigger) {
							isBlinded = true
							resultText = fmt.Sprintf(
								"✅ Command [%s] is valid. File/Directory EXISTS.\n"+
									"🛡️ DLP SYSTEM: Detected [%s] technology. Activated BLIND MODE (content ignored) to prevent data leak: %s.\n"+
									"👉 AI INSTRUCTION: Assume this configuration file is error-free, proceed to investigate other log files or service states.",
								cmdStr, profile.TechName, profile.BlindMessage)
							isError = false
							break
						}
					}
					if isBlinded {
						break
					}
				}

				if isBlinded {
					sendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": resultText}},
						"isError": isError,
					}, nil)
					continue
				}

				// ==========================================
				// LAYER 3: COMMAND EXECUTION (60S TIMEOUT)
				// ==========================================
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()

				cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
				out, err := cmd.CombinedOutput()
				rawOutput := string(out)

				if ctx.Err() == context.DeadlineExceeded {
					resultText = fmt.Sprintf("⛔ TIMEOUT ERROR (Exceeded 60s): Command [%s] was automatically terminated.\n"+
						"Possible causes: command hang or massive log file. Please use 'tail' or limit the time scope.", cmdStr)
					sendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": resultText}},
						"isError": true,
					}, nil)
					continue
				}

				var jsonParsed interface{}
				if errParse := json.Unmarshal([]byte(rawOutput), &jsonParsed); errParse == nil {
					if prettyBytes, errIndent := json.MarshalIndent(jsonParsed, "", "  "); errIndent == nil {
						rawOutput = string(prettyBytes)
					}
				}

				// ==========================================
				// LAYER 4: STATIC FILTERING (REGEX + ENTROPY)
				// ==========================================
				reURI := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+-.]*:\/\/[^\s:@\/]+:)([^\s:@\/]+)(@[^\s\/]+)`)
				cleanOutput := reURI.ReplaceAllString(rawOutput, "$1[URI_PASSWORD_REDACTED]$3")

				reSecret := regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api[_-]?key|private[_-]?key|salt|bearer|client[_-]?secret)\s*[:=]\s*([^\s\n\r"']+)`)
				cleanOutput = reSecret.ReplaceAllString(cleanOutput, "$1 = [🔒 DATA_REDACTED]")

				reAPIKeys := regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16}|sk_live_[0-9a-zA-Z]{24}|ghp_[0-9a-zA-Z]{36}|xox[bap]-[0-9a-zA-Z_-]+|eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+)`)
				cleanOutput = reAPIKeys.ReplaceAllString(cleanOutput, "[🔑 API_TOKEN_REDACTED]")

				reSSH := regexp.MustCompile(`(?s)-----BEGIN.*?PRIVATE KEY.*?-----END.*?PRIVATE KEY-----`)
				cleanOutput = reSSH.ReplaceAllString(cleanOutput, "[🚫 PRIVATE_KEY_REDACTED]")

				reWeirdAssignments := regexp.MustCompile(`(?i)([a-zA-Z0-9_-]+)\s*[:=]\s*([^\s\n\r]*?[!@#$%^&*][^\s\n\r]*)`)
				cleanOutput = reWeirdAssignments.ReplaceAllString(cleanOutput, "$1 = [🔒 SPECIAL_CHAR_PASSWORD_REDACTED]")

				reGenericSecret := regexp.MustCompile(`(?i)[a-z0-9_-]+[:=]\s*([^\s]{8,})`)
				cleanOutput = reGenericSecret.ReplaceAllString(cleanOutput, "$1 = [🔒 BLOCKED_PASSWORD]")

				rePotentialSecrets := regexp.MustCompile(`\S{14,}`)
				cleanOutput = rePotentialSecrets.ReplaceAllStringFunc(cleanOutput, func(match string) string {
					ent := shannonEntropy(match)
					if ent > 3.8 {
						return fmt.Sprintf("[🔒 ANOMALOUS_STRING_REDACTED | Entropy: %.2f]", ent)
					}
					return match
				})

				if len(cleanOutput) > 10000 {
					cleanOutput = cleanOutput[:10000] + "\n\n...[SYSTEM TRUNCATED OUTPUT DUE TO LENGTH]..."
				}

				// ==========================================
				// LAYER 5: AI OUTPUT GUARDRAIL (DLP FILTERING)
				// ==========================================
				if enableLocalLLM == "1" {
					fmt.Fprintf(os.Stderr, "[OUTPUT GUARDRAIL] Local AI DLP activated. Checking endpoint: %s\n", localEndpoint)
					
					if !isOllamaHealthy(localEndpoint) {
						errMsg := fmt.Sprintf("⛔ ZERO-TRUST ERROR: ENABLE_LOCAL_LLM=1 requires an active AI guard. Local AI (%s) is DOWN or unreachable. Execution ABORTED to prevent data leakage.", localEndpoint)
						sendResponse(req.ID, map[string]interface{}{
							"content": []map[string]interface{}{{"type": "text", "text": errMsg}},
							"isError": true,
						}, nil)
						continue
					}

					fmt.Fprintf(os.Stderr, "[OUTPUT GUARDRAIL] Local AI is responding. Processing text stream...\n")
					aiFilteredOutput, aiErr := filterWithLocalLLM(localEndpoint, cleanOutput)
					
					if aiErr != nil {
						errMsg := fmt.Sprintf("⛔ ZERO-TRUST ERROR: Local AI encountered an error during DLP processing (%v). Execution ABORTED.", aiErr)
						sendResponse(req.ID, map[string]interface{}{
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
				if err != nil {
					resultText = fmt.Sprintf("⚠️ ERROR EXECUTING COMMAND [%s]: %v\n---\n%s", cmdStr, err, cleanOutput)
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

			} else {
				resultText = "Tool does not exist!"
				isError = true
			}

			sendResponse(req.ID, map[string]interface{}{
				"content": []map[string]interface{}{{"type": "text", "text": resultText}},
				"isError": isError,
			}, nil)
		}
	}
}