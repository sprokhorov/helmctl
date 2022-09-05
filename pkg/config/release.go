package config

import (
    "os"
    "path/filepath"
)

// Release represents helm release with values.
type Release struct {
    Name          string
    Chart         string
    Version       string
    Namespace     string
    BeforeScripts []*string
    AfterScripts  []*string
    Atomic        *bool
    Repository    *Repository
    Values        []*Value
    ValueFiles    []*ValueFile

    IncludePath string
}

// pathAppend appends path prefix to scripts and value-files.
func (r *Release) pathUpdate() {
    for idx, s := range r.BeforeScripts {
        realPath := filepath.Join(ConfigFilePath, filepath.Dir(r.IncludePath), *s)
        r.BeforeScripts[idx] = &realPath
    }
    for idx, s := range r.AfterScripts {
        realPath := filepath.Join(ConfigFilePath, filepath.Dir(r.IncludePath), *s)
        r.AfterScripts[idx] = &realPath
    }
    for idx, vf := range r.ValueFiles {
        realPath := filepath.Join(ConfigFilePath, filepath.Dir(r.IncludePath), vf.Name)
        r.ValueFiles[idx].Name = realPath
    }
}

// Check file existence
func fileIsExists(file *string) error {
    _, err := os.Stat(*file)
    return err
}

// CheckScripts checks if defined script files exists.
func (r *Release) checkScripts() error {
    // check before scripts
    for _, s := range r.BeforeScripts {
        if err := fileIsExists(s); err != nil {
            return err
        }
    }

    // check after scripts
    for _, s := range r.AfterScripts {
        if err := fileIsExists(s); err != nil {
            return err
        }
    }

    for _, vf := range r.ValueFiles {
        if err := fileIsExists(&vf.Name); err != nil {
            return err
        }
    }

    return nil
}

func (r *Release) setDefaults() {
    // value files
    if r.ValueFiles == nil {
        r.ValueFiles = []*ValueFile{}
    }
    for _, vf := range r.ValueFiles {
        vf.setDefaults()
    }
    // values
    if r.Values == nil {
        r.Values = []*Value{}
    }

    if r.Atomic == nil {
        f := false
        r.Atomic = &f
    }
    if r.AfterScripts == nil {
        r.AfterScripts = []*string{}
    }
    if r.BeforeScripts == nil {
        r.BeforeScripts = []*string{}
    }
    if r.Namespace == "" {
        r.Namespace = r.Name
    }
    if r.Repository == nil {
        r.Repository = &Repository{}
    }
}
