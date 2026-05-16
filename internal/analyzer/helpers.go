package analyzer

import (
	"strings"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

var highCardPatterns = []string{
	"uuid", "guid", "request_id", "trace_id", "session_id",
}

// High-cardinality id columns that are valid ORDER BY prefixes on ReplacingMergeTree dedup tables.
var replacingMergeTreeOKPrefixes = []string{
	"incident_id", "transaction_id", "event_id", "order_id",
}

var stringAbuseNamePatterns = []string{
	"timestamp", "created_at", "updated_at", "_at",
	"_ip", "ip_address",
	"amount", "price", "revenue",
}

// OTel / observability columns that are intentionally String.
var stringAllowlist = []string{
	"metricdescription", "statusmessage", "resourcelogname", "serviceinstanceid",
	"traceid", "spanid", "body", "description",
}

func detectHighCardFirst(tbl *graph.Table, sortingKey string, cols []graph.Column) string {
	if sortingKey == "" {
		return ""
	}
	first := strings.TrimSpace(strings.Split(sortingKey, ",")[0])
	first = strings.Trim(strings.Trim(first, "`"), "()")
	lower := strings.ToLower(first)

	tblName := strings.ToLower(tbl.Key.Name)
	engine := strings.ToLower(tbl.Engine)

	// OTel trace-by-id tables intentionally sort on TraceId.
	if strings.Contains(tblName, "trace_id") || strings.Contains(tblName, "traces_trace") {
		if lower == "traceid" || lower == "trace_id" {
			return ""
		}
	}
	// DynamoDB mirror tables often use id as the natural primary key.
	if strings.Contains(tblName, "dynamodb") && (lower == "id" || lower == "pk") {
		return ""
	}

	if strings.Contains(engine, "replacingmergetree") {
		for _, ok := range replacingMergeTreeOKPrefixes {
			if lower == ok {
				return ""
			}
		}
		// Entity id as first key is common for ReplacingMergeTree deduplication.
		if strings.HasSuffix(lower, "_id") || lower == "id" {
			return ""
		}
	}

	for _, p := range highCardPatterns {
		if strings.Contains(lower, p) {
			return first
		}
	}
	// *_id suffix is often high cardinality unless ReplacingMergeTree (handled above)
	if strings.HasSuffix(lower, "_id") || lower == "id" || lower == "traceid" {
		return first
	}
	for _, c := range cols {
		if strings.EqualFold(c.Name, first) && strings.Contains(strings.ToLower(c.Type), "uuid") {
			return first
		}
	}
	return ""
}

func detectStringAbuse(engine string, cols []graph.Column) []string {
	eng := strings.ToLower(engine)
	// OTel metric tables use String attributes by design
	if strings.Contains(eng, "mergetree") && (strings.Contains(strings.ToLower(engine), "otel") || len(cols) > 30) {
		// still check but with allowlist
	}

	var bad []string
	for _, c := range cols {
		typ := strings.ToLower(c.Type)
		if !strings.HasPrefix(typ, "string") && !strings.HasPrefix(typ, "nullable(string") {
			continue
		}
		name := strings.ToLower(c.Name)
		if columnAllowlisted(name) {
			continue
		}
		for _, p := range stringAbuseNamePatterns {
			if strings.Contains(name, p) {
				bad = append(bad, c.Name)
				break
			}
		}
		// status/country alone → LowCardinality rule, not critical string abuse
	}
	return bad
}

func columnAllowlisted(lowerName string) bool {
	for _, a := range stringAllowlist {
		if lowerName == a || strings.HasSuffix(lowerName, a) {
			return true
		}
	}
	return false
}

func missingLowCardinality(cols []graph.Column) string {
	for _, c := range cols {
		typ := strings.ToLower(c.Type)
		if strings.HasPrefix(typ, "string") || strings.HasPrefix(typ, "nullable(string") {
			name := strings.ToLower(c.Name)
			if strings.Contains(name, "status") || strings.Contains(name, "type") ||
				strings.Contains(name, "country") || strings.Contains(name, "region") {
				return c.Name
			}
		}
	}
	return ""
}

func countNullable(cols []graph.Column) int {
	n := 0
	for _, c := range cols {
		if strings.HasPrefix(strings.ToLower(c.Type), "nullable") {
			n++
		}
	}
	return n
}

func partitionKeyHint(partitionKey string) string {
	if partitionKey == "" {
		return ""
	}
	lower := strings.ToLower(partitionKey)
	if strings.Contains(lower, "uuid") || strings.Contains(lower, "id") && !strings.Contains(lower, "date") {
		return partitionKey
	}
	if strings.Contains(lower, "toStartOfHour") || strings.Contains(lower, "toStartOfMinute") {
		return partitionKey
	}
	return ""
}

func missingFilterInPK(tbl *graph.Table) string {
	if len(tbl.FilterColumns) == 0 || tbl.SortingKey == "" {
		return ""
	}
	sortCols := parseKeyColumns(tbl.SortingKey)
	for _, fc := range tbl.FilterColumns {
		found := false
		for i, sc := range sortCols {
			if strings.EqualFold(fc, sc) && i < 3 {
				found = true
				break
			}
		}
		if !found {
			return fc
		}
	}
	return ""
}

func missingSkippingIndex(tbl *graph.Table) string {
	ddl := strings.ToLower(tbl.CreateTableQuery)
	if len(tbl.FilterColumns) == 0 {
		return ""
	}
	sortCols := parseKeyColumns(tbl.SortingKey)
	for _, fc := range tbl.FilterColumns {
		inPK := false
		for _, sc := range sortCols {
			if strings.EqualFold(fc, sc) {
				inPK = true
				break
			}
		}
		if inPK {
			continue
		}
		if !strings.Contains(ddl, "index") || !strings.Contains(ddl, strings.ToLower(fc)) {
			return fc
		}
	}
	return ""
}

func parseKeyColumns(key string) []string {
	parts := strings.Split(key, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, "`"))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func oversizedIntegerType(cols []graph.Column) string {
	for _, c := range cols {
		lt := strings.ToLower(c.Type)
		if strings.Contains(lt, "int128") || strings.Contains(lt, "int256") ||
			strings.Contains(lt, "uint128") || strings.Contains(lt, "uint256") {
			return c.Name
		}
	}
	return ""
}

func enumStyleStringColumn(cols []graph.Column) string {
	candidates := []string{"status", "kind", "state", "tier", "environment", "env", "level", "severity"}
	for _, c := range cols {
		lt := strings.ToLower(c.Type)
		if !strings.HasPrefix(lt, "string") && !strings.HasPrefix(lt, "nullable(string") {
			continue
		}
		if strings.Contains(lt, "lowcardinality") {
			continue
		}
		name := strings.ToLower(c.Name)
		for _, want := range candidates {
			if name == want || strings.HasSuffix(name, "_"+want) {
				return c.Name
			}
		}
	}
	return ""
}

func mergeTreePartitionedWithoutTTL(tbl *graph.Table) bool {
	eng := strings.ToLower(tbl.Engine)
	if !strings.Contains(eng, "mergetree") || tbl.PartitionKey == "" {
		return false
	}
	q := strings.ToLower(tbl.CreateTableQuery)
	if strings.Contains(q, " ttl ") || strings.Contains(q, "\nttl ") {
		return false
	}
	pk := strings.ToLower(tbl.PartitionKey)
	return strings.Contains(pk, "toyyyy") || strings.Contains(pk, "todate") || strings.Contains(pk, "today()")
}

func partitionKeyUsesYYYYMMDD(pk string) bool {
	return strings.Contains(strings.ToLower(pk), "toyyyymmdd")
}

func jsonOrObjectColumn(cols []graph.Column) string {
	for _, c := range cols {
		lt := strings.ToLower(c.Type)
		if strings.HasPrefix(lt, "json") || strings.Contains(lt, "object('json')") || strings.Contains(lt, `object("json")`) {
			return c.Name
		}
	}
	return ""
}

func mvIncrementalHeuristic(tbl *graph.Table) bool {
	if !strings.Contains(strings.ToLower(tbl.Engine), "materializedview") {
		return false
	}
	q := strings.ToLower(tbl.CreateTableQuery)
	if !strings.Contains(q, "group by") {
		return false
	}
	ef := strings.ToLower(tbl.EngineFull)
	return !strings.Contains(ef, "aggregatingmergetree") && !strings.Contains(ef, "summingmergetree")
}
