//GoFetch by cosmoblade

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {
	blue := color.New(color.FgHiBlue).PrintfFunc()
	white := color.New(color.FgHiWhite).PrintfFunc()

	const hamsterArt = `
         ,_---~~~~~----._         
  _,,_,*^____      _____''*g*"*, 
 / __/ /'     ^.  /      \ ^@q   f 
[  @f | @))    |  | @))   l  0 _/  
 '/   \~____ / __ \_____/    \   
  |           _l__l_           I   
  }          [______]           I  
  ]            | | |            |  
  ]             ~ ~             |  
  |                            |   
   |                           |   
`

	info := getSystemInfo()
	info["uptime"] = getUptime()

	artLines := strings.Split(hamsterArt, "\n")
	infoLines := []string{
		"",
		fmt.Sprintf("%s@%s", os.Getenv("USER"), getHostname()),
		"-------------------",
		fmt.Sprintf("OS: %s", info["os"]),
		fmt.Sprintf("Kernel: %s", info["kernel"]),
		fmt.Sprintf("Uptime: %s", info["uptime"]),
		fmt.Sprintf("CPU: %s", info["cpu"]),
		fmt.Sprintf("GPU: %s", info["gpu"]),
		fmt.Sprintf("RAM: %s", info["ram"]),
		fmt.Sprintf("Disk: %s", info["disk"]),
		fmt.Sprintf("Go: %s", runtime.Version()),
		fmt.Sprintf("Platform: %s/%s", runtime.GOOS, runtime.GOARCH),
	}

	maxWidth := 0
	for _, line := range artLines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	for i := 0; i < len(artLines); i++ {
		blue("%s", artLines[i])
		if i < len(infoLines) {
			fmt.Print(strings.Repeat(" ", maxWidth-len(artLines[i])+5))
			white("%s\n", infoLines[i])
		} else {
			fmt.Println()
		}
	}
}

func getSystemInfo() map[string]string {
	info := make(map[string]string)

	switch runtime.GOOS {
	case "darwin":
		if out, err := exec.Command("sw_vers", "-productName").Output(); err == nil {
			info["os"] = strings.TrimSpace(string(out))
			if ver, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
				info["os"] += " " + strings.TrimSpace(string(ver))
			}
		}

		if out, err := exec.Command("uname", "-r").Output(); err == nil {
			info["kernel"] = "Darwin " + strings.TrimSpace(string(out))
		}

		if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			info["cpu"] = strings.TrimSpace(string(out))
		}

		if out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Chipset Model") {
					info["gpu"] = strings.TrimSpace(strings.Split(line, ":")[1])
					break
				}
			}
		}

		info["ram"] = getMacMemory()
		info["disk"] = getMacDisk()

	case "linux":
		if out, err := exec.Command("lsb_release", "-d").Output(); err == nil {
			info["os"] = strings.TrimSpace(strings.SplitN(string(out), ":", 2)[1])
		} else if _, err := os.Stat("/etc/os-release"); err == nil {
			if out, err := exec.Command("cat", "/etc/os-release").Output(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "PRETTY_NAME=") {
						info["os"] = strings.Trim(line[len("PRETTY_NAME="):], "\"")
						break
					}
				}
			}
		}

		if out, err := exec.Command("uname", "-r").Output(); err == nil {
			info["kernel"] = strings.TrimSpace(string(out))
		}

		if out, err := exec.Command("cat", "/proc/cpuinfo").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "model name") {
					info["cpu"] = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
					break
				}
			}
		}

		if out, err := exec.Command("lspci").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "vga") || strings.Contains(strings.ToLower(line), "3d") {
					info["gpu"] = strings.TrimSpace(line)
					break
				}
			}
		}

		info["ram"] = getLinuxMemory()
		info["disk"] = getLinuxDisk()

	case "windows":
		if ver, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
			info["os"] = "Windows " + strings.TrimSpace(string(ver))
		}

		info["kernel"] = "NT " + getWindowsNTVersion()

		if out, err := exec.Command("wmic", "cpu", "get", "name").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				info["cpu"] = strings.TrimSpace(lines[1])
			}
		}

		if out, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				info["gpu"] = strings.TrimSpace(lines[1])
			}
		}

		info["ram"] = getWindowsMemory()
		info["disk"] = getWindowsDisk()
	}

	return info
}

func getMacMemory() string {
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if totalBytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64); err == nil {
			totalGB := float64(totalBytes) / (1 << 30)
			if out, err := exec.Command("vm_stat").Output(); err == nil {
				lines := strings.Split(string(out), "\n")
				var freeBytes uint64
				for _, line := range lines {
					if strings.Contains(line, "Pages free") {
						fields := strings.Fields(line)
						if len(fields) >= 3 {
							if pages, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
								freeBytes = pages * 4096
							}
						}
					}
				}
				usedGB := float64(totalBytes-freeBytes) / (1 << 30)
				percentage := usedGB * 100 / float64(totalBytes)
				return fmt.Sprintf("%.1f GB / %.1f GB (%.0f%%)", usedGB, totalGB, percentage)
			}
		}
	}
	return "unknown"
}

