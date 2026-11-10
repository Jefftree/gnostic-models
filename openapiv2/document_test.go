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

package openapi_v2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/gnostic-models/compiler"
)

func TestParseAndResolveEdgeCases(t *testing.T) {
	for _, in := range [][]byte{nil, {}, []byte(" \n\t "), []byte("# c\n"), []byte("-"), []byte("[1]"), []byte("[a,b,c]"), []byte("[a,b]"), []byte("hello"), []byte("null"), []byte("swagger: '2.0'\ninfo: {title: x, version: '1'}\npaths: {}\ndefinitions: [a,b,c]\n")} {
		if _, err := ParseDocument(in); err == nil {
			t.Fatalf("ParseDocument(%q) succeeded", in)
		}
	}
	if doc, err := ParseDocument([]byte("swagger: '2.0'\ninfo: {title: '', version: ''}\npaths:\n")); err != nil || doc == nil {
		t.Fatalf("empty paths failed: (%v, %v)", doc, err)
	}
	p := filepath.Join(t.TempDir(), "spec.yaml")
	hdr := "swagger: '2.0'\ninfo: {title: x, version: '1'}\n"
	for i, tc := range []struct{ spec, want string }{
		{hdr + "paths: {}\ndefinitions:\n  A: {$ref: '#/definitions/A'}\n", "circular $ref"},
		{hdr + "paths: {}\ndefinitions:\n  A: {$ref: '#/definitions/B'}\n  B: {$ref: '#/definitions/A'}\n", "circular $ref"},
		{hdr + "paths: {}\ndefinitions:\n  Node: {type: object, properties: {next: {$ref: '#/definitions/Node'}}}\n", "circular $ref"},
		{hdr + "x-p: {$ref: '#/x-p'}\npaths:\n  /i:\n    get:\n      parameters: [{$ref: '#/x-p'}]\n      responses: {'200': {description: ok}}\n", "circular $ref"},
		{hdr + "x-pi:\n  a: {$ref: '#/x-pi/b'}\n  b: {$ref: '#/x-pi/a'}\npaths:\n  /i: {$ref: '#/x-pi/a'}\n", "circular $ref"},
		{hdr + "paths: {}\ndefinitions:\n  L: {type: string}\n  B: {type: object, properties: {a: {$ref: '#/definitions/L'}, b: {$ref: '#/definitions/L'}}}\n  R: {type: object, properties: {x: {$ref: '#/definitions/B'}, y: {$ref: '#/definitions/B'}}}\n", ""},
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
