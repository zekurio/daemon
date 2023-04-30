package inits

import (
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/codeexec"
)

func InitCodeexec(ctn di.Container) codeexec.ExecutorWrapper {
	return codeexec.NewJdoodleExecutor(ctn)
}
