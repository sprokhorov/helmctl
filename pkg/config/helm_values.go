package config

import (
    "fmt"

    "reflect"
)

// ValueFile represents helm value file.
type ValueFile struct {
    Name    string
    Decrypt *bool
}

func (vf *ValueFile) GetDecrypt() bool {
    if reflect.ValueOf(vf.Decrypt).Kind() == reflect.Ptr {
        return *vf.Decrypt
    }
    return false
}

func (vf *ValueFile) setDefaults() {
    if vf.Decrypt == nil {
        f := false
        vf.Decrypt = &f
    }
}

// Value represents helm value. This value will be setted via --set argument to helm.
type Value struct {
    Name  string
    Value interface{}
    Type  string
}

func (v *Value) GetKeyValuePair() string {
    return fmt.Sprintf("%s=%v", v.Name, v.Value)
}
