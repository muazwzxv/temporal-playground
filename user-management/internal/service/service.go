package service

import (
	user "github.com/muazwzxv/user-management/internal/service/user"
	"github.com/samber/do/v2"
)

func InjectServices(i do.Injector) {
	do.Provide(i, user.NewUserService)
}
