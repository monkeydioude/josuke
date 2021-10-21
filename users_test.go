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
