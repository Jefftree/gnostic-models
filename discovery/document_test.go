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

package discovery_v1

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/gnostic-models/compiler"
)

func TestParseAndResolveEdgeCases(t *testing.T) {
	for _, in := range [][]byte{nil, {}, []byte(" \n\t "), []byte("# c\n"), []byte("-"), []byte("[1]"), []byte("[a,b,c]"), []byte("[a,b]"), []byte("hello"), []byte("null"), []byte("kind: discovery#restDescription\nschemas: [a,b,c]\n")} {
		if _, err := ParseDocument(in); err == nil {
			t.Fatalf("ParseDocument(%q) succeeded", in)
		}
	}
	spec := "kind: discovery#restDescription\ndiscoveryVersion: v1\nschemas:\n  Node: {type: object, properties: {next: {$ref: '#/schemas/Node'}}}\n"
	p := filepath.Join(t.TempDir(), "cycle.yaml")
	_ = os.WriteFile(p, []byte(spec), 0644)
	compiler.ClearCaches()
	if doc, err := ParseDocument([]byte(spec)); err != nil {
		t.Fatal(err)
	} else if _, err := doc.ResolveReferences(p); err == nil || !strings.Contains(err.Error(), "circular $ref") {
		t.Fatalf("expected circular $ref error, got %v", err)
	}
}
