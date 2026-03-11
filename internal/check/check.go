package check

import (
	"fmt"
	"os/exec"

	"github.com/six2dez/OneListForAll/internal/config"
)

type Result struct {
	ConfigOK bool
	GitOK    bool
	SevenZip bool
}

func Run(configPath string) (Result, error) {
	_, err := config.Load(configPath)
	if err != nil {
		return Result{}, err
	}

	res := Result{ConfigOK: true}
	res.GitOK = commandExists("git")
	res.SevenZip = commandExists("7z")

	if !res.GitOK {
		return res, fmt.Errorf("missing dependency: git")
	}

	return res, nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
