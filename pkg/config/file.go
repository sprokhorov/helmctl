package config

import (
    "fmt"
    "path/filepath"

    "github.com/imdario/mergo"
    "github.com/mitchellh/mapstructure"
    "github.com/sirupsen/logrus"
)

// NewConfigFromFile returns new config.
func NewConfigFromFile(file string, schema string, logger *logrus.Logger, dryRun bool) Config {
    if logger == nil {
        logger = logrus.New()
    }

    // Define package global variable to use it in custom Unmarshaller
    ConfigFilePath = filepath.Dir(file)
    DryRun = dryRun

    return &File{
        configFile: file,
        schemaPath: schema,
        l:          logger,
    }
}

// File is Config implementation.
type File struct {
    Spec struct {
        Repositories []*Repository
        Releases     []*Release
        Installs     Installs
    }

    l          *logrus.Logger
    configFile string
    schemaPath string
}

// Load loads config from file.
func (cf *File) Load() (err error) {
    cf.l.Debugf("config.*Config.LoadFromFile file=%s *Config=%+v", cf.configFile, *cf)
    cf.l.Debugf("config.*Config.LoadFromFile: Unmarshal file content %s", cf.configFile)

    var schemaObject map[string]interface{}

    schemaObject, err = parseConfigFile(cf.configFile, cf.schemaPath, cf.l)
    if err != nil {
        return fmt.Errorf("%s: %v", cf.configFile, err)
    }

    spec, ok := schemaObject["spec"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("Unexpected internal error with %s: %v", cf.configFile, err)
    }

    if repos, ok := spec["repositories"].([]interface{}); ok {
        // Decoding parsed file into Go struct
        err = mapstructure.Decode(repos, &cf.Spec.Repositories)
        if err != nil {
            return fmt.Errorf("%s: %v", cf.configFile, err)
        }
    } else {
        // If releases is not intialized generate empty
        cf.setDefaults()
    }

    counter := map[string]int{}
    for _, repo := range cf.Spec.Repositories {
        counter[repo.Name] += 1
    }
    for repoName, count := range counter {
        if count > 1 {
            return fmt.Errorf("Duplicated repo %s", repoName)
        }
    }

    if releases, ok := spec["releases"].([]interface{}); ok {
        err = mapstructure.Decode(releases, &cf.Spec.Releases)
        if err != nil {
            return fmt.Errorf("%s: %v", cf.configFile, err)
        }
    }

    counter = map[string]int{}
    for _, release := range cf.Spec.Releases {
        counter[release.Name] += 1
    }
    for releaseName, count := range counter {
        if count > 1 {
            return fmt.Errorf("Duplicated release %s", releaseName)
        }
    }

    if installs, ok := spec["installs"].(map[string]interface{}); ok {

        // Process Environments
        if environments, ok := installs["environments"].(map[string]interface{}); ok {
            cf.Spec.Installs.Environments = make(map[string][]*Environment)

            var newEnvironmentValues map[string][]interface{}
            err = mapstructure.Decode(environments, &newEnvironmentValues)
            if err != nil {
                return fmt.Errorf("%s: %v", cf.configFile, err)
            }

            // Block Start
            // This Block provides additional mapping for field which can be in two possible types
            // string and map[string]interface{}
            // This looks like a boilerplate code and can be possibly rewrited
            for environmentName, environmentValues := range newEnvironmentValues {

                cf.Spec.Installs.Environments[environmentName] = []*Environment{}

                for _, environmentValue := range environmentValues {
                    switch environment := environmentValue.(type) {
                    case string:
                        {
                            newEnvironment := NewEnvironmentSimple(environment)
                            cf.Spec.Installs.Environments[environmentName] = append(
                                cf.Spec.Installs.Environments[environmentName], &newEnvironment)
                            break
                        }
                    case map[string]interface{}:
                        {
                            if name, ok := environment["name"]; !ok {
                                cf.l.Errorf("Cannot find name for environment: %v\n", environment)
                            } else {
                                newEnvironment := NewEnvironmentComplex(name.(string))
                                err = newEnvironment.SetValues(environment)
                                if err != nil {
                                    return fmt.Errorf("Cannot parse env additional variables as Release: %s: %v", cf.configFile, err)
                                }

                                cf.Spec.Installs.Environments[environmentName] = append(
                                    cf.Spec.Installs.Environments[environmentName], &newEnvironment)
                            }
                            break
                        }
                    }
                }
            }
            // Block End
        }

        // Process Projects
        if projects, ok := installs["projects"].(map[string]interface{}); ok {
            cf.Spec.Installs.Projects = make(map[string][]*Project)

            var newProjectValues map[string][]interface{}
            err = mapstructure.Decode(projects, &newProjectValues)
            if err != nil {
                return fmt.Errorf("%s: %v", cf.configFile, err)
            }

            // Block Start
            // This Block provides additional mapping for field which can be in two possible types
            // string and map[string]interface{}
            // This looks like a boilerplate code and can be possibly rewrited
            for projectName, projectValues := range newProjectValues {

                cf.Spec.Installs.Projects[projectName] = []*Project{}

                for _, projectValue := range projectValues {
                    switch project := projectValue.(type) {
                    case string:
                        {
                            newProject := NewProjectSimple(project)
                            cf.Spec.Installs.Projects[projectName] = append(
                                cf.Spec.Installs.Projects[projectName], &newProject)
                            break
                        }
                    case map[string]interface{}:
                        {
                            if name, ok := project["name"]; !ok {
                                cf.l.Errorf("Cannot find name for project: %v\n", project)
                            } else {
                                newProject := NewProjectComplex(name.(string))
                                newProject.SetValues(project)
                                if err != nil {
                                    return fmt.Errorf("Cannot parse project additional variables as Release: %s: %v", cf.configFile, err)
                                }
                                cf.Spec.Installs.Projects[projectName] = append(
                                    cf.Spec.Installs.Projects[projectName], &newProject)
                            }
                            break
                        }
                    }
                }
            }
            // Block End
        }

    }
    err = cf.Spec.Installs.Validate()
    if err != nil {
        return err
    }

    // prepare releases
    for _, r := range cf.Spec.Releases {
        r.pathUpdate()

        if err := r.checkScripts(); err != nil {
            return fmt.Errorf("invalid script, %v", err)
        }

        r.setDefaults()
    }

    if err := cf.checkInstallations(); err != nil {
        return err
    }

    return nil
}

