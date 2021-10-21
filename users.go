package josuke

import (
	"log"
	"os"
	"os/user"
	"regexp"
	"strconv"
)

type User struct {
	Uid  uint32
	Gid  uint32
	Name string
}

var currentUser User = User{}
var defaultUser User = User{}

func GetCurrentUser() User {
	return currentUser
}

func isWindows() bool {
    return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

func init() {
	if isWindows() {
		return
	}
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		panic(err)
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		panic(err)
	}
	defaultUser = User{
		Uid:  uint32(uid),
		Gid:  uint32(gid),
		Name: user.Username,
	}
	switchToDefaultUser()
}

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

func getUserID(userName string) (*user.User, error) {
	u, err := user.Lookup(userName)

	if err != nil {
		return nil, err
	}
	return u, nil
}

func SwitchUser(userName string) error {
	user, err := getUserID(userName)

	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return err
	}

	log.Printf("[INFO] switching to %+v\n", user)
	currentUser = User{
		Uid:  uint32(uid),
		Gid:  uint32(gid),
		Name: user.Username,
	}
	return nil
}

func switchToDefaultUser() {
	currentUser = defaultUser
}
