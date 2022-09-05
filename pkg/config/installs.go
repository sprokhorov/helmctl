package config

import (
    "fmt"
)

// Installs maps releases with environments and projects.
type Installs struct {
    Environments map[string][]*Environment
    Projects     map[string][]*Project
}

// Validate that all parts of Installs do not have duplicates
func (i *Installs) Validate() error {
    for _, envValues := range i.Environments {
        uniq := make(map[string]int)
        for _, envVal := range envValues {
            uniq[(*envVal).GetName()] += 1
        }
        for key, values := range uniq {
            if values > 1 {
                return fmt.Errorf("Duplicate Environment component name: %s\n", key)
            }
        }
    }

    for _, prjValues := range i.Projects {
        uniq := make(map[string]int)
        for _, prjVal := range prjValues {
            uniq[(*prjVal).GetName()] += 1
        }
        for key, values := range uniq {
            if values > 1 {
                return fmt.Errorf("Duplicate Project component name: %s\n", key)
            }
        }
    }
    return nil
}