func getLinuxMemory() string {
	if out, err := exec.Command("free", "-m").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 3 {
				totalMB, _ := strconv.ParseFloat(fields[1], 64)
				usedMB, _ := strconv.ParseFloat(fields[2], 64)
				percentage := usedMB * 100 / totalMB
				return fmt.Sprintf("%.1f GB / %.1f GB (%.0f%%)", 
					usedMB/1024, totalMB/1024, percentage)
			}
		}
	}
	return "unknown"
}

func getWindowsMemory() string {
	if out, err := exec.Command("wmic", "OS", "get", "TotalVisibleMemorySize,FreePhysicalMemory", "/Value").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		var totalKB, freeKB uint64
		for _, line := range lines {
			if strings.HasPrefix(line, "TotalVisibleMemorySize") {
				parts := strings.Split(line, "=")
				if len(parts) > 1 {
					totalKB, _ = strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
				}
			}
			if strings.HasPrefix(line, "FreePhysicalMemory") {
				parts := strings.Split(line, "=")
				if len(parts) > 1 {
					freeKB, _ = strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
				}
			}
		}
		if totalKB > 0 {
			usedKB := totalKB - freeKB
			percentage := float64(usedKB) * 100 / float64(totalKB)
			return fmt.Sprintf("%.1f GB / %.1f GB (%.0f%%)",
				float64(usedKB)/1024/1024,
				float64(totalKB)/1024/1024,
				percentage)
		}
	}
	return "unknown"
}

func getMacDisk() string {
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				return fmt.Sprintf("%s used of %s (%s free)", fields[2], fields[1], fields[3])
			}
		}
	}
	return "unknown"
}

func getLinuxDisk() string {
	if out, err := exec.Command("df", "-h", "--output=used,size,pcent", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 3 {
				return fmt.Sprintf("%s / %s (%s used)", fields[0], fields[1], fields[2])
			}
		}
	}
	return "unknown"
}

func getWindowsDisk() string {
	if out, err := exec.Command("wmic", "logicaldisk", "where", "drivetype=3", "get", "size,freespace", "/Value").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		var totalBytes, freeBytes uint64
		for _, line := range lines {
			if strings.HasPrefix(line, "Size") {
				parts := strings.Split(line, "=")
				if len(parts) > 1 {
					totalBytes, _ = strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
				}
			}
			if strings.HasPrefix(line, "FreeSpace") {
				parts := strings.Split(line, "=")
				if len(parts) > 1 {
					freeBytes, _ = strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
				}
			}
		}
		if totalBytes > 0 {
			usedBytes := totalBytes - freeBytes
			percentage := float64(usedBytes) * 100 / float64(totalBytes)
			return fmt.Sprintf("%.1f GB / %.1f GB (%.0f%%)",
				float64(usedBytes)/1024/1024/1024,
				float64(totalBytes)/1024/1024/1024,
				percentage)
		}
	}
	return "unknown"
}

func getUptime() string {
	switch runtime.GOOS {
	case "darwin":
		if out, err := exec.Command("uptime").Output(); err == nil {
			uptimeStr := string(out)
			if idx := strings.Index(uptimeStr, "up"); idx != -1 {
				uptimePart := strings.TrimSpace(uptimeStr[idx+2:])
				if endIdx := strings.Index(uptimePart, ","); endIdx != -1 {
					return strings.TrimSpace(uptimePart[:endIdx])
				}
				return strings.TrimSpace(uptimePart)
			}
		}

	case "linux":
		if out, err := exec.Command("cat", "/proc/uptime").Output(); err == nil {
			fields := strings.Fields(string(out))
			if len(fields) > 0 {
				if uptimeSec, err := strconv.ParseFloat(fields[0], 64); err == nil {
					uptime := time.Duration(uptimeSec) * time.Second
					return formatDuration(uptime)
				}
			}
		}

	case "windows":
		if out, err := exec.Command("wmic", "os", "get", "lastbootuptime").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) >= 2 {
				bootTimeStr := strings.TrimSpace(lines[1])
				if len(bootTimeStr) >= 14 {
					year, _ := strconv.Atoi(bootTimeStr[0:4])
					month, _ := strconv.Atoi(bootTimeStr[4:6])
					day, _ := strconv.Atoi(bootTimeStr[6:8])
					hour, _ := strconv.Atoi(bootTimeStr[8:10])
					minute, _ := strconv.Atoi(bootTimeStr[10:12])
					second, _ := strconv.Atoi(bootTimeStr[12:14])
					
					bootTime := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
					uptime := time.Since(bootTime)
					return formatDuration(uptime)
				}
			}
		}
	}
	return "unknown"
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hours", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d mins", minutes))
	}

	return strings.Join(parts, " ")
}

func getWindowsNTVersion() string {
	out, err := exec.Command("cmd", "/c", "reg", "query", "HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "/v", "CurrentVersion").Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CurrentVersion") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return "unknown"
}

func getHostname() string {
	if host, err := os.Hostname(); err == nil {
		return host
	}
	return "unknown"
}
