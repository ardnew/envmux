package parse

// import (
// 	"fmt"
// 	"strings"
// )

// func (v Value) String() string {
// 	return fmt.Sprintf(`{ "Literal": %q }`, v.Literal.String())
// }

// func (c Capture) String() string {
// 	return fmt.Sprintf(`%q`, c.Name)
// }

// func (c Captures) String() string {
// 	s := make([]string, len(c.List))
// 	for i, cap := range c.List {
// 		s[i] = cap.String()
// 	}
// 	return fmt.Sprintf(`"Capt": [ %v ]`, strings.Join(s, ", "))
// }

// func (m Mapping) String() string {
// 	return fmt.Sprintf(`{ %q: %v }`, m.Key, m.Value)
// }

// func (d Dictionary) String() string {
// 	s := make([]string, len(d.Map))
// 	for i, m := range d.Map {
// 		s[i] = m.String()
// 	}
// 	return fmt.Sprintf(`"Dict": [ %v ]`, strings.Join(s, ", "))
// }

// func (p Package) String() string {
// 	return fmt.Sprintf(`%q`, p.Path.String())
// }

// func (p Packages) String() string {
// 	s := make([]string, len(p.List))
// 	for i, pkg := range p.List {
// 		s[i] = pkg.String()
// 	}
// 	return fmt.Sprintf(`"Pack": [ %v ]`, strings.Join(s, ", "))
// }

// func (d Definition) String() string {
// 	s := make([]string, 0, 3)
// 	if d.Capt != nil {
// 		s = append(s, d.Capt.String())
// 	}
// 	if d.Dict != nil {
// 		s = append(s, d.Dict.String())
// 	}
// 	if d.Pack != nil {
// 		s = append(s, d.Pack.String())
// 	}
// 	if len(s) == 0 {
// 		return "{}"
// 	}
// 	return fmt.Sprintf(`{ %v }`, strings.Join(s, ", "))
// }

// func (n Namespace) String() string {
// 	name := n.Name
// 	if name == "" {
// 		name = "DEFAULT"
// 	}
// 	return fmt.Sprintf(`%q: %v`, name, n.Definition)
// }

// func (n Namespaces) String() string {
// 	s := make([]string, len(n.List))
// 	for i, ns := range n.List {
// 		s[i] = ns.String()
// 	}
// 	return fmt.Sprintf(`{ %v }`, strings.Join(s, ", "))
// }
