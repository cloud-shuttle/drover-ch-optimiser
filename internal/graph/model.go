package graph

import "time"

type TableKey struct {
	Database string `json:"db"`
	Name     string `json:"name"`
}

func (k TableKey) String() string { return k.Database + "." + k.Name }

type NodeType string

const (
	NodeTable       NodeType = "table"
	NodeMV          NodeType = "mv"
	NodeView        NodeType = "view"
	NodeDictionary  NodeType = "dict"
	NodeDistributed NodeType = "distributed"
)

type Column struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DefaultKind string `json:"default_kind"`
	Comment     string `json:"comment"`
}

type Table struct {
	Key              TableKey   `json:"key"`
	Engine           string     `json:"engine"`
	EngineFull       string     `json:"engine_full"`
	CreateTableQuery string     `json:"create_table_query"`
	PartitionKey     string     `json:"partition_key"`
	SortingKey       string     `json:"sorting_key"`
	PrimaryKey       string     `json:"primary_key"`
	DependenciesDB   []string   `json:"deps_db"`
	DependenciesTbl  []string   `json:"deps_tbl"`
	Columns          []Column   `json:"columns"`
	TotalRows        uint64     `json:"total_rows"`
	TotalBytes       uint64     `json:"total_bytes"`
	ActiveParts          uint32 `json:"active_parts"`
	IncompleteMutations  uint32 `json:"incomplete_mutations"`
	HealthScore      float64    `json:"health_score"`
	HotQueries       int        `json:"hot_queries"`
	LastUsed         time.Time  `json:"last_used,omitempty"`
	NodeType         NodeType   `json:"node_type"`
	Color            string     `json:"color"`
	FilterColumns    []string   `json:"filter_columns"`
}

type EdgeType string

const (
	EdgePopulates   EdgeType = "populates"
	EdgeDependsOn   EdgeType = "depends_on"
	EdgeJoins       EdgeType = "joins"
	EdgeDistributed EdgeType = "distributed"
	EdgeQueryUses   EdgeType = "query_uses"
	EdgeDictLookup  EdgeType = "dict_lookup"
)

type Edge struct {
	From    TableKey `json:"from"`
	To      TableKey `json:"to"`
	Type    EdgeType `json:"type"`
	Weight  float64  `json:"weight"`
	Label   string   `json:"label"`
	Columns []string `json:"columns,omitempty"`
}

type Graph struct {
	Nodes    map[TableKey]*Table `json:"nodes"`
	Edges    []Edge              `json:"edges"`
	ByEngine map[string][]TableKey
	HotPaths []TableKey `json:"hot_paths"`
}

type Options struct {
	FocusTable string
	DaysBack   int
	Database   string
	Exclude    ExcludeSet
	Verbose    bool
}
