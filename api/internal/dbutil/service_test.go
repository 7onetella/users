package dbutil

import (
	"fmt"
	"github.com/7onetella/users/api"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/internal/testutil"
	"github.com/google/uuid"
	"testing"
	"time"
)

var testuser User

func TestRegister(t *testing.T) {
	spec := &testutil.GSpec{t}

	userService := UserService{main.testDB}
	user := User{
		FirstName: "JohnFirstName",
		Created:   time.Now().Unix(),
	}
	spec.Given(fmt.Sprintf("%#v", user))

	user.ID = uuid.New().String()
	user.Created = CurrentTimeInSeconds()
	user.PlatformName = "web"

	err := userService.Register(user)
	spec.When("UserService.Register(user)")

	spec.Then()

	spec.Expect(func() {
		spec.AssertAndFailNow(err == nil, "err is nil", err)
		spec.AssertAndFailNow(len(user.ID) > 0, "id is not empty", user.ID)
	})

	testuser = user
	testuser.ID = id
}

func TestGet(t *testing.T) {
	spec := &testutil.GSpec{t}

	userService := UserService{main.testDB}
	spec.Given("id =" + testuser.ID)

	user, err := userService.Get(testuser.ID)
	spec.When("UserService.Get(id)")

	spec.Then()

	spec.Expect(func() {
		spec.AssertAndFailNow(err == nil, "err is nil", err)
		spec.AssertAndFailNow(user.FirstName == testuser.FirstName, "first name is JohnFirstName", user.FirstName)
	})
}
