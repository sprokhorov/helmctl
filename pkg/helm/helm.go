package helm

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sprokhorov/helmctl/pkg/config"
	helmctlKubernetes "github.com/sprokhorov/helmctl/pkg/kubernetes"
	"go.mozilla.org/sops/v3/decrypt"
	"k8s.io/client-go/kubernetes"
)

// Helm represents helm.
type Helm interface {
	Install(in *InstallOptions) error
}

// NewShellClient return new ShellClient.
func NewShellClient(cfg config.Config, opts *ShellClientOptions) (Helm, error) {
	if cfg == nil {
		return nil, errors.New("invalid config")
	}

	if opts == nil {
		opts = NewShellClientOptions(nil)
	}

	if opts.Logger == nil {
		opts.Logger = logrus.New()
	}

	return &ShellClient{cfg: cfg, opts: opts, l: opts.Logger}, nil
}

// ShellClient implements Helm as a shell call to helm binary client.
type ShellClient struct {
	cfg  config.Config
	l    *logrus.Logger
	opts *ShellClientOptions
}

// ShellClientOptions contains options for ShellClient.
type ShellClientOptions struct {
	// Run in debug mode.
	Debug bool
	// Use helm diff plugin.
	Diff bool
	// Enable dry run mode.
	DryRun bool
	// Set logger.
	Logger *logrus.Logger
	// If true repositories adding will be skipped.
	SkipRepositories bool
	// If true scripts running will be skipped.
	SkipScripts bool
	// Path to sops config file.
	SopsConfig string
	// Allow scripts running in dry run mode.
	WithScripts bool
	// Helm binary path
	HelmPath string
}

// NewShellClientOptions creates new ShellClientOptions object.
func NewShellClientOptions(logger *logrus.Logger) *ShellClientOptions {
	// getting default helm path from shell
	helmPath := "helm"
	if logger == nil {
		return &ShellClientOptions{Logger: logrus.New(), HelmPath: helmPath}
	}
	return &ShellClientOptions{Logger: logger, HelmPath: helmPath}
}

// InstallOptions contains arguments for Install method.
type InstallOptions struct {
	Release          string
	Target           string
	TargetType       config.TargetType
	KubernetesClient *kubernetes.Clientset
}

// NewInstallOptions creates new InstallOptions object.
func NewInstallOptions() *InstallOptions {
	return &InstallOptions{TargetType: config.TargetEnvironments}
}

// Install installs release or releases from config.
func (sc *ShellClient) Install(in *InstallOptions) error {
	// add global repos
	if err := sc.reposAdd(); err != nil {
		sc.l.Errorf("Failed to add helm repositories, %v", err)
		return err
	}

	var err error = nil
	in.KubernetesClient, err = helmctlKubernetes.GetKubernetesClient("")
	if err != nil {
		sc.l.Errorf("Cannot create Kubernetes client, %v", err)
		return err
	}

	// install
	if in.Release == "all" {
		return sc.installAll(in)
	}
	return sc.installOne(in)
}

func (sc *ShellClient) installOne(in *InstallOptions) error {
	r, err := sc.cfg.TargetRelease(in.Release, in.Target, in.TargetType)
	if err != nil {
		return err
	}

	if sc.opts.Diff {
		var output bytes.Buffer
		err = sc.releaseInstall(r, in, &output)
		if err != nil {
			return err
		}
		sc.l.Infof("Helm diff: %s", output.String())
		return nil
	} else {
		return sc.releaseInstall(r, in, nil)
	}
}

func (sc *ShellClient) installAll(in *InstallOptions) error {
	releases, err := sc.cfg.TargetReleases(in.Target, in.TargetType)
	if err != nil {
		return err
	}

	var output bytes.Buffer
	for _, r := range releases {
		if sc.opts.Diff {
			if err := sc.releaseInstall(r, in, &output); err != nil {
				return err
			}
		} else {
			if err := sc.releaseInstall(r, in, nil); err != nil {
				return err
			}
		}
	}

	if sc.opts.Diff {
		sc.l.Infof("Helm diff:\n%s", output.String())
	}

	return nil
}

// releaseInstall installs helm release.
func (sc *ShellClient) releaseInstall(r *config.Release, in *InstallOptions, outputBuffer *bytes.Buffer) error {
	sc.l.Infof("Install helm release %s", r.Name)
	if err := sc.scriptsExecute(r.BeforeScripts); err != nil {
		return err
	}

	// add repo
	if *r.Repository != (config.Repository{}) {
		if err := sc.repoAdd(r.Repository); err != nil {
			return err
		}
	}
	// decrypt sops
	if err := sc.sopsDecrypt(r.ValueFiles); err != nil {
		return err
	}

	// check existence of namespace and create it
	if err := helmctlKubernetes.CheckNamespace(
		in.KubernetesClient,
		r.Namespace,
		sc.opts.DryRun); err != nil {
		return err
	}

	// install
	args := sc.buildArgs(r)
	sc.l.Infof("Execute helm command: %s %s", sc.opts.HelmPath, strings.Join(args, " "))
	out, err := exec.Command(sc.opts.HelmPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s, %v", strings.ReplaceAll(string(out), "\n", ""), err)
	}

	outString := strings.ReplaceAll(string(out), "\n", "\n\t")
	if outputBuffer != nil {
		if outString != "" {
			outputBuffer.WriteString(fmt.Sprintf("Release %s:\n%s\n", r.Name, outString))
		} else {
			outputBuffer.WriteString(fmt.Sprintf("Release %s: no changes\n", r.Name))
		}
	} else {
		sc.l.Info(outString)
	}

	if err := sc.scriptsExecute(r.AfterScripts); err != nil {
		return err
	}
	sc.l.Infof("Helm release %s was installed", r.Name)

	return nil
}

