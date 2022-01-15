package auth

import (
	"fmt"
	"log"
	"strings"

	"github.com/prabhatsharma/zinc/pkg/zutil"
)

func init() {
	firstStart, err := IsFirstStart()
	if err != nil {
		fmt.Println(err)
	}
	if firstStart {
		// create default user from environment variable
		admin := zutil.GetEnv("ZINC_ADMIN", "admin:admin")
		user, password := Cut(admin, ":")
		if user == "" {
			log.Fatal("ZINC_ADMIN must be set on first start. You should also change the credentials after first login.")
		}
		CreateUser(user, "", password, "admin")
	}
}

func Cut(s, sep string) (a, b string) {
	idx := strings.Index(s, sep)
	if idx < 0 {
		return s, ""
	}

	return s[:idx], s[idx+len(sep):]
}

func IsFirstStart() (bool, error) {
	userList, err := GetAllUsersWorker()
	if err != nil {
		return true, err
	}

	if userList.Hits.Total.Value == 0 {
		return true, nil
	}

	return false, nil
}
