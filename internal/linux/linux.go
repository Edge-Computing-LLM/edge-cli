package linux

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func CheckUbuntu22Plus() (string, error) {
	values, err := osRelease()
	if err != nil {
		return "", err
	}
	if values["ID"] != "ubuntu" {
		return "", fmt.Errorf("unsupported distro %q: expected Ubuntu/Xubuntu base", values["ID"])
	}
	version := values["VERSION_ID"]
	majorText := strings.Split(version, ".")[0]
	major, err := strconv.Atoi(majorText)
	if err != nil {
		return "", fmt.Errorf("cannot parse Ubuntu version %q", version)
	}
	if major < 22 {
		return "", fmt.Errorf("unsupported Ubuntu version %s: expected 22.04 or newer", version)
	}
	return "Ubuntu " + version, nil
}

func osRelease() (map[string]string, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[key] = strings.Trim(val, `"`)
	}
	return values, scanner.Err()
}