func (sc *ShellClient) buildArgs(r *config.Release) []string {
	args := []string{}

	if sc.opts.Diff {
		args = append(args, "diff", "upgrade", "--allow-unreleased", r.Name, "--namespace", r.Namespace)
	} else {
		args = append(args, "upgrade", "-i", r.Name, "--namespace", r.Namespace)
	}

	if r.Version != "" {
		args = append(args, "--version", r.Version)
	}
	if *r.Atomic {
		args = append(args, "--atomic")
	}

	for _, vf := range r.ValueFiles {
		args = append(args, "-f", vf.Name)
	}

	for _, v := range r.Values {
		if v.Type == "string" {
			args = append(args, "--set-string", v.GetKeyValuePair())
			continue
		}
		args = append(args, "--set", v.GetKeyValuePair())
	}

	if sc.opts.DryRun && !sc.opts.Diff {
		args = append(args, "--dry-run")
	}

	args = append(args, r.Chart)

	return args
}

// scriptsExecute executes scripts.
func (sc *ShellClient) scriptsExecute(scripts []*string) error {
	for _, script := range scripts {

		switch {
		case sc.opts.DryRun && !sc.opts.WithScripts:
			sc.l.Infof("Skip %s script running, because of DryRun flag", *script)
			continue
		case sc.opts.SkipScripts:
			sc.l.Infof("Skip %s script running, because of SkipScripts flag", *script)
			continue
		}

		sc.l.Infof("Run script %s", *script)
		err := os.Chmod(*script, 0777)
		if err != nil {
			fmt.Println(err)
		}
		out, err := exec.Command("./" + *script).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s, %v", strings.ReplaceAll(string(out), "\n", ""), err)
		}
	}
	return nil
}

// reposAdd all helm repositories defined in config.
func (sc *ShellClient) reposAdd() error {
	for _, repo := range sc.cfg.Repositories() {
		if err := sc.repoAdd(repo); err != nil {
			return err
		}
	}
	return nil
}

// repoAdd adds new helm repository.
func (sc *ShellClient) repoAdd(repo *config.Repository) error {
	if sc.opts.SkipRepositories {
		sc.l.Infof("Skip %s repository adding, because of SkipRepositories flag", repo.Name)
		return nil
	}

	sc.l.Infof("Add helm repository %s", repo.Name)
	u, err := url.Parse(repo.URL)
	if err != nil {
		return err
	}

	if u.Host == "" || u.Scheme == "" {
		return errors.New("invalid repository url: <schema>://<host> format required")
	}

	if repo.User != "" && repo.Password != "" {
		ui := url.UserPassword(repo.User, repo.Password)
		u.User = ui
	}

	out, err := exec.Command(sc.opts.HelmPath, "repo", "add", repo.Name, u.String()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s, %v", strings.ReplaceAll(string(out), "\n", ""), err)
	}

	return nil
}

// ReposUpdate updates helm repositories.
func (sc *ShellClient) reposUpdate() error {
	sc.l.Info("Update helm repositories")
	out, err := exec.Command(sc.opts.HelmPath, "repo", "update").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s, %v", strings.ReplaceAll(string(out), "\n", ""), err)
	}

	if strings.Contains(string(out), "Unable to get") {
		return fmt.Errorf("Error: %s", strings.ReplaceAll(string(out), "\n", ""))
	}

	return nil
}

// ReposRemove removes all helm repositories defined in config.
func (sc *ShellClient) reposRemove() error {
	for _, repo := range sc.cfg.Repositories() {
		if err := sc.repoRemove(repo); err != nil {
			return err
		}
	}
	return nil
}

// RepoRemove removes helm repository.
func (sc *ShellClient) repoRemove(repo *config.Repository) error {
	out, err := exec.Command(sc.opts.HelmPath, "repo", "remove", repo.Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s, %v", strings.ReplaceAll(string(out), "\n", ""), err)
	}

	return nil
}

// sopsDecrypt decrypts file encrypted with sops.
func (sc *ShellClient) sopsDecrypt(vfs []*config.ValueFile) error {
	for _, vf := range vfs {
		if vf.GetDecrypt() {
			sc.l.Infof("Decrypt helm value file %s", vf.Name)
			b, err := decrypt.File(vf.Name, "yaml")
			if err != nil {
				return err
			}
			vf.Name = strings.ReplaceAll(vf.Name, ".yaml", ".decr.yaml")
			if err := ioutil.WriteFile(vf.Name, b, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
