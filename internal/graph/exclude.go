package graph

import (
	"fmt"
	"strings"
)

// ExcludeSet holds user-supplied --exclude patterns (databases, tables, prefixes).
type ExcludeSet struct {
	databases      map[string]struct{}
	tables         map[TableKey]struct{}
	dbPrefixes     []string
	tableWildcards []tableWildcard
}

type tableWildcard struct {
	database string
	prefix   string
}

// ParseExcludes parses repeated --exclude values:
//   - database           — entire database
//   - database.table     — one table
//   - otel*              — databases with prefix otel
//   - default.events*    — tables in default whose name starts with events
func ParseExcludes(specs []string) (ExcludeSet, error) {
	e := ExcludeSet{
		databases: make(map[string]struct{}),
		tables:    make(map[TableKey]struct{}),
	}
	for _, raw := range specs {
		spec := strings.TrimSpace(raw)
		if spec == "" {
			continue
		}
		if strings.Contains(spec, "*") {
			if !strings.Contains(spec, ".") {
				prefix := strings.TrimSuffix(spec, "*")
				if prefix == "" {
					return e, fmt.Errorf("exclude %q: empty prefix before *", raw)
				}
				e.dbPrefixes = append(e.dbPrefixes, prefix)
				continue
			}
			db, pattern, ok := strings.Cut(spec, ".")
			if !ok || db == "" {
				return e, fmt.Errorf("exclude %q: invalid wildcard table pattern", raw)
			}
			prefix := strings.TrimSuffix(pattern, "*")
			if prefix == "" {
				return e, fmt.Errorf("exclude %q: empty table prefix before *", raw)
			}
			e.tableWildcards = append(e.tableWildcards, tableWildcard{database: db, prefix: prefix})
			continue
		}
		if db, tbl, ok := strings.Cut(spec, "."); ok && tbl != "" {
			e.tables[TableKey{Database: db, Name: tbl}] = struct{}{}
			continue
		}
		e.databases[spec] = struct{}{}
	}
	return e, nil
}

// Match reports whether a table should be excluded from discovery and rules.
func (e ExcludeSet) Match(key TableKey) bool {
	if len(e.databases) == 0 && len(e.tables) == 0 && len(e.dbPrefixes) == 0 && len(e.tableWildcards) == 0 {
		return false
	}
	if _, ok := e.databases[key.Database]; ok {
		return true
	}
	for _, p := range e.dbPrefixes {
		if strings.HasPrefix(key.Database, p) {
			return true
		}
	}
	if _, ok := e.tables[key]; ok {
		return true
	}
	for _, tw := range e.tableWildcards {
		if tw.database == key.Database && strings.HasPrefix(key.Name, tw.prefix) {
			return true
		}
	}
	return false
}

// SQLFilter returns extra AND clauses for system.tables discovery (may be empty).
func (e ExcludeSet) SQLFilter() string {
	var parts []string
	if len(e.databases) > 0 {
		var quoted []string
		for db := range e.databases {
			quoted = append(quoted, "'"+escapeSQL(db)+"'")
		}
		parts = append(parts, "database NOT IN ("+strings.Join(quoted, ", ")+")")
	}
	for _, p := range e.dbPrefixes {
		parts = append(parts, "NOT startsWith(database, '"+escapeSQL(p)+"')")
	}
	for key := range e.tables {
		parts = append(parts, fmt.Sprintf(
			"NOT (database = '%s' AND name = '%s')",
			escapeSQL(key.Database), escapeSQL(key.Name),
		))
	}
	for _, tw := range e.tableWildcards {
		parts = append(parts, fmt.Sprintf(
			"NOT (database = '%s' AND startsWith(name, '%s'))",
			escapeSQL(tw.database), escapeSQL(tw.prefix),
		))
	}
	return strings.Join(parts, " AND ")
}

// InScope is true when the table passes built-in filters and is not excluded.
func InScope(key TableKey, engine string, exclude ExcludeSet) bool {
	if !ShouldAnalyze(key, engine) {
		return false
	}
	return !exclude.Match(key)
}

func (g *Graph) pruneExcluded(exclude ExcludeSet) {
	if len(exclude.databases) == 0 && len(exclude.tables) == 0 &&
		len(exclude.dbPrefixes) == 0 && len(exclude.tableWildcards) == 0 {
		return
	}
	for key := range g.Nodes {
		if exclude.Match(key) {
			delete(g.Nodes, key)
		}
	}
	var edges []Edge
	for _, e := range g.Edges {
		if _, ok := g.Nodes[e.From]; !ok {
			continue
		}
		if _, ok := g.Nodes[e.To]; !ok {
			continue
		}
		edges = append(edges, e)
	}
	g.Edges = edges
	for eng, keys := range g.ByEngine {
		var kept []TableKey
		for _, k := range keys {
			if _, ok := g.Nodes[k]; ok {
				kept = append(kept, k)
			}
		}
		if len(kept) == 0 {
			delete(g.ByEngine, eng)
		} else {
			g.ByEngine[eng] = kept
		}
	}
}
