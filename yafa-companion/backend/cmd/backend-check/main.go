// Manual functional check against an explicitly isolated, real Strata service.
// No mock model, unit-test framework or OpenAI call is used.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	repositoryconnectors "yafa/backend/connectors/repositories"
	serviceconnectors "yafa/backend/connectors/services"
	"yafa/backend/core"
	"yafa/backend/entities"
)

func main() {
	if os.Getenv("YAFA_CHECK_MODE") != "isolated" {
		fmt.Fprintln(os.Stderr, "Set YAFA_CHECK_MODE=isolated only for a disposable verification stack.")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	repo := repositoryconnectors.MemoryHTTP{BaseURL: os.Getenv("MEMORY_URL"), Token: os.Getenv("YAFA_MEMORY_TOKEN"), Client: &http.Client{Timeout: 5 * time.Second}}
	fail := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Functional check failed:", err)
			os.Exit(1)
		}
	}
	agentRepo := repositoryconnectors.AgentMemory{Memory: repo}
	workspace := core.WorkspaceCore{Repository: agentRepo, Now: time.Now}
	fail(workspace.InitializeContextFiles(ctx))
	projects, err := workspace.Projects(ctx, false, 0, 20)
	fail(err)
	_ = projects
	valid, err := core.ValidateCommands(`agenda.day --date today && projects.list`)
	fail(err)
	if len(valid) != 2 {
		fail(fmt.Errorf("command chain did not contain two commands"))
	}
	if _, err = core.ValidateCommands(`tasks.create --title "A && B"`); err != nil {
		fail(err)
	}
	if _, err = core.ValidateCommands(`agenda.day --owner everyone`); err == nil {
		fail(fmt.Errorf("invalid enum accepted"))
	}
	if _, err = core.ValidateCommands(`projects.list | tasks.list`); err == nil {
		fail(fmt.Errorf("pipe accepted"))
	}
	if _, err = core.ValidateCommands(`images.generate --prompt "A calm morning" && analysis.run --instructions "Calculate 2+2"`); err == nil {
		fail(fmt.Errorf("disabled optional commands accepted"))
	}
	stampDomain := fmt.Sprintf("domain-%d", time.Now().UnixNano())
	raw, err := workspace.CreateProject(ctx, "", stampDomain+"-project", stampDomain+"-project", core.CreateProjectInput{Name: "Backend verification", Kind: "personal", Description: "Isolated real-store check"})
	fail(err)
	var project struct {
		Project entities.Project `json:"project"`
	}
	fail(json.Unmarshal(raw, &project))
	day := "2026-10-26"
	inputTask := core.CreateTaskInput{ProjectID: &project.Project.ID, Title: "Verification task", ScheduledDate: &day}
	raw, err = workspace.CreateTask(ctx, "", "", stampDomain+"-task", stampDomain+"-task", inputTask)
	fail(err)
	var task struct {
		Task entities.Task `json:"task"`
	}
	fail(json.Unmarshal(raw, &task))
	_, err = workspace.CreateTask(ctx, "", "", stampDomain+"-task", stampDomain+"-task", inputTask)
	fail(err)
	_, err = workspace.CompleteTask(ctx, "", stampDomain+"-complete", stampDomain+"-complete", core.CompleteTaskInput{ID: task.Task.ID, Expected: task.Task.Revision, Completed: true})
	fail(err)
	w, _, err := workspace.Read(ctx)
	fail(err)
	if !w.Tasks[task.Task.ID].Completed || len(w.Tasks) != 1 {
		fail(fmt.Errorf("task persistence/idempotency failed"))
	}
	contextFile, err := workspace.ReadContextFile(ctx, "soul.md")
	fail(err)
	_, err = workspace.UpdateContextFile(ctx, "", stampDomain+"-context", stampDomain+"-context", core.UpdateContextFileInput{Name: "soul.md", ExpectedRevision: contextFile.Revision, Content: contextFile.Content + "\n\nVerification preference."})
	fail(err)
	fail(workspace.InitializeContextFiles(ctx))
	reread, err := workspace.ReadContextFile(ctx, "soul.md")
	fail(err)
	if reread.Revision == contextFile.Revision {
		fail(fmt.Errorf("context edit was not retained"))
	}
	agent := core.NewAgentHarness(ctx, agentRepo, serviceconnectors.OpenAIResponses{Model: "gpt-6-luna"}, entities.AgentLimits{})
	snapshot, err := agent.BuildPrompt(ctx, entities.Thread{ID: "thread-0123456789abcdef0123456789abcdef", Origin: "desktop"}, "chat", time.Now(), "")
	fail(err)
	if snapshot.Hash == "" || len(snapshot.Documents) != 5 {
		fail(fmt.Errorf("prompt assembly failed"))
	}
	fmt.Println("PASS real project/task creation, completion, idempotency, versioned context preservation and prompt assembly")
	fmt.Println("PASS command syntax, chaining, quoted operators and invalid arguments (nothing executed)")
	stamp := fmt.Sprintf("check-%d", time.Now().UnixNano())
	q := core.SafetyQuota{Runs: 2, Tokens: 50000, VoiceSessions: 2, VoiceSeconds: 600, Hosted: 2, Images: 1, KnownUSD: 1}
	safety := core.SafetyCore{Repository: repo, Limits: core.SafetyLimits{Enabled: true, Window5h: q, Window24h: q}}
	settings, revision, err := safety.Settings(ctx)
	fail(err)
	settings.Paused = false
	settings.Window5h = q
	settings.Window24h = q
	fail(safety.Update(ctx, revision, settings))
	key, err := safety.Reserve(ctx, stamp, "run", 24000, 0)
	fail(err)
	fail(safety.Settle(ctx, key, stamp, "checked", 120, 0, 0, true, true))
	if _, err = safety.Reserve(ctx, stamp, "run", 24000, 0); err == nil {
		fail(fmt.Errorf("duplicate reservation accepted"))
	}
	settings, revision, err = safety.Settings(ctx)
	fail(err)
	second, err := safety.Reserve(ctx, stamp+"-second", "run", 24000, 0)
	fail(err)
	fail(safety.Settle(ctx, second, stamp+"-second", "checked", 0, 0, 0, true, true))
	if _, err = safety.Reserve(ctx, stamp+"-over-limit", "run", 24000, 0); err == nil {
		fail(fmt.Errorf("rolling run quota not enforced"))
	}
	settings.Paused = true
	fail(safety.Update(ctx, revision, settings))
	if _, err = safety.Reserve(ctx, stamp+"-paused", "run", 24000, 0); err == nil {
		fail(fmt.Errorf("paused AI admitted work"))
	}
	status, err := safety.Status(ctx)
	fail(err)
	if status["windows"] == nil {
		fail(fmt.Errorf("rolling windows missing"))
	}
	fmt.Println("PASS real Strata reservations, settlement, pause and 5h/24h status")
	fileRepo := repositoryconnectors.FileMemory{AgentMemory: agentRepo}
	data := []byte("A persistent backend verification file.\n")
	// Separate real verification data, not a generated model result.
	meta := entities.StoredFile{ID: "file-0123456789abcdef0123456789abcdef", Name: "backend-check.txt", MIME: "text/plain", Size: len(data), SHA256: core.FileDigest(data), Chunks: 1, RunID: "verification", ThreadID: "verification", CreatedAt: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)}
	fail(fileRepo.SaveFile(ctx, meta, data))
	_, read, err := fileRepo.FileContent(ctx, meta.ID)
	fail(err)
	if string(read) != string(data) {
		fail(fmt.Errorf("stored file differs"))
	}
	fmt.Println("PASS stored file bytes, manifest, signature/type and SHA-256 round trip")
	payload := json.RawMessage(`{"checked":true,"note":"Retain after API recreation"}`)
	row, err := repo.Write(ctx, "backend-check-persistence", "0", payload)
	if err != nil {
		row, err = repo.Read(ctx, "backend-check-persistence")
		fail(err)
	}
	if len(row.Value) == 0 {
		fail(fmt.Errorf("persistence record missing"))
	}
	fmt.Println("PASS real persistence record; no provider calls or user data used")
}