// Repositories returns list of defined repositories.
func (cf *File) Repositories() []*Repository {
    return cf.Spec.Repositories
}

// Releases returns list defined of releases.
func (cf *File) Releases() []*Release {
    return cf.Spec.Releases
}

// Merge release with additional params
func (cf *File) mergeReleaseParams(targetRelease *Release, targetName string, targetType TargetType) error {
    switch targetType {
    case TargetProjects:
        {
            for _, release := range cf.Spec.Installs.Projects[targetName] {
                if (*release).GetName() == (*targetRelease).Name {
                    values := (*release).GetValues()
                    if err := mergo.Merge(
                        targetRelease, values,
                        mergo.WithAppendSlice,
                        mergo.WithOverride); err != nil {
                        return fmt.Errorf("Unexpected internal error during merging project params: %v", err)
                    }
                    return nil
                }
            }
            break
        }
    case TargetEnvironments:
        {
            for _, release := range cf.Spec.Installs.Environments[targetName] {
                if (*release).GetName() == (*targetRelease).Name {
                    values := (*release).GetValues()
                    if err := mergo.Merge(
                        targetRelease, values,
                        mergo.WithAppendSlice,
                        mergo.WithOverride); err != nil {
                        return fmt.Errorf("Unexpected internal error during merging project params: %v", err)
                    }
                    return nil
                }
            }
            break
        }
    }
    return fmt.Errorf("Release %s is not found for %s - %s", (*targetRelease).Name, targetType, targetName)
}

