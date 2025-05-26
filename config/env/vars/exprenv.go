package vars

import (
	"bufio"
	"os"
	"os/user"
	"strings"
)

func getTarget() Target {
	return GetTarget()
}

func getPlatform() Target {
	return GetPlatform()
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func getUser() *user.User {
	user, err := user.Current()
	if err != nil {
		return nil
	}
	return user
}

func getShell() string {
	shell, ok := os.LookupEnv("SHELL")
	if ok {
		return shell
	}
	u := getUser()
	if u == nil || u.Username == "" {
		return ""
	}
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text()
		e := strings.Split(l, ":")
		if len(e) > 6 && e[0] == u.Username {
			return e[6]
		}
	}
	return ""
}
