package executor

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Blacklist contains shell patterns that are strictly forbidden in READ-ONLY mode.
var Blacklist = []string{
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

// IsBlocked returns true if the command string contains any blacklisted pattern.
func IsBlocked(cmdStr string) bool {
	cmdLower := strings.ToLower(cmdStr)
	for _, badWord := range Blacklist {
		if strings.Contains(cmdLower, badWord) {
			return true
		}
	}
	return false
}

// IsCurlBlocked returns true when curl/wget is used but targets a non-localhost address
// or uses a forbidden flag (POST, file-write, etc.).
func IsCurlBlocked(cmdStr string) (bool, string) {
	cmdLower := strings.ToLower(cmdStr)

	if !strings.Contains(cmdLower, "curl ") && !strings.Contains(cmdLower, "wget ") {
		return false, ""
	}

	if !strings.Contains(cmdLower, "localhost") && !strings.Contains(cmdLower, "127.0.0.1") {
		return true, "⛔ WARNING: curl/wget is ONLY PERMITTED to call [localhost] or [127.0.0.1] to prevent data exfiltration."
	}

	curlBlacklist := []string{
		" -o", "--output", "--remote-name",
		" -x", "--request",
		" -d", "--data", "--form",
	}
	for _, badFlag := range curlBlacklist {
		if strings.Contains(cmdLower, badFlag) {
			return true, "⛔ WARNING: curl is restricted to pure HTTP GET. POST/PUT/DELETE, data payloads, and file saving are blocked."
		}
	}

	return false, ""
}

// RunBash executes cmdStr in a bash shell with a 60-second timeout.
// Returns a RunResult containing the combined output, any execution error, and a timeout flag.
func RunBash(cmdStr string) RunResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()

	return RunResult{
		Output:  string(out),
		Err:     err,
		Timeout: ctx.Err() == context.DeadlineExceeded,
	}
}
