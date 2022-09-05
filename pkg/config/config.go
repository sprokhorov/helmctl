/*
 Caveat:

 * In this package we have global variable to define loading path.

 * It means that we cannot have two separate non-loaded config files from this module.

 CleanUp Interface

 These functions serve to just one purpose: Make YAML be compatible with JSON.

 In JSON we can have map keys with only one type
  string -> (map[string]interface{})
 but in YAML we can use other types. So we have a problem: YAML decoder can convert some nested map
 as map[interface{}]interface{} and this type cannot be validated with JSON.
 So we need to convert all map[interface{}]interface{} -> map[string]interface{}.

 Additional info about this support from gopkg.in/yaml.v3:
 https://github.com/go-yaml/yaml/issues/139.

 Environments

 There are two kind of environments:

 * simple - provided as string

 * complex - provided as struct wih ability to override options

 We use interface to hide implemenetation.

 Projects

 There are two kind of projects:

 * simple - provided as string

 * complex - provided as struct wih ability to override options

 We use interface to hide implemenetation.

 JSON Schema

 Support JSON schema validation and data parsing from different formats with Conflate pkg
*/
package config

// Config defines helmctl configuration.
type Config interface {
    Load() error
    Repositories() []*Repository
    Releases() []*Release
    TargetRelease(name string, target string, targetType TargetType) (*Release, error)
    TargetReleases(target string, targetType TargetType) ([]*Release, error)
    Environments() []string
    Projects() []string
}

// TargetType is a helm target type.
type TargetType string

// Define target types
const (
    TargetEnvironments TargetType = "environments"
    TargetProjects     TargetType = "projects"
)

// Define global variable
var ConfigFilePath string = ""
var DryRun bool = false
