package config

import (
    "fmt"
    "path"
    "strings"
    "testing"

    "github.com/kr/pretty"
    "github.com/sirupsen/logrus"
)

func checkErr(expected string, got error) error {
    switch {
    case expected == "":
        return got
    case expected != "" && expected != got.Error():
        return fmt.Errorf("Wrong error, expected %v got %v", expected, got)
    case expected != "" && expected == got.Error():
        return nil
    }
    return nil
}

// EnvironMock implements Environ.
type EnvironMock struct{}

// Get mocks Environ.Get method.
func (env *EnvironMock) Get(s string) string {
    if strings.HasPrefix(s, "$NF") {
        return ""
    }
    return "MOCKED"
}

func TestFile(t *testing.T) {
    log := logrus.New()

    files := map[string]string{
        "helmctl.yaml":                      "",
        "helmctl-env-lookup-not-found.yaml": "Could not unmarshal data : Env variable: NF_REPO_NAME is not found, line: 4. The data could not be unmarshalled as yaml",
        "helmctl-env-lookup.yaml":           "",
        "helmctl-broken-include.yaml":       "Could not unmarshal data : Include something_missing : open testdata/something_missing: no such file or directory, line: 4. The data could not be unmarshalled as yaml",
    }

    for file, errMsg := range files {
        _, err := parseConfigFile(path.Join("testdata", file), "", log)
        if err != nil && strings.Trim(err.Error(), "\n") != errMsg {
            t.Errorf("%s: %v\n", file, pretty.Sprint(err))
        }
    }
}
