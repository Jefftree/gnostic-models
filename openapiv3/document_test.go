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

package openapi_v3

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/gnostic-models/compiler"
)

func TestParseAndResolveEdgeCases(t *testing.T) {
	for _, in := range [][]byte{nil, {}, []byte(" \n\t "), []byte("# c\n"), []byte("-"), []byte("[1]"), []byte("[a,b,c]"), []byte("[a,b]"), []byte("hello"), []byte("null"), []byte("openapi: '3.0.0'\ninfo: {title: x, version: '1'}\npaths: [a,b,c]\n")} {
		if _, err := ParseDocument(in); err == nil {
			t.Fatalf("ParseDocument(%q) succeeded", in)
		}
	}
	if doc, err := ParseDocument([]byte("openapi: '3.0.0'\ninfo:\n  title:\n  version:\npaths:\n")); err != nil || doc == nil {
		t.Fatalf("empty paths failed: (%v, %v)", doc, err)
	}
	p := filepath.Join(t.TempDir(), "spec.yaml")
	hdr := "openapi: '3.0.0'\ninfo: {title: x, version: '1'}\n"
	for i, tc := range []struct{ spec, want string }{
		{hdr + "paths: {}\ncomponents:\n  schemas:\n    A: {$ref: '#/components/schemas/A'}\n", "circular $ref"},
		{hdr + "paths: {}\ncomponents:\n  schemas:\n    A: {$ref: '#/components/schemas/B'}\n    B: {$ref: '#/components/schemas/A'}\n", "circular $ref"},
		{hdr + "paths:\n  /t: {$ref: '#/x-pi/loop'}\nx-pi:\n  loop:\n    post:\n      responses: {}\n      callbacks:\n        cb:\n          '{$url}': {$ref: '#/x-pi/loop'}\n", "circular $ref"},
		{hdr + "paths: {}\ncomponents:\n  schemas:\n    A: {$ref: '#/components/schemas/B/required/z'}\n    B: {type: object, required: [x, y, z]}\n", "could not resolve"},
		{hdr + "paths: {}\ncomponents:\n  schemas:\n    L: {type: string}\n    D: {$ref: '#/components/schemas/L'}\n    R: {type: object, properties: {a: {$ref: '#/components/schemas/D'}, b: {$ref: '#/components/schemas/D'}}}\n", ""},
	} {
		compiler.ClearCaches()
		_ = os.WriteFile(p, []byte(tc.spec), 0644)
		if doc, err := ParseDocument([]byte(tc.spec)); err != nil {
			t.Fatalf("case %d parse failed: %v", i, err)
		} else if _, err = doc.ResolveReferences(p); (tc.want == "") != (err == nil) || (tc.want != "" && !strings.Contains(err.Error(), tc.want)) {
			t.Fatalf("case %d err=%v, want %q", i, err, tc.want)
		}
	}
}
