package xsql

import (
	"encoding/json"
	"fmt"
	"github.com/lf-edge/ekuiper/pkg/ast"
	"reflect"
	"testing"
)

func Test_MessageValTest(t *testing.T) {
	var tests = []struct {
		key     string
		message Message
		exptV   interface{}
		exptOk  bool
	}{
		{
			key: "key1",
			message: Message{
				"key1": "val1",
				"key2": "val2",
			},
			exptV:  "val1",
			exptOk: true,
		},

		{
			key: "key0",
			message: Message{
				"key1": "val1",
				"key2": "val2",
			},
			exptV:  nil,
			exptOk: false,
		},

		{
			key: "key1",
			message: Message{
				"Key1": "val1",
				"key2": "val2",
			},
			exptV:  "val1",
			exptOk: true,
		},

		{
			key: "key1" + ast.COLUMN_SEPARATOR + "subkey",
			message: Message{
				"Key1":   "val1",
				"subkey": "subval",
			},
			exptV:  "subval",
			exptOk: true,
		},

		{
			key: "192.168.0.1",
			message: Message{
				"Key1":        "val1",
				"192.168.0.1": "000",
			},
			exptV:  "000",
			exptOk: true,
		},

		{
			key: "parent" + ast.COLUMN_SEPARATOR + "child",
			message: Message{
				"key1":         "val1",
				"child":        "child_val",
				"parent.child": "demo",
			},
			exptV:  "child_val",
			exptOk: true,
		},

		{
			key: "parent.child",
			message: Message{
				"key1":         "val1",
				"child":        "child_val",
				"parent.child": "demo",
			},
			exptV:  "demo",
			exptOk: true,
		},

		{
			key: "parent.Child",
			message: Message{
				"key1":         "val1",
				"child":        "child_val",
				"parent.child": "demo",
			},
			exptV:  "demo",
			exptOk: true,
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		//fmt.Printf("Parsing SQL %q.\n", tt.s)
		v, ok := tt.message.Value(tt.key)
		if tt.exptOk != ok {
			t.Errorf("%d. error mismatch:\n  exp=%t\n  got=%t\n\n", i, tt.exptOk, ok)
		} else if tt.exptOk && !reflect.DeepEqual(tt.exptV, v) {
			t.Errorf("%d. \n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.exptV, v)
		}
	}
}

func Test_StreamFieldsMarshall(t *testing.T) {
	var tests = []struct {
		sf ast.StreamFields
		r  string
	}{{
		sf: []ast.StreamField{
			{Name: "USERID", FieldType: &ast.BasicType{Type: ast.BIGINT}},
			{Name: "FIRST_NAME", FieldType: &ast.BasicType{Type: ast.STRINGS}},
			{Name: "LAST_NAME", FieldType: &ast.BasicType{Type: ast.STRINGS}},
			{Name: "NICKNAMES", FieldType: &ast.ArrayType{Type: ast.STRINGS}},
			{Name: "Gender", FieldType: &ast.BasicType{Type: ast.BOOLEAN}},
			{Name: "ADDRESS", FieldType: &ast.RecType{
				StreamFields: []ast.StreamField{
					{Name: "STREET_NAME", FieldType: &ast.BasicType{Type: ast.STRINGS}},
					{Name: "NUMBER", FieldType: &ast.BasicType{Type: ast.BIGINT}},
				},
			}},
		},
		r: `[{"FieldType":"bigint","Name":"USERID"},{"FieldType":"string","Name":"FIRST_NAME"},{"FieldType":"string","Name":"LAST_NAME"},{"FieldType":{"Type":"array","ElementType":"string"},"Name":"NICKNAMES"},{"FieldType":"boolean","Name":"Gender"},{"FieldType":{"Type":"struct","Fields":[{"FieldType":"string","Name":"STREET_NAME"},{"FieldType":"bigint","Name":"NUMBER"}]},"Name":"ADDRESS"}]`,
	}, {
		sf: []ast.StreamField{
			{Name: "USERID", FieldType: &ast.BasicType{Type: ast.BIGINT}},
		},
		r: `[{"FieldType":"bigint","Name":"USERID"}]`,
	}}
	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		r, err := json.Marshal(tt.sf)
		if err != nil {
			t.Errorf("%d. \nmarshall error: %v", i, err)
			t.FailNow()
		}
		result := string(r)
		if !reflect.DeepEqual(tt.r, result) {
			t.Errorf("%d. \nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.r, result)
		}
	}
}