// ReleaseGet returns release object.
func (cf *File) ReleaseGet(name string) (*Release, error) {
    for _, r := range cf.Spec.Releases {
        if r.Name == name {
            return r, nil
        }
    }
    return nil, fmt.Errorf("release %s not found", name)
}

// TargetRelease returns releases associated to the target
func (cf *File) TargetRelease(name string, target string, targetType TargetType) (*Release, error) {
    for _, r := range cf.Spec.Releases {
        if r.Name == name {
            if err := cf.mergeReleaseParams(r, target, targetType); err != nil {
                return r, err
            }
            r.setDefaults()
            return r, nil
        }
    }
    return nil, fmt.Errorf("release %s not found", name)
}

// TargetReleases returns releases associated to the target.
func (cf *File) TargetReleases(target string, targetType TargetType) ([]*Release, error) {
    switch targetType {
    case TargetProjects:
        {
            targets := cf.Spec.Installs.Projects
            if _, exists := targets[target]; !exists {
                return []*Release{}, fmt.Errorf("unknown target %s", target)
            }

            // create index
            idx := make(map[string]struct{}, len(targets[target]))
            for _, r := range targets[target] {
                idx[(*r).GetName()] = struct{}{}
            }

            i := 0
            releases := make([]*Release, len(targets[target]))

            for _, r := range cf.Spec.Releases {
                if _, indexed := idx[r.Name]; indexed {
                    if err := cf.mergeReleaseParams(r, target, targetType); err != nil {
                        return []*Release{}, fmt.Errorf("Unexpected internal error during merging project params: %v", err)
                    }
                    r.setDefaults()
                    releases[i] = r
                    i++
                }
            }
            return releases, nil
        }

    case TargetEnvironments:
        {
            targets := cf.Spec.Installs.Environments
            if _, exists := targets[target]; !exists {
                return []*Release{}, fmt.Errorf("unknown target %s", target)
            }

            // create index
            idx := make(map[string]struct{}, len(targets[target]))
            for _, r := range targets[target] {
                idx[(*r).GetName()] = struct{}{}
            }

            i := 0
            releases := make([]*Release, len(targets[target]))

            for _, r := range cf.Spec.Releases {
                if _, indexed := idx[r.Name]; indexed {
                    if err := cf.mergeReleaseParams(r, target, targetType); err != nil {
                        return []*Release{}, fmt.Errorf("Unexpected internal error during merging environment params: %v", err)
                    }
                    r.setDefaults()
                    releases[i] = r
                    i++
                }
            }

            return releases, nil

        }
    default:
        return []*Release{}, fmt.Errorf("unknown target type %s", targetType)
    }
}

func (cf *File) checkInstallations() error {
    for env, releases := range cf.Spec.Installs.Environments {
        for _, r := range releases {
            if _, err := cf.ReleaseGet((*r).GetName()); err != nil {
                return fmt.Errorf("environment %s: %v", env, err)
            }
        }
    }
    for prj, releases := range cf.Spec.Installs.Projects {
        for _, r := range releases {
            if _, err := cf.ReleaseGet((*r).GetName()); err != nil {
                return fmt.Errorf("project %s: %v", prj, err)
            }
        }
    }
    return nil
}

func (cf *File) setDefaults() {
    if cf.Spec.Repositories == nil {
        cf.Spec.Repositories = []*Repository{}
    }
}

func (cf *File) Environments() []string {
    envs := []string{}
    for env, _ := range cf.Spec.Installs.Environments {
        envs = append(envs, env)
    }
    return envs
}

func (cf *File) Projects() []string {
    projects := []string{}
    for project, _ := range cf.Spec.Installs.Projects {
        projects = append(projects, project)
    }
    return projects
}
