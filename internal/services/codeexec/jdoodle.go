package codeexec

import (
	"errors"
	"github.com/zekurio/daemon/pkg/arrayutils"
	"strings"

	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
	"github.com/zekurio/daemon/pkg/jdoodle"
)

var (
	langs = []string{"java", "c", "cpp", "c99", "cpp14", "php", "perl", "python3", "ruby", "go", "scala", "bash", "sql", "pascal", "csharp",
		"vbn", "haskell", "objc", "ell", "swift", "groovy", "fortran", "brainfuck", "lua", "tcl", "hack", "rust", "d", "ada", "r", "freebasic",
		"verilog", "cobol", "dart", "yabasic", "clojure", "nodejs", "scheme", "forth", "prolog", "octave", "coffeescript", "icon", "fsharp", "nasm",
		"gccasm", "intercal", "unlambda", "picolisp", "spidermonkey", "rhino", "bc", "clisp", "elixir", "factor", "falcon", "fantom", "pike", "smalltalk",
		"mozart", "lolcode", "racket", "kotlin"}
)

type JdoodleWrapper struct {
	db database.Database
}

var _ ExecutorWrapper = (*JdoodleWrapper)(nil)

func NewJdoodleExecutor(ctn di.Container) *JdoodleWrapper {
	return &JdoodleWrapper{
		db: ctn.Get(static.DiDatabase).(database.Database),
	}
}

func (j *JdoodleWrapper) HasSupport(lang string) bool {
	return arrayutils.Contains(langs, lang)
}

func (j *JdoodleWrapper) NewExecutor(guildID string) (Executor, error) {
	creds, err := j.db.GetJDoodleKey(guildID)
	if err != nil {
		return nil, err
	}

	credsSplit := strings.Split(creds, "::")

	// check for length
	if len(credsSplit) != 2 {
		return &JdoodleExecutor{}, errors.New("jdoodle creds not formatted correctly")
	}

	return &JdoodleExecutor{
		clientID:     credsSplit[0],
		clientSecret: credsSplit[1],
	}, nil

}

type JdoodleExecutor struct {
	clientID     string
	clientSecret string
}

func (j *JdoodleExecutor) Execute(payload Payload) (Response, error) {
	jClient := jdoodle.New(j.clientID, j.clientSecret)
	r, err := jClient.Execute(payload.Language, payload.Code)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StdOut:  r.Output,
		MemUsed: r.Memory,
		CPUTime: r.CPUTime,
	}, nil
}
