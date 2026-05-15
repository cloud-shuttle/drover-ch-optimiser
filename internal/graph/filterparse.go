package graph

import (
	"regexp"
	"sort"
	"strings"
)

var (
	whereSplitRe = regexp.MustCompile(`(?i)\b(?:PREWHERE|WHERE)\s+`)
	clauseEndRe  = regexp.MustCompile(`(?i)\b(?:GROUP\s+BY|ORDER\s+BY|LIMIT|SETTINGS|FORMAT|UNION|INTO\s+OUTFILE)\b`)
)

// ExtractFilterColumns finds column names used in WHERE/PREWHERE that exist on the table.
func ExtractFilterColumns(query string, columnNames []string) []string {
	if query == "" || len(columnNames) == 0 {
		return nil
	}
	known := make(map[string]string, len(columnNames))
	for _, c := range columnNames {
		known[strings.ToLower(c)] = c
	}
	counts := make(map[string]int)
	for _, clause := range extractWhereClauses(query) {
		lowerClause := strings.ToLower(clause)
		for lower, canon := range known {
			if columnReferenced(lowerClause, lower) {
				counts[canon]++
			}
		}
	}
	if len(counts) == 0 {
		return nil
	}
	type pair struct {
		name  string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for n, c := range counts {
		pairs = append(pairs, pair{n, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].name < pairs[j].name
		}
		return pairs[i].count > pairs[j].count
	})
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = p.name
	}
	return out
}

func extractWhereClauses(query string) []string {
	// Drop line comments and string literals crudely to reduce false positives.
	q := stripSQLComments(query)
	idxs := whereSplitRe.FindAllStringIndex(q, -1)
	if len(idxs) == 0 {
		return nil
	}
	var clauses []string
	for i, loc := range idxs {
		start := loc[1]
		end := len(q)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		chunk := q[start:end]
		if m := clauseEndRe.FindStringIndex(chunk); m != nil {
			chunk = chunk[:m[0]]
		}
		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			clauses = append(clauses, chunk)
		}
	}
	return clauses
}

func stripSQLComments(q string) string {
	var b strings.Builder
	lines := strings.Split(q, "\n")
	for _, line := range lines {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func columnReferenced(clause, column string) bool {
	// word boundary match: col, `col`, table.col, alias.col
	patterns := []string{
		"`" + column + "`",
		"." + column,
	}
	for _, p := range patterns {
		if strings.Contains(clause, p) {
			return true
		}
	}
	// standalone identifier
	for _, tok := range splitSQLTokens(clause) {
		if tok == column {
			return true
		}
	}
	return false
}

func splitSQLTokens(s string) []string {
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '.' || r == '`' {
			return r
		}
		return ' '
	}, s)
	var out []string
	for _, p := range strings.Fields(s) {
		p = strings.Trim(p, "`.")
		if p != "" {
			out = append(out, strings.ToLower(p))
		}
	}
	return out
}

// mergeFilterColumns combines ranked columns from query_log into table.FilterColumns.
func mergeFilterColumns(existing, added []string, max int) []string {
	if max <= 0 {
		max = 8
	}
	seen := make(map[string]int)
	order := make([]string, 0, len(existing)+len(added))
	for _, c := range existing {
		if _, ok := seen[c]; !ok {
			seen[c] = len(order)
			order = append(order, c)
		}
	}
	for _, c := range added {
		if _, ok := seen[c]; !ok {
			seen[c] = len(order)
			order = append(order, c)
		}
	}
	if len(order) > max {
		order = order[:max]
	}
	return order
}
