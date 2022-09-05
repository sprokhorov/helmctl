package cmd

import (
	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/sprokhorov/helmctl/pkg/config"
	"github.com/sprokhorov/helmctl/pkg/helm"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
}

// globalOptions contains values of defined flags.
type globalOptions struct {
	ConfigFile string
	SchemaFile string
	HelmPath   string
	Debug      bool
	DryRun     bool
}

// New returns root command object.
func New() *cobra.Command {
	opts := &globalOptions{}

	cmd := &cobra.Command{
		Use:   "helmctl",
		Short: "helmctl is a helm release wrapper",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opts.Debug {
				log.SetLevel(logrus.DebugLevel)
				log.Debug("Debug mode enabled")
			}

			if opts.DryRun {
				log.Info("Dry-run mode is enabled")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "file", "f", "helmctl.yaml", "path to config file")
	cmd.PersistentFlags().StringVarP(&opts.SchemaFile, "schema", "s", "", "path to json schema")
	cmd.PersistentFlags().BoolVar(&opts.Debug, "debug", false, "debug mode")
	cmd.PersistentFlags().BoolVar(&opts.DryRun, "dry-run", false, "dry run mode")

	cmd.AddCommand(
		newValidateCmd(opts), newInstallCmd(opts), newPlanCmd(opts))

	return cmd
}

// newValidateCmd returns new validate command.
func newValidateCmd(gopts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file.",
		Run: func(cmd *cobra.Command, args []string) {
			validate(gopts)
		},
	}

	return cmd
}

// validate conduct config file valiadation
func validate(gopts *globalOptions) config.Config {
	log.Infof("Load config from file %s", gopts.ConfigFile)

	cfg := config.NewConfigFromFile(gopts.ConfigFile, gopts.SchemaFile, log, gopts.DryRun)

	if err := cfg.Load(); err != nil {
		log.Fatalf("Failed to load config file %v", err)
	}

	log.Info("Config is valid")

	return cfg
}

// planOptions contains values of defined flags for plan command.
type planOptions struct {
	Release     string
	Environment string
	ProjectID   string
	Config      bool
}

// plan write to stdout what releases with what params will be installed
// implements next logic:
// if env and release provided -> plan this release for this env
// if prj and release provided -> plan this release for this prj
// if env provided -> plan for this env
// if prj provided -> plan for this prj
// else -> print for all envs and prjs
func plan(gopts *globalOptions, planOpts *planOptions) {
	cfg := validate(gopts)
	if planOpts.Config {
		pretty.Println("Configuration Go object:")
		pretty.Println(cfg)
	}

	if planOpts.Environment != "" && planOpts.Release != "" {
		release, err := cfg.TargetRelease(planOpts.Release, planOpts.Environment, config.TargetEnvironments)
		if err != nil {
			log.Error(err)
			return
		}
		pretty.Printf("Environment '%s' release:\n", planOpts.Environment)
		pretty.Println(release)
	} else if planOpts.ProjectID != "" && planOpts.Release != "" {
		release, err := cfg.TargetRelease(planOpts.Release, planOpts.ProjectID, config.TargetProjects)
		if err != nil {
			log.Error(err)
			return
		}
		pretty.Printf("ProjectID '%s' release:\n", planOpts.ProjectID)
		pretty.Println(release)
	} else if planOpts.Environment != "" {
		releases, err := cfg.TargetReleases(planOpts.Environment, config.TargetEnvironments)
		if err != nil {
			log.Error(err)
			return
		}
		pretty.Printf("Environment '%s' releases:\n", planOpts.Environment)
		for _, r := range releases {
			pretty.Println(r)
		}
	} else if planOpts.ProjectID != "" {
		releases, err := cfg.TargetReleases(planOpts.ProjectID, config.TargetProjects)
		if err != nil {
			log.Error(err)
			return
		}
		pretty.Printf("Project '%s' releases:\n", planOpts.ProjectID)
		for _, r := range releases {
			pretty.Println(r)
		}
	} else {
		for _, environment := range cfg.Environments() {
			releases, err := cfg.TargetReleases(environment, config.TargetEnvironments)
			if err != nil {
				log.Error(err)
				return
			}
			pretty.Printf("Environment '%s' releases:\n", environment)
			for _, r := range releases {
				pretty.Println(r)
			}
		}
		for _, project := range cfg.Projects() {
			releases, err := cfg.TargetReleases(project, config.TargetProjects)
			if err != nil {
				log.Error(err)
				return
			}
			pretty.Printf("Project '%s' releases:\n", project)
			for _, r := range releases {
				pretty.Println(r)
			}
		}
	}
}

