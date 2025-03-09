package parse

import (
	"bufio"
	"os"
	"os/user"
	"runtime"
	"strings"
	"sync"
)

type EnvContext struct {
	Arch  string
	OS    string
	Host  string
	User  *user.User
	Shell string
}

var envContext = sync.OnceValue(func() EnvContext {
	return EnvContext{
		Arch:  getArch(),
		OS:    getOS(),
		Host:  getHost(),
		User:  getUser(),
		Shell: getShell(),
	}
})

func getArch() string {
	return runtime.GOARCH
}

func getOS() string {
	return runtime.GOOS
}

func getHost() string {
	hostName, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostName
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
