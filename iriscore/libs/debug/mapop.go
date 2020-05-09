package debug

import (
	"strconv"
	"strings"
)

func GetValueString(name string, vmap map[string]interface{}) string {
	v, ok := vmap[name]
	if !ok {
		return ""
	}
	return v.(string)
}

func GetValueBoolString(name string, vmap map[string]interface{}) string {
	v, ok := vmap[name]
	if !ok {
		return "false"
	}
	if bv, ok := v.(bool); !ok {
		return "false"
	} else if bv {
		return "true"
	} else {
		return "fale"
	}
}

func GetValueBool(name string, vmap map[string]interface{}) bool {
	v, ok := vmap[name]
	if !ok {
		return false
	}
	if bv, ok := v.(bool); !ok {
		return false
	} else if bv {
		return true
	} else {
		return false
	}
}

func GetValueInt(name string, vmap map[string]interface{}) int {
	v, ok := vmap[name]
	if !ok {
		return 0
	}
	switch v.(type) {
	case int:
		return v.(int)
	case float64:
		return int(v.(float64))
	}
	return 0
}

func GetValueMap(name string, vmap map[string]interface{}) map[string]interface{} {
	v, ok := vmap[name]
	if !ok {
		return nil
	}
	return v.(map[string]interface{})
}

func GetValueArray(name string, vmap map[string]interface{}) []interface{} {
	v, ok := vmap[name]
	if !ok {
		return nil
	}
	return v.([]interface{})
}

func MB(v string) int {
	vv := strings.ToUpper(v)
	//log.Debugf(" ===disk vol %s", vv)
	switch {
	case strings.HasSuffix(vv, "KB"):
		data, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(vv, "KB")), 32)
		return int(data / 1024)
	case strings.HasSuffix(vv, "MB"):
		data, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(vv, "MB")), 32)
		return int(data)
	case strings.HasSuffix(vv, "GB"):
		data, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(vv, "GB")), 32)
		return int(data * 1024)
	case strings.HasSuffix(vv, "TB"):
		data, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(vv, "TB")), 32)
		return int(data * 1024 * 1024)
	case strings.HasSuffix(vv, "PB"):
		data, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(vv, "PB")), 32)
		return int(data * 1024 * 1024 * 1024)
	case strings.HasSuffix(vv, "B"):
		data, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(vv, "B")), 32)
		return int(data / 1024 / 1024)
	default:
		return 0
	}
}
