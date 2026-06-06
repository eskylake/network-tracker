package check

import "strings"

func XVPNChecker(command string) Checker {
	return CommandChecker(NameXVPN, CategoryVPN, command, nil, func(output string, err error) (Status, string) {
		lower := strings.ToLower(output)
		if err != nil {
			if strings.Contains(lower, "password") || strings.Contains(lower, "sudo") {
				return StatusWarning, "xvpn command should not require sudo"
			}
			return StatusWarning, "xvpn status unavailable"
		}
		if strings.Contains(lower, "connected") {
			return StatusOK, "connected"
		}
		if strings.Contains(lower, "disconnect") || strings.Contains(lower, "not connected") {
			return StatusWarning, "not connected"
		}
		return StatusUnknown, firstLine(output, "status unknown")
	})
}

func V2RayAServiceChecker(command string) Checker {
	return CommandChecker(NameV2RayAService, CategoryVPN, command, nil, func(output string, err error) (Status, string) {
		trimmed := strings.TrimSpace(output)
		if err != nil {
			if trimmed == "inactive" || trimmed == "failed" {
				return StatusWarning, trimmed
			}
			return StatusWarning, "service status unavailable"
		}
		if trimmed == "active" {
			return StatusOK, "active"
		}
		return StatusWarning, firstLine(trimmed, "not active")
	})
}

func ProcessChecker(name, pattern string) Checker {
	return CommandChecker(name, CategoryVPN, "pgrep -a "+pattern, nil, func(output string, err error) (Status, string) {
		if err != nil || strings.TrimSpace(output) == "" {
			return StatusWarning, "process not found"
		}
		return StatusOK, "process found"
	})
}
