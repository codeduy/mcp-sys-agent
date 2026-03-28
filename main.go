package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// --- Cấu trúc JSON-RPC chuẩn MCP ---
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
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

// --- THUẬT TOÁN ĐO LƯỜNG ĐỘ HỖN LOẠN (SHANNON ENTROPY) ---
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	// Tăng kích thước Buffer để đọc log dài
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
					"version": "Ultimate-DLP-6.0",
				},
				"instructions": `Bạn là Kỹ sư Hệ thống (DevOps) và Quản trị viên Linux chuyên nghiệp. Nhiệm vụ của bạn là rà soát và chẩn đoán Server.
                    QUY TẮC TỐI THƯỢNG:
                    1. GIAO TIẾP: 100% bằng TIẾNG VIỆT.
                    2. QUYỀN HẠN: Bạn đang ở chế độ Strict READ-ONLY. Mọi lệnh làm thay đổi trạng thái hệ thống (rm, systemctl, chmod...) sẽ bị chặn cứng. Đừng cố gắng lách luật.
                    3. TRUY VẤN API: Nếu cần gọi API local (như Prometheus/Grafana), CHỈ dùng lệnh 'curl' gọi tới localhost hoặc 127.0.0.1 bằng phương thức GET (Truyền tham số thẳng vào URL, cấm tuyệt đối dùng cờ -d, --data, -X POST hoặc lưu file).
                    4. HIỆU NĂNG TÌM KIẾM: TUYỆT ĐỐI KHÔNG dùng lệnh 'cat' để đọc các file log/cấu hình lớn. Hãy ưu tiên dùng 'tail -n 50', 'head', 'grep -i', 'find', hoặc 'journalctl -n 50' để lọc thông tin chính xác.
                    5. ỨNG XỬ VỚI BẢO MẬT (DLP): Nếu kết quả trả về có chứa 'MÙ LÒA', 'CHUỖI_BẤT_THƯỜNG', 'REDACTED', hãy tự động hiểu rằng hệ thống DLP đã che dữ liệu nhạy cảm. KHÔNG cố gắng tìm lệnh khác để đọc lại các file đó.
                    6. TỰ KIỂM DUYỆT (SELF-CENSORSHIP): Trước khi gọi lệnh, hãy tự đánh giá. Tuyệt đối KHÔNG thực thi các lệnh có rủi ro gây tổn hại hệ thống hoặc có mục đích truy xuất/leak nội dung nhạy cảm (như password, token). Hãy tự động bỏ qua các hướng tiếp cận đó và tìm phương pháp chẩn đoán an toàn hơn.
                    7. XỬ LÝ CHE NHẦM (FALSE POSITIVES): Do hệ thống DLP bảo vệ rất mạnh, đôi khi thông số vô hại bị che nhầm. Nếu phần dữ liệu bị che là **BẮT BUỘC PHẢI CÓ** để chẩn đoán, HÃY DỪNG LẠI VÀ HỎI TÔI (USER): "Tôi cần giá trị bị che ở trường [X], bạn có thể copy thủ công dán vào đây không?". Tuyệt đối không tự lách luật.
                    8. ĐÓNG GÓP TỐI ƯU AGENT (FEEDBACK LOOP): Trong quá trình phân tích, nếu bạn nhận thấy logic bộ lọc của Agent (file main.go) có điểm bất cập gây cản trở quá trình debug, HOẶC phát hiện lỗ hổng bảo mật có thể làm leak dữ liệu nhạy cảm, hãy CHỦ ĐỘNG GỢI Ý giải pháp tối ưu/sửa code. Tôi sẽ xem xét và tự cập nhật lại file main.go.`,
				"capabilities": map[string]interface{}{"tools": map[string]interface{}{}},
			}, nil)

		case "tools/list":
			sendResponse(req.ID, map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "run_bash_command",
						"description": "Thực thi lệnh Bash READ-ONLY trên Server. Tự động chặn lệnh phá hoại và tự động làm mù/che mờ các thông tin nhạy cảm.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"command": map[string]interface{}{
									"type":        "string",
									"description": "Lệnh bash cần chạy. Ví dụ: 'journalctl -xe | tail -n 50' hoặc 'free -m'.",
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

				// ==========================================
                // LỚP 1: BỘ LỌC THÉP (BLACKLIST LỆNH)
                // ==========================================
                blacklist := []string{
                    "rm ", "mv ", "cp ", "touch ", "mkdir ", "rmdir ", "dd ", "ln ", "truncate ", "tee ",
                    ">", ">>", "nano ", "vim ", "vi ", "editor ", "sed -i ", "awk -i ",
                    "chmod ", "chown ", "chgrp ", "passwd ", "useradd ", "userdel ", "usermod ", "su ", "sudo ", "visudo ",
                    "scp ", "rsync ", "nc ", "netcat ", "ftp ", "sftp ", // Đã mở khóa wget và curl
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
                    resultText = fmt.Sprintf("⛔ CẢNH BÁO: Lệnh [%s] bị CHẶN! Agent chỉ chạy ở chế độ READ-ONLY.", cmdStr)
                    sendResponse(req.ID, map[string]interface{}{
                        "content": []map[string]interface{}{{"type": "text", "text": resultText}},
                        "isError": true,
                    }, nil)
                    continue // Nhảy qua vòng lặp, chờ lệnh mới từ AI
                }

                // ==========================================
                // LỚP 1.5: KIỂM SOÁT GIAO TIẾP API (STRICT HTTP GET LOCAL)
                // ==========================================
                if strings.Contains(cmdLower, "curl ") || strings.Contains(cmdLower, "wget ") {
                    // Quy tắc 1: CẤM GIAO TIẾP NGOÀI (Chống tuồn dữ liệu)
                    if !strings.Contains(cmdLower, "localhost") && !strings.Contains(cmdLower, "127.0.0.1") {
                        sendResponse(req.ID, map[string]interface{}{
                            "content": []map[string]interface{}{{"type": "text", "text": "⛔ CẢNH BÁO: Lệnh curl/wget CHỈ ĐƯỢC PHÉP gọi tới [localhost] hoặc [127.0.0.1] để bảo vệ mạng."}},
                            "isError": true,
                        }, nil)
                        continue
                    }

                    // Quy tắc 2: CHỈ HTTP GET & CẤM LƯU FILE
                    // Đã sửa lại để không chặn nhầm (Phải có dấu cách)
                    curlBlacklist := []string{
                        " -o", "--output", "--remote-name", // Chặn lưu file
                        " -x", "--request", // Chặn đổi HTTP Method (Sẽ cover luôn -X POST, -X DELETE...)
                        " -d", "--data", "--form", // Chặn gửi Body data
                    }
                    
                    for _, badFlag := range curlBlacklist {
                        if strings.Contains(cmdLower, badFlag) {
                            sendResponse(req.ID, map[string]interface{}{
                                "content": []map[string]interface{}{{"type": "text", "text": "⛔ CẢNH BÁO: Lệnh curl bị giới hạn ở HTTP GET thuần túy. Cấm POST/PUT/DELETE, cấm --data và cấm lưu file."}},
                                "isError": true,
                            }, nil)
                            isBlocked = true
                            break
                        }
                    }
                    // Nếu cờ bị chặn ở Lớp 1.5 bật lên, chặn không cho chạy tiếp
                    if isBlocked {
                        continue
                    }
                }

				// ==========================================
				// LỚP 2: MÙ LÒA THEO NGỮ CẢNH (CONTEXT DLP)
				// ==========================================
				type TechProfile struct {
					TechName     string
					TriggerWords []string
					BlindMessage string
				}

				techProfiles := []TechProfile{
					{"Node.js / Web Frameworks", []string{".env"}, "chứa các biến môi trường và API Keys"},
					{"WordPress", []string{"wp-config.php"}, "chứa mật khẩu Database clear-text"},
					{"Docker / Container", []string{"docker-compose.yml", "docker-compose.yaml"}, "chứa tham số nhạy cảm của các services"},
					{"Kubernetes", []string{".kube/config", "kubeconfig"}, "chứa chứng chỉ quản trị Cluster K8s"},
					{"AWS Cloud", []string{".aws/credentials", ".aws/config"}, "chứa AWS Access/Secret Key"},
					{"SSH / Git Keys", []string{"id_rsa", "id_ed25519", "authorized_keys", ".git/credentials"}, "chứa Private Key / Khóa ngoại"},
					{"Hệ thống Linux Core", []string{"/etc/shadow", "/etc/sudoers"}, "chứa mã băm mật khẩu và phân quyền"},
					
					// [BẢN VÁ]: Bịt kín lỗ hổng đọc lén lịch sử gõ phím
					{"Lịch sử Terminal", []string{".bash_history", ".zsh_history", ".mysql_history", ".psql_history"}, "chứa các lệnh và mật khẩu từng được gõ thủ công"},
					
					// Tử huyệt DirectAdmin
					{"DirectAdmin Core", []string{"/usr/local/directadmin/conf/"}, "chứa cấu hình lõi và mật khẩu Root MySQL"},
					{"DirectAdmin Users Data", []string{"/usr/local/directadmin/data/users/"}, "chứa dữ liệu nhạy cảm của toàn bộ khách hàng thuê Hosting"},
					{"DirectAdmin Setup Log", []string{"setup.txt"}, "chứa mật khẩu Admin cài đặt ban đầu (Clear-text)"},
					{"Mail Server Configs", []string{"/etc/virtual/"}, "chứa danh sách email và mã băm mật khẩu hòm thư khách hàng"},
					{"MariaDB / MySQL", []string{"my.cnf", ".my.cnf", "debian.cnf"}, "chứa thông tin xác thực Database quyền cao nhất"},
					{"CSF Firewall", []string{"/etc/csf/csf.allow", "/etc/csf/csf.deny", "/etc/csf/csf.ignore"}, "chứa danh sách IP khách hàng và rule tường lửa"},
					{"Source Code Khách hàng", []string{"wp-config.php", "configuration.php", ".env"}, "chứa mật khẩu Database của từng website"},
				}

				isBlinded := false
				for _, profile := range techProfiles {
					for _, trigger := range profile.TriggerWords {
						if strings.Contains(cmdLower, trigger) {
							isBlinded = true
							resultText = fmt.Sprintf(
								"✅ Lệnh [%s] hợp lệ. Tệp/Thư mục CÓ TỒN TẠI.\n"+
								"🛡️ HỆ THỐNG DLP: Nhận diện công nghệ [%s]. Đã kích hoạt cơ chế MÙ LÒA (Bỏ qua nội dung) để ngăn rò rỉ dữ liệu %s lên Cloud.\n"+
								"👉 HƯỚNG DẪN CHO AI: Hãy giả định file cấu hình này không có lỗi, tiếp tục điều tra các file log hoặc trạng thái service khác.",
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
				// LỚP 3 & 4: THỰC THI & RÂY LỌC ĐẦU RA (REGEX + ENTROPY)
				// ==========================================
				
				// [BẢN VÁ CRITICAL]: Ép Timeout 60 giây (Đủ dài cho log nặng, đủ ngắn để chống treo)
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				
				cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
				out, err := cmd.CombinedOutput()
				rawOutput := string(out)

				// Xử lý báo lỗi thông minh nếu lệnh bị Timeout ép chết
				if ctx.Err() == context.DeadlineExceeded {
					resultText = fmt.Sprintf("⛔ LỖI TIMEOUT (Quá 60 giây): Lệnh [%s] đã bị ngắt tự động.\n"+
						"Nguyên nhân có thể do:\n"+
						"1. Lệnh của bạn bị treo vô tận (VD: dùng 'top', 'tail -f', 'ping' không giới hạn).\n"+
						"2. File log quá lớn khiến lệnh 'grep' / 'cat' chạy quá lâu.\n"+
						"👉 HƯỚNG DẪN CHO AI: Hãy tối ưu lại lệnh! Hãy dùng 'tail -n 1000' trước khi grep, hoặc thu hẹp mốc thời gian của 'journalctl'.", cmdStr)
					
					sendResponse(req.ID, map[string]interface{}{
						"content": []map[string]interface{}{{"type": "text", "text": resultText}},
						"isError": true,
					}, nil)
					continue
				}

				// --- [BẢN VÁ MỚI]: TỰ ĐỘNG RÃ ĐÔNG JSON ---
                // Tránh việc Minified JSON (không dấu cách) bị Regex/Entropy nuốt trọn
                var jsonParsed interface{}
                if errParse := json.Unmarshal([]byte(rawOutput), &jsonParsed); errParse == nil {
                    // Nếu đúng là JSON, format lại cho đẹp (thêm newline và khoảng trắng)
                    if prettyBytes, errIndent := json.MarshalIndent(jsonParsed, "", "  "); errIndent == nil {
                        rawOutput = string(prettyBytes)
                    }
                }

				// Lớp 3: Regex chuẩn
				reURI := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+-.]*:\/\/[^\s:@\/]+:)([^\s:@\/]+)(@[^\s\/]+)`)
				cleanOutput := reURI.ReplaceAllString(rawOutput, "$1[URI_PASSWORD_REDACTED]$3")

				reSecret := regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api[_-]?key|private[_-]?key|salt|bearer|client[_-]?secret)\s*[:=]\s*([^\s\n\r"']+)`)
				cleanOutput = reSecret.ReplaceAllString(cleanOutput, "$1 = [🔒 DATA_REDACTED]")

				reAPIKeys := regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16}|sk_live_[0-9a-zA-Z]{24}|ghp_[0-9a-zA-Z]{36}|xox[bap]-[0-9a-zA-Z_-]+|eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+)`)
				cleanOutput = reAPIKeys.ReplaceAllString(cleanOutput, "[🔑 API_TOKEN_REDACTED]")

				reSSH := regexp.MustCompile(`(?s)-----BEGIN.*?PRIVATE KEY.*?-----END.*?PRIVATE KEY-----`)
				cleanOutput = reSSH.ReplaceAllString(cleanOutput, "[🚫 PRIVATE_KEY_REDACTED]")

				// Lớp 3.5: Bắt mọi phép gán biến có chứa ký tự đặc biệt (Mật khẩu dị)
				reWeirdAssignments := regexp.MustCompile(`(?i)([a-zA-Z0-9_-]+)\s*[:=]\s*([^\s\n\r]*?[!@#$%^&*][^\s\n\r]*)`)
				cleanOutput = reWeirdAssignments.ReplaceAllString(cleanOutput, "$1 = [🔒 MẬT_KHẨU_KÝ_TỰ_ĐẶC_BIỆT]")

				// Lớp 3.6: Bắt mọi thứ trông giống mật khẩu/key sau dấu gán
				// (Bắt chuỗi không có dấu cách, dài > 8 ký tự, có trộn lẫn chữ/số/ký tự đặc biệt)
				reGenericSecret := regexp.MustCompile(`(?i)[a-z0-9_-]+[:=]\s*([^\s]{8,})`)
				cleanOutput = reGenericSecret.ReplaceAllString(cleanOutput, "$1 = [🔒 MẬT_KHẨU_BỊ_CHẶN]")

				// Lớp 4: Entropy Scanner (Đánh hơi chuỗi hỗn loạn)
				// Bắt mọi chuỗi LIỀN MẠCH (không có dấu cách) từ 14 ký tự trở lên
				rePotentialSecrets := regexp.MustCompile(`\S{14,}`)
				
				cleanOutput = rePotentialSecrets.ReplaceAllStringFunc(cleanOutput, func(match string) string {
					ent := shannonEntropy(match)
					if ent > 3.8 {
						return fmt.Sprintf("[🔒 CHUỖI_BẤT_THƯỜNG_ĐÃ_CHE | Entropy: %.2f]", ent)
					}
					return match
				})

				// Xử lý kết quả trả về
				if err != nil {
					resultText = fmt.Sprintf("⚠️ LỖI CHẠY LỆNH [%s]: %v\n---\n%s", cmdStr, err, cleanOutput)
					isError = true
				} else {
					if len(cleanOutput) == 0 {
						resultText = fmt.Sprintf("✅ Lệnh [%s] chạy thành công (Không có output trả về).", cmdStr)
					} else {
						// Chống tràn RAM (Cắt ở 10,000 ký tự)
						if len(cleanOutput) > 10000 {
							cleanOutput = cleanOutput[:10000] + "\n\n...[HỆ THỐNG ĐÃ CẮT BỚT OUTPUT VÌ QUÁ DÀI]..."
						}
						resultText = fmt.Sprintf("Kết quả lệnh [%s]:\n---\n%s", cmdStr, cleanOutput)
					}

					footer := "\n\n[SYSTEM REMINDER: Bạn đang ở chế độ READ-ONLY. Tuyệt đối tuân thủ Rule 6 (Tự kiểm duyệt) và Rule 7 (Hỏi User nếu dữ liệu quan trọng bị che).]"
                	resultText += footer
				}

			} else {
				resultText = "Tool không tồn tại!"
				isError = true
			}

			sendResponse(req.ID, map[string]interface{}{
				"content": []map[string]interface{}{{"type": "text", "text": resultText}},
				"isError": isError,
			}, nil)
		}
	}
}