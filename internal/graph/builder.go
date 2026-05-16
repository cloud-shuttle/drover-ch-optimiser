package graph

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
)

const systemDBFilter = `database NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')`

func BuildGraph(conn chclient.Querier, opts Options) (*Graph, error) {
	if opts.DaysBack <= 0 {
		opts.DaysBack = 30
	}

	g := &Graph{
		Nodes:    make(map[TableKey]*Table),
		ByEngine: make(map[string][]TableKey),
	}

	if err := fetchTables(conn, g, opts.Database, opts.Exclude); err != nil {
		return nil, err
	}
	if err := fetchColumns(conn, g); err != nil {
		return nil, err
	}
	if err := fetchParts(conn, g); err != nil {
		return nil, err
	}
	_ = fetchMutations(conn, g)
	g.buildDependencyEdges()
	_ = g.enrichWithQueryLog(conn, opts.DaysBack, opts.Exclude)
	classifyNodeTypes(g)
	g.pruneExcluded(opts.Exclude)

	if opts.FocusTable != "" {
		g = g.focusSubgraph(opts.FocusTable)
	}

	return g, nil
}

func fetchTables(conn chclient.Querier, g *Graph, database string, exclude ExcludeSet) error {
	q := fmt.Sprintf(`
SELECT
    database,
    name,
    engine,
    engine_full,
    create_table_query,
    partition_key,
    sorting_key,
    primary_key,
    dependencies_database,
    dependencies_table,
    total_rows,
    total_bytes
FROM system.tables
WHERE %s AND is_temporary = 0
  AND NOT startsWith(name, '.inner_id.')
  AND NOT startsWith(name, '.tmp.')
  AND NOT (database = 'peerdb' AND startsWith(name, '_peerdb_'))
`, systemDBFilter)
	if database != "" {
		q += fmt.Sprintf(" AND database = '%s'", escapeSQL(database))
	}
	if clause := exclude.SQLFilter(); clause != "" {
		q += " AND " + clause
	}
	q += " ORDER BY database, name"

	ctx := context.Background()
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("fetch tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			db, name, engine, engineFull, createQ string
			partKey, sortKey, primKey             string
			depsDB, depsTbl                       []string
			totalRows, totalBytes                 uint64
		)
		if err := rows.Scan(
			&db, &name, &engine, &engineFull, &createQ,
			&partKey, &sortKey, &primKey,
			&depsDB, &depsTbl,
			&totalRows, &totalBytes,
		); err != nil {
			return err
		}

		key := TableKey{Database: db, Name: name}
		t := &Table{
			Key:              key,
			Engine:           engine,
			EngineFull:       engineFull,
			CreateTableQuery: createQ,
			PartitionKey:     partKey,
			SortingKey:       sortKey,
			PrimaryKey:       primKey,
			DependenciesDB:   depsDB,
			DependenciesTbl:  depsTbl,
			TotalRows:        totalRows,
			TotalBytes:       totalBytes,
			HealthScore:      100,
			Color:            "#22c55e",
		}
		g.Nodes[key] = t
		g.ByEngine[engine] = append(g.ByEngine[engine], key)
	}
	return rows.Err()
}

func fetchColumns(conn chclient.Querier, g *Graph) error {
	q := fmt.Sprintf(`
SELECT database, table, name, type, default_kind, comment
FROM system.columns
WHERE %s
`, systemDBFilter)

	ctx := context.Background()
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("fetch columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var db, tbl, name, typ, defaultKind, comment string
		if err := rows.Scan(&db, &tbl, &name, &typ, &defaultKind, &comment); err != nil {
			return err
		}
		key := TableKey{Database: db, Name: tbl}
		node, ok := g.Nodes[key]
		if !ok {
			continue
		}
		node.Columns = append(node.Columns, Column{
			Name: name, Type: typ, DefaultKind: defaultKind, Comment: comment,
		})
	}
	return rows.Err()
}

func fetchParts(conn chclient.Querier, g *Graph) error {
	q := fmt.Sprintf(`
SELECT database, table, count() AS active_parts
FROM system.parts
WHERE active AND %s
GROUP BY database, table
`, systemDBFilter)

	ctx := context.Background()
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil // parts may be unavailable on some setups
	}
	defer rows.Close()

	for rows.Next() {
		var db, tbl string
		var parts uint64
		if err := rows.Scan(&db, &tbl, &parts); err != nil {
			return err
		}
		if node, ok := g.Nodes[TableKey{Database: db, Name: tbl}]; ok {
			node.ActiveParts = uint32(parts)
		}
	}
	return rows.Err()
}

