package config

import (
    "os"

    "github.com/miracl/conflate"
    "github.com/sirupsen/logrus"
)

// initConflate add custom unmarshaller for YAML To conflate
func initConflate() {
    // define the unmarshallers for the given file extensions, blank extension is the global unmarshaller
    conflate.Unmarshallers = conflate.UnmarshallerMap{
        ".json": {conflate.JSONUnmarshal},
        ".jsn":  {conflate.JSONUnmarshal},
        ".yaml": {customYAMLUnmarshal},
        ".yml":  {customYAMLUnmarshal},
        "":      {conflate.JSONUnmarshal, customYAMLUnmarshal, conflate.TOMLUnmarshal},
    }
}

// parseConfigFile return parsed config file
func parseConfigFile(path string, schemaPath string, logger *logrus.Logger) (config map[string]interface{}, err error) {
    initConflate()

    // read file into conflate structure
    cft, err := conflate.FromFiles(path)
    if err != nil {
        return config, err
    }

    var schema *conflate.Schema

    if _, err := os.Stat(schemaPath); !os.IsNotExist(err) {
        logger.Debugf("Parsing schema from file: %s", schemaPath)
        schema, err = getSchemaFromFile(schemaPath)
        if err != nil {
            return config, err
        }
    } else {
        logger.Debugln("Parsing existed schema")
        schema, err = getExistedSchema()
        if err != nil {
            return config, err
        }
    }

    // apply defaults defined in schema to merged data
    err = cft.ApplyDefaults(schema)
    if err != nil {
        return config, err
    }

    // validate merged data against schema
    err = cft.Validate(schema)
    if err != nil {
        return config, err
    }

    err = cft.Unmarshal(&config)
    if err != nil {
        return config, err
    }

    return config, nil
}

// getSchemaFromFile returns parsed schema from file
func getSchemaFromFile(path string) (*conflate.Schema, error) {
    // load a json schema
    schema, err := conflate.NewSchemaFile(path)
    if err != nil {
        return nil, err
    }
    return schema, err
}

// getExistedSchema returns parsed schema from constant value - ConflateJSONSchema
func getExistedSchema() (*conflate.Schema, error) {
    // load a json schema
    schema, err := conflate.NewSchemaData([]byte(ConflateJSONSchema))
    if err != nil {
        return nil, err
    }
    return schema, err
}

const (
    // Default Schema for helmctl file in JSON format
    ConflateJSONSchema string = `
{
    "title": "helmctl_config",
    "type": "object",
    "default": {},

    "definitions": {
        "repository": {
            "type": "object",
            "properties": {
                    "name": {"type": "string"},
                    "url": {"type": "string"},
                    "user": {"type": "string"},
                    "password": {"type": "string"}
                },
            "additionalProperties": false,
            "required": ["name", "url"]
        },

        "value": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "value": {"oneOf": [
                    {"type": "string"},
                    {"type": "number"},
                    {"type": "boolean"}
                ]},
                "type": {"type": "string"}
            },
            "additionalProperties": false,
            "required": ["name", "value"]
        },

        "valueFile": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "decrypt": {"type": "boolean"}
            },
            "additionalProperties": false,
            "required": ["name"]
        },

        "release": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "chart": {"type": "string"},
                "version": {"type": "string"},
                "namespace": {"type": "string"},
                "beforeScripts": {"type": "array", "items": {"type": "string"}},
                "afterScripts": {"type": "array", "items": {"type": "string"}},
                "atomic": {"type": "boolean"},
                "repository": { "$ref": "#/definitions/repository" },
                "values": {"type": "array", "items": {"$ref": "#/definitions/value"}},
                "valueFiles": {"type": "array", "items": {"$ref": "#/definitions/valueFile"}},
                "IncludePath": {"type": "string"}
            },
            "additionalProperties": false,
            "required": ["name", "chart"]
        },

        "releaseNonStrict": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "chart": {"type": "string"},
                "version": {"type": "string"},
                "include": {"type": "string"},
                "namespace": {"type": "string"},
                "beforeScripts": {"type": "array", "items": {"type": "string"}},
                "afterScripts": {"type": "array", "items": {"type": "string"}},
                "atomic": {"type": "boolean"},
                "repository": { "$ref": "#/definitions/repository" },
                "values": {"type": "array", "items": {"$ref": "#/definitions/value"}},
                "valueFiles": {"type": "array", "items": {"$ref": "#/definitions/valueFile"}},
                "IncludePath": {"type": "string"}
            },
            "additionalProperties": false,
            "required": ["name"]
        },

        "customMap": {
            "type": "object",
            "patternProperties": {
                "^.*$": {
                    "type": "array",
                    "items": {"$ref": "#/definitions/customMapObject"}
                }
            },
            "properties": {},
            "additionalProperties": false
        },

        "customMapObject": {
            "oneOf": [
                {"type": "string"},
                {"$ref": "#/definitions/releaseNonStrict"}
            ]
        },


        "project": {
            "oneOf": [
                {"type": "string"},
                {"type": "object", "properties": {}, "additionalProperties": true}
            ]
        }

    },

    "properties": {

        "version": {
            "type": "string"
        },

        "spec": {
            "type": "object",
            "properties": {
                "repositories": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/repository"
                    },
                    "uniqueItems": true
                },

                "releases": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/release"
                    },
                    "uniqueItems": true
                },

                "installs": {
                    "type": "object",
                    "properties": {
                        "environments": {"$ref": "#/definitions/customMap"},
                        "projects": {"$ref": "#/definitions/customMap"}
                    },
                    "additionalProperties": false
                }
            },
            "required": [
                "releases",
                "installs"
            ],
            "additionalProperties": false
        }
    },

    "required": [
      "version",
      "spec"
    ]
}
`
)
