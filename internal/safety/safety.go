package safety

import "strings"

var blockedPatterns = []string{
	"rm -rf /",
	"dd if=/dev/zero of=/dev/",
	"dd if=/dev/random of=/dev/",
	":(){:|:&};:",
	"mkfs",
	"wipefs",
	"> /dev/sd",
	"> /dev/nvme",
	"fdisk",
	"parted",
}

var dangerousPatterns = []string{
	"rm ",
	"rm\t",
	"sudo",
	"chmod",
	"chown",
	"shred",
	"truncate",
	"shutdown",
	"reboot",
	"halt",
	"poweroff",
	"kill ",
	"killall",
	"pkill",
	"passwd",
	"useradd",
	"userdel",
	"usermod",
	"crontab -r",
	"iptables",
	"ufw",
	"systemctl stop",
	"systemctl disable",
	"service stop",
	"launchctl unload",
}

func IsBlocked(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, p := range blockedPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func IsDangerous(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, p := range dangerousPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
