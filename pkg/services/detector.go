package services

import (
	"strings"
)

var portMap = map[int]string{
	22:    "SSH",
	80:    "HTTP",
	443:   "HTTPS",
	1433:  "MSSQL",
	1521:  "Oracle DB",
	3000:  "React/Next.js",
	3001:  "Dev Server",
	3306:  "MySQL",
	4000:  "Phoenix/Elixir",
	4200:  "Angular",
	5000:  "Flask/Docker",
	5001:  "ASP.NET",
	5173:  "Vite",
	5432:  "PostgreSQL",
	6379:  "Redis",
	8000:  "Django/PHP",
	8080:  "Tomcat/Java",
	8443:  "HTTPS Alt",
	9000:  "PHP-FPM/Minio",
	27017: "MongoDB",
}

// Detect identifies a service based on port and process name.
func Detect(port int, processName string) string {
	processName = strings.ToLower(processName)

	// Check process name first for high confidence
	if strings.Contains(processName, "node") {
		return "Node.js"
	}
	if strings.Contains(processName, "mysql") {
		return "MySQL"
	}
	if strings.Contains(processName, "postgres") {
		return "PostgreSQL"
	}
	if strings.Contains(processName, "redis") {
		return "Redis"
	}
	if strings.Contains(processName, "oracle") || strings.Contains(processName, "tnslsnr") {
		return "Oracle"
	}
	if strings.Contains(processName, "java") {
		return "Java"
	}
	if strings.Contains(processName, "python") {
		return "Python"
	}
	if strings.Contains(processName, "docker") {
		return "Docker"
	}

	// Fallback to common port mapping
	if service, ok := portMap[port]; ok {
		return service
	}

	return "Unknown"
}

// IsDev returns true if the service/port is likely a developer tool or web service.
func IsDev(port int, service string) bool {
	if service != "Unknown" {
		return true
	}
	// Common dev ranges
	if (port >= 3000 && port <= 10000) || (port >= 27000 && port <= 28000) {
		return true
	}
	// Port 80/443 are dev ports if we are developing web apps
	if port == 80 || port == 443 {
		return true
	}
	return false
}

