package config

import (
    "path"
    "strings"
    "testing"

    "github.com/sirupsen/logrus"
)

func TestConfig(t *testing.T) {
    log := logrus.New()

    files := map[string]string{
        "helmctl-script-missing.yaml":              "invalid script, stat testdata/notfound.sh: no such file or directory",
        "helmctl-unknown-release.yaml":             "environment development: release test-release not found",
        "helmctl-duplicate-release.yaml":           "Duplicated release origin-name",
        "helmctl-duplicate-repositories.yaml":      "Duplicated repo something",
        "helmctl-duplicate-target-in-env.yaml":     "Duplicate Environment component name: origin-name",
        "helmctl-duplicate-target-in-project.yaml": "Duplicate Project component name: origin-name",
    }

    for file, errMsg := range files {
        cfg := NewConfigFromFile(path.Join("testdata", file), "", log, false)
        err := cfg.Load()
        if err != nil && strings.Trim(err.Error(), "\n") != errMsg {
            t.Errorf("Config Test cannot load config file %s: %v", file, err)
        }
    }

    cfg := NewConfigFromFile(path.Join("testdata", "helmctl.yaml"), "", log, false)
    if err := cfg.Load(); err != nil {
        t.Errorf("Config Test cannot load config file: %v", err)
    }
}
