package helm

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sprokhorov/helmctl/pkg/config"
)

func TestHelmInstallOne(t *testing.T) {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	cfg := config.NewConfigFromFile("testdata/helmctl.yaml", "", log, false)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	shellClientOptions := NewShellClientOptions(log)
	shellClientOptions.DryRun = true

	h, err := NewShellClient(cfg, shellClientOptions)
	if err != nil {
		t.Fatalf("Failed to create ShellClient, %v", err)
	}

	in := &InstallOptions{
		Release:    "gitlab-runner-two",
		Target:     "development",
		TargetType: config.TargetEnvironments,
	}
	if err := h.Install(in); err != nil {
		t.Fatal(err)
	}
}

func TestHelmInstallAll(t *testing.T) {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	cfg := config.NewConfigFromFile("testdata/helmctl.yaml", "", log, false)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	shellClientOptions := NewShellClientOptions(log)
	shellClientOptions.DryRun = true

	h, err := NewShellClient(cfg, shellClientOptions)
	if err != nil {
		t.Fatalf("Failed to create ShellClient, %v", err)
	}

	in := &InstallOptions{
		Release:    "all",
		Target:     "development",
		TargetType: config.TargetEnvironments,
	}
	if err := h.Install(in); err != nil {
		t.Fatal(err)
	}
}
