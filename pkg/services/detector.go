package services

import (
	"strings"
)

var portMap = map[int]string{
	22: "SSH", 80: "HTTP", 443: "HTTPS", 1433: "MSSQL", 1521: "Oracle DB",
	3000: "React/Next.js", 3001: "Dev Server", 3306: "MySQL", 4000: "Phoenix",
	5000: "Flask/Docker", 5173: "Vite", 5432: "PostgreSQL", 6379: "Redis",
	8000: "Django/PHP", 8080: "Tomcat/Java", 9000: "PHP-FPM/Minio", 27017: "MongoDB",
}

// Detect identifies a service based on port and process name.
func Detect(port int, processName string) string {
	processName = strings.ToLower(processName)
	
	// High-Priority Port-Based Labels for known dev environments
	if port == 3000 || port == 3001 {
		if strings.Contains(processName, "node") || strings.Contains(processName, "next") || strings.Contains(processName, "npm") {
			return "Next.js/React"
		}
	}

	// Keyword detection
	if strings.Contains(processName, "node") { return "Node.js" }
	if strings.Contains(processName, "next") { return "Next.js" }
	if strings.Contains(processName, "npm") { return "npm/Node" }
	if strings.Contains(processName, "mysql") { return "MySQL" }
	if strings.Contains(processName, "postgres") { return "PostgreSQL" }
	if strings.Contains(processName, "oracle") || strings.Contains(processName, "tnslsnr") { return "Oracle" }
	if strings.Contains(processName, "java") || strings.Contains(processName, "tomcat") { return "Java/Tomcat" }
	if strings.Contains(processName, "python") { return "Python" }
	if strings.Contains(processName, "docker") { return "Docker" }

	if service, ok := portMap[port]; ok {
		return service
	}
	return "Unknown"
}


// IsDev returns true if the port/process belongs to a developer tool.
func IsDev(port int, service string, processName string) bool {
	p := strings.ToLower(processName)

	// 1. Explicitly HIDE non-dev noise
	noise := []string{"chrome", "msedge", "firefox", "brave", "slack", "discord", 
		"spotify", "teams", "onedrive", "svchost", "lsass", "services"}
	for _, n := range noise {
		if strings.Contains(p, n) {
			return false 
		}
	}

	// 2. Explicitly SHOW dev stack
	devStack := []string{"node", "npm", "yarn", "java", "tomcat", "oracle", "mysql", 
		"postgres", "mongod", "redis", "httpd", "apache", "python", "php", "go", "docker", "vbox"}
	for _, d := range devStack {
		if strings.Contains(p, d) {
			return true
		}
	}

	// 3. Known dev ports fallback
	if service != "Unknown" { return true }
	if (port >= 3000 && port <= 10000) || (port >= 27000 && port <= 28000) {
		return true
	}
	return false
}