func (g *Graph) buildDependencyEdges() {
	for key, node := range g.Nodes {
		engine := strings.ToLower(node.Engine)
		if engine == "materializedview" && len(node.DependenciesDB) > 0 {
			for i, depDB := range node.DependenciesDB {
				if i >= len(node.DependenciesTbl) {
					break
				}
				src := TableKey{Database: depDB, Name: node.DependenciesTbl[i]}
				if _, ok := g.Nodes[src]; ok {
					g.Edges = append(g.Edges, Edge{
						From: src, To: key, Type: EdgeDependsOn, Label: "MV reads",
					})
				}
				// MV target is often in engine_full / AS SELECT — edge to self population
				// Target table from TO clause is modeled via dependencies, not self-loops.
			}
		}
		if engine == "distributed" {
			// Heuristic: local table often shares name in same DB
			local := TableKey{Database: key.Database, Name: strings.TrimPrefix(key.Name, "dist_")}
			if _, ok := g.Nodes[local]; ok && local != key {
				g.Edges = append(g.Edges, Edge{
					From: key, To: local, Type: EdgeDistributed, Label: "distributed→local",
				})
			}
		}
	}
}

func fetchMutations(conn chclient.Querier, g *Graph) error {
	q := fmt.Sprintf(`
SELECT database, table, count() AS incomplete,
    countIf(positionCaseInsensitive(ifNull(command, ''), 'DELETE') > 0) AS incomplete_delete
FROM system.mutations
WHERE NOT is_done AND %s
GROUP BY database, table
`, systemDBFilter)

	ctx := context.Background()
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var db, tbl string
		var n, del uint64
		if err := rows.Scan(&db, &tbl, &n, &del); err != nil {
			return err
		}
		if node, ok := g.Nodes[TableKey{Database: db, Name: tbl}]; ok {
			node.IncompleteMutations = uint32(n)
			node.IncompleteDeleteMutations = uint32(del)
		}
	}
	return rows.Err()
}

func (g *Graph) enrichWithQueryLog(conn chclient.Querier, days int, exclude ExcludeSet) error {
	q := `
SELECT
    arrayJoin(tables) AS table_full,
    count() AS freq,
    sum(query_duration_ms) AS total_duration_ms,
    max(event_time) AS last_seen
FROM system.query_log
WHERE type = 'QueryFinish'
  AND query_kind = 'Select'
  AND event_date >= today() - ?
  AND length(tables) > 0
  AND table_full NOT LIKE 'system.%'
GROUP BY table_full
HAVING freq >= 3
ORDER BY total_duration_ms DESC
LIMIT 500
`

	ctx := context.Background()
	rows, err := conn.Query(ctx, q, days)
	if err != nil {
		return nil // query_log often disabled
	}
	defer rows.Close()

	type hot struct {
		key    TableKey
		weight float64
	}
	var hots []hot

	for rows.Next() {
		var tableFull string
		var freq uint64
		var totalMs float64
		var lastSeen time.Time
		if err := rows.Scan(&tableFull, &freq, &totalMs, &lastSeen); err != nil {
			return err
		}
		key := parseTableRef(tableFull)
		if !InScope(key, "", exclude) {
			continue
		}
		node, ok := g.Nodes[key]
		if !ok {
			continue
		}
		weight := float64(freq) + totalMs/1000
		node.HotQueries++
		if lastSeen.After(node.LastUsed) {
			node.LastUsed = lastSeen
		}
		g.Edges = append(g.Edges, Edge{
			From:   TableKey{Database: "_queries", Name: "workload"},
			To:     key,
			Type:   EdgeQueryUses,
			Weight: weight,
			Label:  "query",
		})
		hots = append(hots, hot{key: key, weight: weight})
	}

	// Top hot paths
	for i := 0; i < len(hots) && i < 10; i++ {
		g.HotPaths = append(g.HotPaths, hots[i].key)
	}

	_ = g.enrichFilterColumns(conn, days, exclude)
	_ = g.enrichQueryPatterns(conn, days, exclude)
	return rows.Err()
}

