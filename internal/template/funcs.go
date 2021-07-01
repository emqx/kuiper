package template

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"reflect"
	"strconv"
	"text/template"
)

var FuncMap template.FuncMap

func init() {
	FuncMap = template.FuncMap(sprig.FuncMap())
	FuncMap["json"] = FuncMap["toJson"]
	FuncMap["base64"] = Base64Encode
}

func Base64Encode(para interface{}) (string, error) {
	v := reflect.ValueOf(para)
	if !v.IsValid() {
		return "", fmt.Errorf("based64 error for nil")
	}
	switch v.Kind() {
	case reflect.Bool:
		bv := strconv.FormatBool(v.Bool())
		return base64.StdEncoding.EncodeToString([]byte(bv)), nil
	case reflect.Int, reflect.Int64:
		iv := strconv.FormatInt(v.Int(), 10)
		return base64.StdEncoding.EncodeToString([]byte(iv)), nil
	case reflect.Uint64:
		iv := strconv.FormatUint(v.Uint(), 10)
		return base64.StdEncoding.EncodeToString([]byte(iv)), nil
	case reflect.Float32:
		fv := strconv.FormatFloat(v.Float(), 'f', -1, 32)
		return base64.StdEncoding.EncodeToString([]byte(fv)), nil
	case reflect.Float64:
		fv := strconv.FormatFloat(v.Float(), 'f', -1, 64)
		return base64.StdEncoding.EncodeToString([]byte(fv)), nil
	case reflect.String:
		return base64.StdEncoding.EncodeToString([]byte(v.String())), nil
	case reflect.Map:
		if a, err := json.Marshal(para); err != nil {
			return "", err
		} else {
			en := base64.StdEncoding.EncodeToString(a)
			return en, nil
		}
	default:
		return "", fmt.Errorf("Unsupported data type %s for base64 function.", v.Kind())
	}
}
