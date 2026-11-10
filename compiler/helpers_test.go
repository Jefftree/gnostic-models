// Copyright 2026 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compiler

import (
	"os"
	"path/filepath"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestUnpackMapAndReadInfoForRef(t *testing.T) {
	doc, _ := ReadInfoFromBytes("", []byte("openapi: '3.0.0'\n"))
	if m, ok := UnpackMap(doc); !ok || m == nil || m.Kind != yaml.MappingNode {
		t.Fatalf("UnpackMap(DocumentNode) = (%v, %v)", m, ok)
	}
	if m, ok := UnpackMap(&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Line: 6, Column: 7}); !ok || m.Kind != yaml.MappingNode || len(m.Content) != 0 || m.Line != 6 || m.Column != 7 {
		t.Fatalf("UnpackMap(!!null) = (%+v, %v)", m, ok)
	}
	bad := []*yaml.Node{nil, {Kind: yaml.DocumentNode}, {Kind: yaml.MappingNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Value: "k"}}}}
	for _, raw := range []string{"[1]", "[a,b,c]", "[a,b]", "-", "hello"} {
		n, _ := ReadInfoFromBytes("", []byte(raw))
		bad = append(bad, n, n.Content[0])
	}
	for i, n := range bad {
		if m, ok := UnpackMap(n); ok || m != nil {
			t.Fatalf("case %d: UnpackMap = (%v, %v)", i, m, ok)
		}
	}
	spec := filepath.Join(t.TempDir(), "spec.yaml")
	_ = os.WriteFile(spec, []byte("components:\n  schemas:\n    B: {type: object, required: [x, y, z]}\n    E: {type: object, required: [x, y]}\n    L: {type: string}\n"), 0644)
	_ = os.WriteFile(filepath.Join(filepath.Dir(spec), "empty.yaml"), []byte("# comment\n"), 0644)
	for _, ref := range []string{"#/components/schemas/B/required/z", "#/components/schemas/E/required/z", "#/components/schemas/B/required/0", "#/components/schemas/L/type/sub", "#bad", "#/a#b", "empty.yaml#/x", "bad%zz.yaml#/x"} {
		ClearCaches()
		if n, err := ReadInfoForRef(spec, ref); err == nil || n != nil {
			t.Fatalf("ReadInfoForRef(%q) = (%v, %v)", ref, n, err)
		}
	}
	EnableInfoCache()
	ClearCaches()
	defer ClearCaches()
	if n, err := ReadInfoForRef(spec, "#/components/schemas/L"); err != nil || n == nil || n.Kind != yaml.MappingNode {
		t.Fatalf("valid ref = (%v, %v)", n, err)
	}
	if _, err := ReadInfoForRef(spec, "#/missing"); err == nil {
		t.Fatal("expected error on first unresolved ref")
	} else if n, err := ReadInfoForRef(spec, "#/missing"); err != nil || n != nil {
		t.Fatalf("expected cached (nil, nil), got (%v, %v)", n, err)
	}
}