// newValidateCmd returns new validate command.
func newPlanCmd(gopts *globalOptions) *cobra.Command {

	planOpts := planOptions{}

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Print configuration file to stdout.",
		Run: func(cmd *cobra.Command, args []string) {
			plan(gopts, &planOpts)
		},
	}

	cmd.Flags().StringVarP(&planOpts.Environment, "environment", "e", "", "environment name")
	cmd.Flags().StringVarP(&planOpts.ProjectID, "project", "p", "", "GCP project id")
	cmd.Flags().StringVarP(&planOpts.Release, "release", "r", "", "Release name")
	cmd.Flags().BoolVar(&planOpts.Config, "config", false, "Dump configuration")

	return cmd
}

// installOptions contains values of defined flags for install command.
type installOptions struct {
	release        string
	environment    string
	projectID      string
	helmClientOpts *helm.ShellClientOptions
	cfg            config.Config
}

// newInstallCmd returns new install command.
func newInstallCmd(gopts *globalOptions) *cobra.Command {
	helmClientOpts := &helm.ShellClientOptions{
		Logger: log,
	}
	iopts := &installOptions{helmClientOpts: helmClientOpts}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install release.",
		Run: func(cmd *cobra.Command, args []string) {
			dr, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				log.Fatal(err)
			}
			iopts.helmClientOpts.DryRun = dr

			d, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Fatal(err)
			}
			iopts.helmClientOpts.Debug = d

			if len(args) < 1 {
				log.Fatal("Release is missing, please set release name or all")
			}
			iopts.release = args[0]
			iopts.cfg = validate(gopts)
			install(iopts)
		},
	}

	cmd.Flags().StringVarP(&iopts.environment, "environment", "e", "", "environment name")
	cmd.Flags().StringVarP(&iopts.projectID, "project", "p", "", "GCP project id")
	cmd.Flags().StringVar(&helmClientOpts.SopsConfig, "sops-config", ".sops.yaml", "path to sops config")
	cmd.Flags().BoolVar(&helmClientOpts.Diff, "diff", false, "show helm diff")
	cmd.Flags().BoolVar(&helmClientOpts.SkipRepositories, "skip-repositories", false, "skip processing repositories")
	cmd.Flags().BoolVar(&helmClientOpts.SkipScripts, "skip-scripts", false, "skip defined scripts")
	cmd.Flags().BoolVar(&helmClientOpts.WithScripts, "with-scripts", false, "enable defined scripts")
	cmd.Flags().StringVarP(&helmClientOpts.HelmPath, "helm", "H", "helm", "path to helm binary")

	return cmd
}

func install(iopts *installOptions) {
	if iopts.environment != "" && iopts.projectID != "" {
		log.Fatal("Only one target allowed, please set --project or --environment")
	}

	if iopts.environment == "" && iopts.projectID == "" {
		log.Fatal("No target, please set --project or --environment")
	}

	h, err := helm.NewShellClient(iopts.cfg, iopts.helmClientOpts)
	if err != nil {
		log.Fatal(err)
	}

	in := helm.NewInstallOptions()
	in.Release = iopts.release
	in.Target = iopts.environment

	if iopts.projectID != "" {
		in.Target = iopts.projectID
		in.TargetType = config.TargetProjects
	}

	if err := h.Install(in); err != nil {
		log.Fatalf("Failed to install release, %v", err)
	}
}
