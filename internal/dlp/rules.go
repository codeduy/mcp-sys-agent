package dlp

import (
	"fmt"
	"strings"
)

// TechProfile defines a technology category, its command trigger words,
// and the message to display when the DLP blinds the command output.
type TechProfile struct {
	TechName     string
	TriggerWords []string
	BlindMessage string
}

// TechProfiles is the master list of all contextual blindness rules (Layer 2 DLP).
var TechProfiles = []TechProfile{
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

// CheckContextualBlindness checks the command string against all TechProfiles.
// Returns (isBlinded bool, resultText string).
func CheckContextualBlindness(cmdStr string) (bool, string) {
	cmdLower := strings.ToLower(cmdStr)
	for _, profile := range TechProfiles {
		for _, trigger := range profile.TriggerWords {
			if strings.Contains(cmdLower, trigger) {
				resultText := fmt.Sprintf(
					"✅ Command [%s] is valid. File/Directory EXISTS.\n"+
						"🛡️ DLP SYSTEM: Detected [%s] technology. Activated BLIND MODE (content ignored) to prevent data leak: %s.\n"+
						"👉 AI INSTRUCTION: Assume this configuration file is error-free, proceed to investigate other log files or service states.",
					cmdStr, profile.TechName, profile.BlindMessage)
				return true, resultText
			}
		}
	}
	return false, ""
}
