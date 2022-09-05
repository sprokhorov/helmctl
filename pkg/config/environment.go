package config

import (
    "fmt"

    "github.com/mitchellh/mapstructure"
)

// Environment interface type is used to hide different kind of envs
type Environment interface {
    GetName() string
    GetValues() *Release
    SetValues(values map[string]interface{}) error
}

// EnvironmentSimple simple environment type
type EnvironmentSimple string

// GetName return a name of env
func (e *EnvironmentSimple) GetName() string {
    return string(*e)
}

// GetValues returns empty values
func (e *EnvironmentSimple) GetValues() *Release {
    return &Release{}
}

// SetValues does nothing
func (e *EnvironmentSimple) SetValues(values map[string]interface{}) error {
    return nil
}

// NewEnvironmentSimple returns new simple env
func NewEnvironmentSimple(name string) Environment {
    newEnv := EnvironmentSimple(name)
    return &newEnv
}

// EnvironmentComplex complex env type
type EnvironmentComplex struct {
    r Release
}

// GetName returns a name of env
func (e *EnvironmentComplex) GetName() string {
    return e.r.Name
}

// GetValues returns values
func (e *EnvironmentComplex) GetValues() *Release {
    return &(e.r)
}

// SetValues set values
func (e *EnvironmentComplex) SetValues(values map[string]interface{}) error {
    err := mapstructure.Decode(values, &e.r)
    if err != nil {
        return fmt.Errorf("Cannot decode environment into Release: %v", err)
    }
    e.r.pathUpdate()
    if err = e.r.checkScripts(); err != nil {
        return fmt.Errorf("Non valid paths for files in  environment: %v", err)
    }
    return nil
}

// NewEnvironmentComplex return new complex env with empty values
func NewEnvironmentComplex(name string) Environment {
    newEnv := EnvironmentComplex{
        r: Release{Name: name},
    }
    return &newEnv
}
