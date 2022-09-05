package config

import "fmt"

// cleanUpInterfaceArray converter for []interface{} type
func cleanUpInterfaceArray(in []interface{}) []interface{} {
    result := make([]interface{}, len(in))
    for i, v := range in {
        result[i] = cleanUpMapValue(v)
    }
    return result
}

// cleanUpInterfaceMap convert map[interface{}]interface{} type
func cleanUpInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    for k, v := range in {
        result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
    }
    return result
}

// cleanUpStrInterfaceMap converter for map[string]interface{} type
// This function is needed due to not stop converting in the middle of the work.
// E.g.:
// map[interface{}]interface{} {
//   map[string]interface{} {
//     map[interface{}]interface{} {}
//   }
// }
//
func cleanUpStrInterfaceMap(in map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    for k, v := range in {
        result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
    }
    return result
}

// cleanUpMapValue type switch for converters
func cleanUpMapValue(v interface{}) interface{} {
    switch v := v.(type) {
    case []interface{}:
        return cleanUpInterfaceArray(v)
    case map[interface{}]interface{}:
        return cleanUpInterfaceMap(v)
    case map[string]interface{}:
        return cleanUpStrInterfaceMap(v)
        // From source, but we need to save boolean type so just return type if it's not an slice or map
        //   case string:
        //       return v
        //   default:
        //       return fmt.Sprintf("%v", v)
    default:
        return v
    }
}
