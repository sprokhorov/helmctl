package config

import (
    "fmt"

    "github.com/mitchellh/mapstructure"
)

// Project interface
type Project interface {
    GetName() string
    GetValues() *Release
    SetValues(values map[string]interface{}) error
}

// ProjectSimple simple project type
type ProjectSimple string

// GetName returns a name of project
func (e *ProjectSimple) GetName() string {
    return string(*e)
}

// GetValues returns empty values
func (e *ProjectSimple) GetValues() *Release {
    return &Release{}
}

// SetValues does nothing
func (e *ProjectSimple) SetValues(values map[string]interface{}) error {
    return nil
}

// NewProjectSimple returns new simple project
func NewProjectSimple(name string) Project {
    newEnv := ProjectSimple(name)
    return &newEnv
}

// ProjectComplex complex project type
type ProjectComplex struct {
    r Release
}

// GetName returns a name of project
func (e *ProjectComplex) GetName() string {
    return e.r.Name
}

// GetValues returns values
func (e *ProjectComplex) GetValues() *Release {
    return &(e.r)
}

// SetValues set alues for complex project
func (e *ProjectComplex) SetValues(values map[string]interface{}) error {
    err := mapstructure.Decode(values, &e.r)
    if err != nil {
        return fmt.Errorf("Cannot decode project into Release: %v", err)
    }
    e.r.pathUpdate()
    if err = e.r.checkScripts(); err != nil {
        return fmt.Errorf("Non valid paths for files in  project: %v", err)
    }
    return nil
}

// NewProjectComplex returns new complex project with empty values
func NewProjectComplex(name string) Project {
    newEnv := ProjectComplex{
        r: Release{Name: name},
    }
    return &newEnv
}
