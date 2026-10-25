// context-import is an explicit owner maintenance command, never an AI tool.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
	"yafa/backend/connectors/importbundle"
	repositoryconnectors "yafa/backend/connectors/repositories"
	"yafa/backend/core"
)

func main() {
	bundle := flag.String("bundle", "", "Private bundle directory; never embedded in the binary")
	apply := flag.Bool("apply", false, "Commit the prepared plan; otherwise read-only")
	expected := flag.String("expected-version", "", "Workspace version returned by a fresh plan; required with --apply")
	fingerprint := flag.String("expected-fingerprint", "", "Bundle fingerprint returned by the plan; required with --apply")
	verify := flag.Bool("verify-only", false, "Read and validate the bundle without any network request")
	flag.Parse()
	if *bundle == "" || (*verify && *apply) || (*apply && (*expected == "" || *fingerprint == "")) {
		fmt.Fprintln(os.Stderr, "Provide --bundle; apply also requires --expected-version and --expected-fingerprint. No default write occurs.")
		os.Exit(2)
	}
	payload, err := importbundle.Load(*bundle)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *apply && *fingerprint != payload.Fingerprint {
		fmt.Fprintln(os.Stderr, "Bundle differs from the reviewed plan; no write attempted.")
		os.Exit(1)
	}
	if *verify {
		json.NewEncoder(os.Stdout).Encode(map[string]any{"bundle_integrity_verified": true, "domain_plan_verified": false, "network_used": false, "projects": len(payload.Projects), "tasks": len(payload.Items), "documents": len(payload.Documents), "fingerprint": payload.Fingerprint})
		return
	}
	origin, err := url.Parse(os.Getenv("MEMORY_URL"))
	token := os.Getenv("YAFA_MEMORY_TOKEN")
	if err != nil || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") || (origin.Scheme != "https" && origin.Scheme != "http") || len(token) < 32 || stringsInvalidHeader(token) {
		fmt.Fprintln(os.Stderr, "Configure MEMORY_URL and the private YAFA_MEMORY_TOKEN in the execution environment. Values are never printed.")
		os.Exit(2)
	}
	client := &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	repo := repositoryconnectors.AgentMemory{Memory: repositoryconnectors.MemoryHTTP{BaseURL: origin.String(), Token: token, Client: client}}
	workspace := core.WorkspaceCore{Repository: repo, Now: time.Now}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	plan, err := workspace.ImportPersonalContext(ctx, payload, *expected, *apply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Import stopped:", err, ". No automatic retry; inspect the stored receipt after an uncertain write.")
		os.Exit(1)
	}
	if err = json.NewEncoder(os.Stdout).Encode(plan); err != nil {
		os.Exit(1)
	}
}
func stringsInvalidHeader(s string) bool {
	for _, r := range s {
		if r == '\r' || r == '\n' {
			return true
		}
	}
	return false
}
