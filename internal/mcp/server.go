// Package mcp implements a minimal Model Context Protocol (JSON-RPC over stdio) server.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/config"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
)

// Run serves MCP on stdin/stdout until EOF.
func Run(ctx context.Context, in io.Reader, out io.Writer) error {
	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)
	for {
		var req rpcRequest
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		resp := handle(ctx, req)
		if req.ID != nil {
			resp.ID = req.ID
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handle(ctx context.Context, req rpcRequest) rpcResponse {
	switch req.Method {
	case "initialize":
		return rpcResponse{JSONRPC: "2.0", Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]string{"name": "drover-ch-mcp", "version": "0.1.0"},
		}}
	case "notifications/initialized", "initialized":
		return rpcResponse{JSONRPC: "2.0"}
	case "tools/list":
		return rpcResponse{JSONRPC: "2.0", Result: map[string]any{
			"tools": []map[string]any{{
				"name":        "clickhouse_audit",
				"description": "Run drover-ch warehouse audit; returns JSON report (health score, issues, graph).",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"database": map[string]string{"type": "string", "description": "Limit to one database"},
						"days":     map[string]any{"type": "integer", "description": "query_log lookback days", "default": 30},
					},
				},
			}},
		}}
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return errResp(-32602, err.Error())
		}
		if p.Name != "clickhouse_audit" {
			return errResp(-32601, "unknown tool: "+p.Name)
		}
		body, err := runAudit(ctx, p.Arguments)
		if err != nil {
			return errResp(-32000, err.Error())
		}
		return rpcResponse{JSONRPC: "2.0", Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": body}},
		}}
	default:
		if req.ID == nil {
			return rpcResponse{JSONRPC: "2.0"}
		}
		return errResp(-32601, "method not found: "+req.Method)
	}
}

func errResp(code int, msg string) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: code, Message: msg}}
}

func runAudit(ctx context.Context, args map[string]any) (string, error) {
	_ = ctx
	if err := config.Load(""); err != nil {
		return "", err
	}
	database, _ := args["database"].(string)
	days := 30
	if v, ok := args["days"].(float64); ok && v > 0 {
		days = int(v)
	}
	conn, err := chclient.NewConnection()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	g, err := graph.BuildGraph(conn, graph.Options{Database: database, DaysBack: days})
	if err != nil {
		return "", err
	}
	issues, summary := analyzer.Run(g)
	payload := map[string]any{
		"health_score": summary.HealthScore,
		"summary":      summary,
		"critical":     issues[analyzer.Critical],
		"high":         issues[analyzer.High],
		"medium":       issues[analyzer.Medium],
		"graph":        g,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Main is the entrypoint helper for cmd/drover-ch-mcp.
func Main() {
	if err := Run(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
