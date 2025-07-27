package bencode

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    interface{}
		wantErr bool
	}{
		{"string", "4:spam", "spam", false},
		{"integer", "i42e", int64(42), false},
		{"negative integer", "i-42e", int64(-42), false},
		{"list", "l4:spami42ee", []interface{}{"spam", int64(42)}, false},
		{"dictionary", "d3:key5:valuee", map[string]interface{}{"key": "value"}, false},
		{
			"complex dictionary",
			"d4:infod6:lengthi1024e4:name8:test.txtee",
			map[string]interface{}{
				"info": map[string]interface{}{
					"length": int64(1024),
					"name":   "test.txt",
				},
			},
			false,
		},
		{"empty string", "0:", "", false},
		{"invalid integer", "ie", nil, true},
		{"unterminated list", "l4:spam", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := Unmarshal(r)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    interface{}
		wantErr bool
	}{
		{
			"simple dict",
			map[string]interface{}{"ut_metadata": 1, "p": 6881},
			// The order of keys in a map is not guaranteed
			[]string{"d1:pi6881e11:ut_metadatai1ee", "d11:ut_metadatai1e1:pi6881ee"},
			false,
		},
		{
			"nested dict",
			map[string]interface{}{"m": map[string]interface{}{"ut_metadata": 1}},
			"d1:md11:ut_metadatai1eee",
			false,
		},
		{"unsupported type", "hello", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Marshal(&buf, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := buf.String()
				if want, ok := tt.want.(string); ok { // Single possibility
					if got != want {
						t.Errorf("Marshal() got = %q, want %q", got, want)
					}
				} else if want, ok := tt.want.([]string); ok { // Multiple possibilities
					found := false
					for _, w := range want {
						if got == w {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Marshal() got = %q, want one of %v", got, want)
					}
				}
			}
		})
	}
}
