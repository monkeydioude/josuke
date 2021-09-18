package josuke

import (
	"fmt"
	"log"
	"regexp"
	"syscall"
)

var rootUID = 0

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
		return rootUID, nil
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

	log.Printf("[INFO] switching to %s uid(%d)\n", user, id)
	return syscall.Setuid(id)
}

func switchToRoot() {
	log.Printf("[INFO] switching back to root uid(%d)\n", rootUID)
	syscall.Setuid(rootUID)
}
