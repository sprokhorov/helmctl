package config

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "reflect"

    "gopkg.in/yaml.v3"
)

// Fragment is used for loading custom variables
type Fragment struct {
    content *yaml.Node
}

// UnmarshalYAML adds additional processor to parser
func (f *Fragment) UnmarshalYAML(value *yaml.Node) (err error) {
    // process custom in fragments
    f.content, err = resolveTags(value)
    return err
}

// CustomProcessor is a constructor struct
type CustomProcessor struct {
    target interface{}
}

// UnmarshalYAML adds additional parser to constructor
func (i *CustomProcessor) UnmarshalYAML(value *yaml.Node) error {
    resolved, err := resolveTags(value)
    if err != nil {
        return err
    }
    err = resolved.Decode(i.target)

    return err
}

// resolveTags provides a logic for parsing custom
func resolveTags(node *yaml.Node) (*yaml.Node, error) {
    switch node.Tag {
    case "!env":
        if node.Kind != yaml.ScalarNode {
            return nil, fmt.Errorf("!env on a non-scalar node")
        }
        envValue, ok := os.LookupEnv(node.Value)

        if !ok && !DryRun {
            return nil, fmt.Errorf("Env variable: %s is not found, line: %d", node.Value, node.Line)
        }
        node.Value = envValue
        return node, nil
    case "!include":
        if node.Kind != yaml.ScalarNode {
            return nil, fmt.Errorf("!env on a non-scalar node, line: %d", node.Line)
        }

        file, err := ioutil.ReadFile(filepath.Join(ConfigFilePath, node.Value))

        if err != nil {
            return nil, fmt.Errorf("Include %s : %v, line: %d", node.Value, err, node.Line)
        }
        var f Fragment
        err = yaml.Unmarshal(file, &f)

        // We can use files with relative pathes in include
        // that is why we need to inject additional pair of key:value
        // "includePath": path
        pathNodeName := yaml.Node{}
        pathNodeName.SetString("IncludePath")

        pathNodeValue := yaml.Node{}
        pathNodeValue.SetString(node.Value)

        f.content.Content = append(f.content.Content, []*yaml.Node{&pathNodeName, &pathNodeValue}...)

        return f.content, err
    }
    if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
        var err error
        for i := range node.Content {
            node.Content[i], err = resolveTags(node.Content[i])
            if err != nil {
                return nil, err
            }
        }
    }
    return node, nil
}

// YAMLUnmarshal defines YAMLUnmarshal to use yaml3 library
func YAMLUnmarshal(data []byte, out interface{}) error {
    err := yaml.Unmarshal(data, out)
    if err != nil {
        return fmt.Errorf("%v. %s", err, "The data could not be unmarshalled as yaml")
    }
    return nil
}

// customYAMLUnmarshal defines YAMLUnmarshal to use yaml3 library
// Custom unmarshaller for YAML
func customYAMLUnmarshal(data []byte, out interface{}) error {
    var result map[string]interface{}

    err := YAMLUnmarshal(data, &CustomProcessor{&result})

    // We need this magic with reflection to override a value by pointer in interface
    rv := reflect.ValueOf(out)
    if rv.Kind() == reflect.Ptr {
        rv = rv.Elem()
    }
    fieldType := rv.Type()

    val := reflect.ValueOf(cleanUpStrInterfaceMap(result))

    rv.Set(val.Convert(fieldType))

    return err
}