func (g *Graph) enrichFilterColumns(conn chclient.Querier, days int, exclude ExcludeSet) error {
	q := `
SELECT
    arrayJoin(tables) AS table_full,
    any(query) AS query_text,
    count() AS freq
FROM system.query_log
WHERE type = 'QueryFinish'
  AND query_kind = 'Select'
  AND event_date >= today() - ?
  AND length(tables) > 0
  AND (positionCaseInsensitive(query, 'WHERE') > 0 OR positionCaseInsensitive(query, 'PREWHERE') > 0)
GROUP BY table_full, normalized_query_hash
HAVING freq >= 2
ORDER BY freq DESC
LIMIT 2000
`
	ctx := context.Background()
	rows, err := conn.Query(ctx, q, days)
	if err != nil {
		return nil
	}
	defer rows.Close()

	perTable := make(map[TableKey]map[string]int)

	for rows.Next() {
		var tableFull, queryText string
		var freq uint64
		if err := rows.Scan(&tableFull, &queryText, &freq); err != nil {
			return err
		}
		key := parseTableRef(tableFull)
		if !InScope(key, "", exclude) {
			continue
		}
		node, ok := g.Nodes[key]
		if !ok {
			continue
		}
		names := make([]string, len(node.Columns))
		for i, c := range node.Columns {
			names[i] = c.Name
		}
		for _, col := range ExtractFilterColumns(queryText, names) {
			if perTable[key] == nil {
				perTable[key] = make(map[string]int)
			}
			perTable[key][col] += int(freq)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for key, counts := range perTable {
		node := g.Nodes[key]
		if node == nil {
			continue
		}
		type pair struct {
			name  string
			count int
		}
		var ranked []pair
		for n, c := range counts {
			ranked = append(ranked, pair{n, c})
		}
		sort.Slice(ranked, func(i, j int) bool {
			if ranked[i].count == ranked[j].count {
				return ranked[i].name < ranked[j].name
			}
			return ranked[i].count > ranked[j].count
		})
		added := make([]string, 0, len(ranked))
		for _, p := range ranked {
			added = append(added, p.name)
		}
		node.FilterColumns = mergeFilterColumns(node.FilterColumns, added, 8)
	}
	return nil
}

func (g *Graph) enrichQueryPatterns(conn chclient.Querier, days int, exclude ExcludeSet) error {
	finalQ := `
SELECT arrayJoin(tables) AS table_full, count() AS n
FROM system.query_log
WHERE type = 'QueryFinish'
  AND query_kind = 'Select'
  AND event_date >= today() - ?
  AND length(tables) > 0
  AND positionCaseInsensitive(query, ' FINAL') > 0
GROUP BY table_full
HAVING n >= 3
LIMIT 500
`
	insertQ := `
SELECT arrayJoin(tables) AS table_full, count() AS ins_n, avg(written_rows) AS avg_rows
FROM system.query_log
WHERE type = 'QueryFinish'
  AND query_kind = 'Insert'
  AND event_date >= today() - ?
  AND length(tables) > 0
  AND written_rows > 0
GROUP BY table_full
HAVING ins_n >= 12 AND avg_rows < 3000
LIMIT 500
`
	ctx := context.Background()

	if rows, err := conn.Query(ctx, finalQ, days); err == nil {
		defer rows.Close()
		for rows.Next() {
			var tableFull string
			var n uint64
			if err := rows.Scan(&tableFull, &n); err != nil {
				return err
			}
			key := parseTableRef(tableFull)
			if !InScope(key, "", exclude) {
				continue
			}
			if node, ok := g.Nodes[key]; ok {
				node.QueryLogSelectFinals = uint32(n)
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	if rows, err := conn.Query(ctx, insertQ, days); err == nil {
		defer rows.Close()
		for rows.Next() {
			var tableFull string
			var insN uint64
			var avgRows float64
			if err := rows.Scan(&tableFull, &insN, &avgRows); err != nil {
				return err
			}
			key := parseTableRef(tableFull)
			if !InScope(key, "", exclude) {
				continue
			}
			if node, ok := g.Nodes[key]; ok {
				node.QueryLogInsertCount = uint32(insN)
				node.QueryLogInsertAvgRows = avgRows
			}
		}
		return rows.Err()
	}
	return nil
}

func classifyNodeTypes(g *Graph) {
	for _, node := range g.Nodes {
		switch strings.ToLower(node.Engine) {
		case "materializedview":
			node.NodeType = NodeMV
		case "view":
			node.NodeType = NodeView
		case "dictionary":
			node.NodeType = NodeDictionary
		case "distributed":
			node.NodeType = NodeDistributed
		default:
			node.NodeType = NodeTable
		}
	}
}

func parseTableRef(s string) TableKey {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "."); idx > 0 {
		return TableKey{Database: s[:idx], Name: s[idx+1:]}
	}
	return TableKey{Database: "default", Name: s}
}

func (g *Graph) focusSubgraph(focus string) *Graph {
	key := parseTableRef(focus)
	sub := &Graph{
		Nodes: make(map[TableKey]*Table),
		Edges: nil,
	}
	if node, ok := g.Nodes[key]; ok {
		sub.Nodes[key] = node
	}
	for _, e := range g.Edges {
		if e.From == key || e.To == key {
			sub.Edges = append(sub.Edges, e)
			if n, ok := g.Nodes[e.From]; ok {
				sub.Nodes[e.From] = n
			}
			if n, ok := g.Nodes[e.To]; ok {
				sub.Nodes[e.To] = n
			}
		}
	}
	return sub
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
