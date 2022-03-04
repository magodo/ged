package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Pattern struct {
	pkg    *regexp.Regexp
	ident  *regexp.Regexp
	method *regexp.Regexp
	field  *regexp.Regexp
}

func parsePattern(p string) (*Pattern, error) {
	if p == "" {
		return nil, fmt.Errorf("no pattern specified")
	}
	segs := strings.Split(p, ":")
	l := len(segs)
	if l < 2 || l > 3 {
		return nil, fmt.Errorf("invalid pattern specified: %s", p)
	}
	pkg, err := compileNonEmptyRegExp(segs[0])
	if err != nil {
		return nil, fmt.Errorf("pkg pattern: %w", err)
	}
	ident, err := compileNonEmptyRegExp(segs[1])
	if err != nil {
		return nil, fmt.Errorf("ident pattern: %w", err)
	}
	pattern := Pattern{
		pkg:   pkg,
		ident: ident,
	}
	if l == 3 {
		isMethod := strings.HasSuffix(segs[2], "()")
		seg := strings.TrimSuffix(segs[2], "()")
		mof, err := compileNonEmptyRegExp(seg)
		if err != nil {
			return nil, fmt.Errorf("method or field pattern: %w", err)
		}
		if isMethod {
			pattern.method = mof
		} else {
			pattern.field = mof
		}
	}
	return &pattern, nil
}

func compileNonEmptyRegExp(p string) (*regexp.Regexp, error) {
	if p == "" {
		return nil, fmt.Errorf("empty pattern")
	}
	if !strings.HasPrefix(p, "^") {
		p = "^" + p
	}
	if !strings.HasSuffix(p, "$") {
		p = p + "$"
	}
	return regexp.Compile(p)
}

func (p Pattern) findDepInPackage(pkg *packages.Package) []token.Position {
	targetIdents := map[*ast.Ident]bool{}
	for ident, obj := range pkg.TypesInfo.Uses {
		if obj.Pkg() == nil {
			continue
		}
		if p.pkg.MatchString(obj.Pkg().Path()) && p.ident.MatchString(obj.Name()) {
			targetIdents[ident] = true
		}
	}
	if len(targetIdents) == 0 {
		return nil
	}
	var positions []token.Position
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			if p.field == nil && p.method == nil {
				ident, ok := node.(*ast.Ident)
				if !ok {
					return true
				}
				if targetIdents[ident] {
					positions = append(positions, pkg.Fset.Position(ident.Pos()))
				}
				return true
			}

			sel, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			tsel, unqualified := pkg.TypesInfo.Selections[sel]
			if !unqualified {
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if _, ok := targetIdents[ident]; !ok {
					return true
				}
				switch {
				case p.field != nil:
					field, ok := pkg.TypesInfo.Uses[sel.Sel].(*types.Var)
					if !ok {
						return true
					}
					selName := field.Name()
					if p.field.MatchString(selName) {
						positions = append(positions, pkg.Fset.Position(sel.Pos()))
					}
				case p.method != nil:
					fun, ok := pkg.TypesInfo.Uses[sel.Sel].(*types.Func)
					if !ok {
						return true
					}
					selName := fun.Name()
					if p.method.MatchString(selName) {
						positions = append(positions, pkg.Fset.Position(sel.Pos()))
					}
				}
				return true
			}

			var receiver types.Object
			switch trecv := tsel.Recv().(type) {
			case *types.Pointer:
				nt, ok := trecv.Elem().(*types.Named)
				if ok {
					receiver = nt.Obj()
				}
			case *types.Named:
				receiver = trecv.Obj()
			}
			if receiver == nil {
				return true
			}
			if receiver.Pkg() == nil {
				return true
			}
			if !p.pkg.MatchString(receiver.Pkg().Path()) {
				return true
			}
			if !p.ident.MatchString(receiver.Name()) {
				return true
			}

			switch {
			case p.field != nil:
				field, ok := tsel.Obj().(*types.Var)
				if !ok {
					return true
				}
				selName := field.Name()
				if p.field.MatchString(selName) {
					positions = append(positions, pkg.Fset.Position(sel.Pos()))
				}
			case p.method != nil:
				fun, ok := tsel.Obj().(*types.Func)
				if !ok {
					return true
				}
				selName := fun.Name()
				if p.method.MatchString(selName) {
					positions = append(positions, pkg.Fset.Position(sel.Pos()))
				}
			}
			return true
		})
	}
	return positions
}

func (p Pattern) FindDepInPackages(gopkgs []string) ([]token.Position, error) {
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	pkgs, err := packages.Load(cfg, gopkgs...)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package loading has error")
	}

	var positions []token.Position
	for _, pkg := range pkgs {
		positions = append(positions, p.findDepInPackage(pkg)...)
	}
	return positions, nil
}

func main() {
	var p string
	flag.StringVar(&p, "p", "", "<pkg path>:<ident>[:[<field>|<method>()]]")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `ged [options] [packages]

Options:`)
		flag.PrintDefaults()
	}
	flag.Parse()
	pattern, err := parsePattern(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse pattern: %v", err)
		os.Exit(1)
	}
	positions, err := pattern.FindDepInPackages(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	for _, pos := range positions {
		fmt.Println(pos.String())
	}
}
