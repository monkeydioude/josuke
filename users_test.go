package josuke

import (
	"testing"
)

func TestICanGetUserUIDFromConf(t *testing.T) {
	if _, u := isSwitchUserCall("%user_test1%"); u != "test1" {
		t.Fail()
	}
}

func TestIShouldThrowErrorWhenGettingUnexpectedUID(t *testing.T) {
	if _, u := isSwitchUserCall("%user_test2%"); u == "test2_1" {
		t.Fail()
	}
}

func TestICanRetrieveIDFromExistingUID(t *testing.T) {
	users := map[string]int{
		"test3": 69,
	}
	user := "test3"

	id, _ := getUserID(user, users)

	if id != 69 {
		t.Fail()
	}
}

func TestICantRetrieveIDFromNonExistingUID(t *testing.T) {
	users := map[string]int{
		"test4": 69,
	}
	user := "test4_1"

	_, ok := getUserID(user, users)

	if ok == nil {
		t.Fail()
	}
}

func TestIRetrieveRootWithoutNeedingToSpecifyIt(t *testing.T) {
	users := map[string]int{
		"test5": 69,
	}
	user := "root"

	id, ok := getUserID(user, users)

	if ok != nil || id != 0 {
		t.Fail()
	}
}
