package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePattern(t *testing.T) {
	cases := []struct {
		p      string
		expect interface{} // error or Pattern
	}{
		{
			p:      ``,
			expect: fmt.Errorf("no pattern specified"),
		},
		{
			p:      `pkg`,
			expect: fmt.Errorf("invalid pattern specified"),
		},
		{
			p: `pkg:ident`,
			expect: Pattern{
				pkg:   regexp.MustCompile(`^pkg$`),
				ident: regexp.MustCompile(`^ident$`),
			},
		},
		{
			p: `pkg:ident:field`,
			expect: Pattern{
				pkg:   regexp.MustCompile(`^pkg$`),
				ident: regexp.MustCompile(`^ident$`),
				field: regexp.MustCompile(`^field$`),
			},
		},
		{
			p: `pkg:ident:method()`,
			expect: Pattern{
				pkg:    regexp.MustCompile(`^pkg$`),
				ident:  regexp.MustCompile(`^ident$`),
				method: regexp.MustCompile(`^method$`),
			},
		},
	}

	for idx, c := range cases {
		out, err := parsePattern(c.p)
		switch c := c.expect.(type) {
		case error:
			if err == nil {
				require.Fail(t, "expect error", idx)
			}
			if !strings.Contains(err.Error(), c.Error()) {
				require.Contains(t, err.Error(), c.Error(), idx)
			}
		case Pattern:
			if (out.pkg == nil) != (c.pkg == nil) {
				require.Fail(t, fmt.Sprintf("package pattern exists=%t", out.pkg == nil), idx)
			}
			if c.pkg != nil {
				require.Equal(t, c.pkg.String(), out.pkg.String(), idx)
			}
			if (out.ident == nil) != (c.ident == nil) {
				require.Fail(t, fmt.Sprintf("ident pattern exists=%t", out.ident == nil), idx)
			}
			if c.ident != nil {
				require.Equal(t, c.ident.String(), out.ident.String(), idx)
			}
			if (out.field == nil) != (c.field == nil) {
				require.Fail(t, fmt.Sprintf("field pattern exists=%t", out.field == nil), idx)
			}
			if c.field != nil {
				require.Equal(t, c.field.String(), out.field.String(), idx)
			}
			if (out.method == nil) != (c.method == nil) {
				require.Fail(t, fmt.Sprintf("method pattern exists=%t", out.method == nil), idx)
			}
			if c.method != nil {
				require.Equal(t, c.method.String(), out.method.String(), idx)
			}
		}
	}
}

func TestFindDepInPackages(t *testing.T) {
	root, _ := os.Getwd()
	cases := []struct {
		name   string
		pwd    string
		p      string
		gopkgs []string
		expect string
	}{
		{
			name:   "Find package type",
			pwd:    "./testdata/usepkg1/pkgtype",
			p:      "uut/pkg1:T1",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1:
	%s/testdata/usepkg1/pkgtype/pkgtype.go:8:11
`, root),
		},
		{
			name:   "Find package func",
			pwd:    "./testdata/usepkg1/pkgfunc",
			p:      "uut/pkg1:F1",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 F1:
	%s/testdata/usepkg1/pkgfunc/pkgfunc.go:8:7
`, root),
		},
		{
			name:   "Find regular struct field",
			pwd:    "./testdata/usepkg1/typefield",
			p:      "uut/pkg1:T1:T1F1",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.T1F1:
	%s/testdata/usepkg1/typefield/typefield.go:10:6
`, root),
		},
		{
			name:   "Find struct embedded field",
			pwd:    "./testdata/usepkg1/typefield",
			p:      "uut/pkg1:T1:T2",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.T2:
	%s/testdata/usepkg1/typefield/typefield.go:12:6
`, root),
		},
		{
			name:   "Find struct field that is a field of an embedded field",
			pwd:    "./testdata/usepkg1/typefield",
			p:      "uut/pkg1:T1:T2F1",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.T2F1:
	%s/testdata/usepkg1/typefield/typefield.go:11:6
`, root),
		},
		{
			name:   "Find field of an embedded struct that is not explicitly referenced should return nothing",
			pwd:    "./testdata/usepkg1/typefield",
			p:      "uut/pkg1:T2:T2F1",
			gopkgs: []string{"."},
			expect: "",
		},
		{
			name:   "Find method on a type, for both the type or a pointer to the type",
			pwd:    "./testdata/usepkg1/typemethod",
			p:      "uut/pkg1:T1:F1()",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.F1():
	%s/testdata/usepkg1/typemethod/typemethod.go:9:2
	%s/testdata/usepkg1/typemethod/typemethod.go:13:2
`, root, root),
		},
		{
			name:   "Find method on the pointer to a type, for both the type or a pointer to the type",
			pwd:    "./testdata/usepkg1/typemethod",
			p:      "uut/pkg1:T1:F2()",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.F2():
	%s/testdata/usepkg1/typemethod/typemethod.go:10:2
	%s/testdata/usepkg1/typemethod/typemethod.go:14:2
`, root, root),
		},
		{
			name:   "Find a method across packages",
			pwd:    "./testdata/crosspkgs",
			p:      "uut/pkg.*:T1:F1()",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.F1():
	%s/testdata/crosspkgs/crosspkgs.go:11:2
uut/pkg2 T1.F1():
	%s/testdata/crosspkgs/crosspkgs.go:12:2
`, root, root),
		},
		{
			name:   "Find method patterns (F.*) across packages",
			pwd:    "./testdata/crosspkgs",
			p:      "uut/pkg.*:T1:F.*()",
			gopkgs: []string{"."},
			expect: fmt.Sprintf(`uut/pkg1 T1.F1():
	%s/testdata/crosspkgs/crosspkgs.go:11:2
uut/pkg2 T1.F1():
	%s/testdata/crosspkgs/crosspkgs.go:12:2
uut/pkg2 T1.F2():
	%s/testdata/crosspkgs/crosspkgs.go:13:2
`, root, root, root),
		},
	}

	for _, c := range cases {
		os.Chdir(c.pwd)
		pattern, err := parsePattern(c.p)
		require.NoError(t, err, c.name)
		matches, err := pattern.FindDepInPackages(c.gopkgs)
		require.NoError(t, err, c.name)
		require.Equal(t, c.expect, matches.String(), c.name)
		os.Chdir(root)
	}
}
