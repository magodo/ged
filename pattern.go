package main

import (
	"fmt"
	"go/ast"
	"go/types"
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

func (p Pattern) findDepInPackage(pkg *packages.Package) Matches {
	targetIdents := map[*ast.Ident]bool{}
	for ident, obj := range pkg.TypesInfo.Uses {
		if obj.Pkg() == nil {
			continue
		}
		if p.pkg.MatchString(obj.Pkg().Path()) && p.ident.MatchString(obj.Name()) {
			targetIdents[ident] = true
		}
	}

	// len(targetIdents) == 0, doens't mean there is no match in current file, since there might be unqualified selector expressions, who .X is not necessarily an ident.
	// So we have to proceed.

	matches := Matches{}
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			if p.field == nil && p.method == nil {
				ident, ok := node.(*ast.Ident)
				if !ok {
					return true
				}
				if targetIdents[ident] {
					matches.Add(
						Match{
							pkg:   pkg.TypesInfo.Uses[ident].Pkg().Path(),
							ident: ident.Name,
						},
						pkg.Fset.Position(ident.Pos()),
					)
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
				if selObj := pkg.TypesInfo.Uses[sel.Sel]; p.matchSelObj(pkg, selObj) {
					m := Match{
						pkg:   selObj.Pkg().Path(),
						ident: ident.Name,
					}
					if p.field != nil {
						m.field = selObj.Name()
					}
					if p.method != nil {
						m.method = selObj.Name()
					}
					matches.Add(m, pkg.Fset.Position(sel.Pos()))
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

			if selObj := tsel.Obj(); p.matchSelObj(pkg, selObj) {
				m := Match{
					pkg:   selObj.Pkg().Path(),
					ident: receiver.Name(),
				}
				if p.field != nil {
					m.field = selObj.Name()
				}
				if p.method != nil {
					m.method = selObj.Name()
				}
				matches.Add(m, pkg.Fset.Position(sel.Pos()))
			}
			return true
		})
	}
	return matches
}

func (p Pattern) matchSelObj(pkg *packages.Package, selObj types.Object) bool {
	switch {
	case p.field != nil:
		field, ok := selObj.(*types.Var)
		if !ok {
			return false
		}
		selName := field.Name()
		return p.field.MatchString(selName)
	case p.method != nil:
		fun, ok := selObj.(*types.Func)
		if !ok {
			return false
		}
		selName := fun.Name()
		return p.method.MatchString(selName)
	default:
		panic("pattern has no field/method pattern")
	}
}

func (p Pattern) FindDepInPackages(gopkgs []string) (Matches, error) {
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	pkgs, err := packages.Load(cfg, gopkgs...)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package loading has error")
	}

	matches := Matches{}
	for _, pkg := range pkgs {
		matches.Merge(p.findDepInPackage(pkg))
	}
	return matches, nil
}
