package josuke

import (
	"fmt"
	"regexp"
	"syscall"
)

func isSwitchUserCall(str string) (bool, string) {
	regxp, err := regexp.Compile("^%user_(.+)%$")
	s := 1
	if err != nil {
		return false, ""
	}

	res := regxp.FindAllStringSubmatch(str, s)

	if len(res) < 1 || len(res[0][s]) < s {
		return false, ""
	}
	return true, res[0][s]
}

func getUserID(user string, users map[string]int) (int, error) {
	if user == "root" {
		return 0, nil
	}
	if _, ok := users[user]; !ok {
		return 0, fmt.Errorf("could not find user's ID for user %s", user)
	}
	return users[user], nil
}

func switchUser(user string, users map[string]int) error {
	id, err := getUserID(user, users)

	if err != nil {
		return err
	}

	return syscall.Setuid(id)
}
