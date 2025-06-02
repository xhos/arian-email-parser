package parser

var registry []Parser

// Register adds a new parser to the registry
func Register(p Parser) { registry = append(registry, p) }

// Find returns the first parser that matches the given EmailMeta
func Find(meta EmailMeta) Parser {
	for _, p := range registry {
		if p.Match(meta) {
			return p
		}
	}
	return nil
}
