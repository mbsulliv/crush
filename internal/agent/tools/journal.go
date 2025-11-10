package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/mbsulliv/crush/internal/filepathext"
	"github.com/mbsulliv/crush/internal/fsext"
	"github.com/mbsulliv/crush/internal/permission"
)

// JournalParams defines the parameters for the journal tool
type JournalParams struct {
	Title    string `json:"title" description:"One-line description/title of the research"`
	Content  string `json:"content" description:"The research content to add to the journal"`
	FilePath string `json:"file_path,omitempty" description:"Journal file path (default: JOURNAL.md)"`
}

// JournalPermissionsParams for permission requests
type JournalPermissionsParams struct {
	FilePath   string `json:"file_path"`
	Title      string `json:"title"`
	NewContent string `json:"new_content,omitempty"`
}

const JournalToolName = "journal"

// NewJournalTool creates a new journal tool for appending timestamped entries
func NewJournalTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JournalToolName,
		"Add timestamped research entries to JOURNAL.md in reverse chronological order",
		func(ctx context.Context, params JournalParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Default to JOURNAL.md if no path specified
			journalPath := params.FilePath
			if journalPath == "" {
				journalPath = "JOURNAL.md"
			}

			// Make absolute path
			journalPath = filepathext.SmartJoin(workingDir, journalPath)

			// Get session ID
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session_id is required")
			}

			// Read existing content if file exists
			var existingContent string
			if data, err := os.ReadFile(journalPath); err == nil {
				existingContent = string(data)
			} else if !os.IsNotExist(err) {
				return fantasy.ToolResponse{}, fmt.Errorf("error reading journal: %w", err)
			}

			// Generate timestamp in the requested format
			timestamp := time.Now().Format("2006-01-02 15:04")

			// Format the new entry with the specified heading format
			newEntry := fmt.Sprintf("# %s: %s\n\n%s\n\n---\n\n",
				timestamp,
				strings.TrimSpace(params.Title),
				strings.TrimSpace(params.Content))

			// Prepend new entry to existing content
			finalContent := newEntry + existingContent

			// Request permission to write
			action := "update journal"
			if existingContent == "" {
				action = "create journal"
			}

			p := permissions.Request(
				permission.CreatePermissionRequest{
					SessionID:   sessionID,
					Path:        fsext.PathOrPrefix(journalPath, workingDir),
					ToolCallID:  call.ID,
					ToolName:    JournalToolName,
					Action:      action,
					Description: fmt.Sprintf("Add research entry to %s", journalPath),
					Params: JournalPermissionsParams{
						FilePath:   journalPath,
						Title:      params.Title,
						NewContent: finalContent,
					},
				},
			)
			if !p {
				return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
			}

			// Write the file
			if err := os.WriteFile(journalPath, []byte(finalContent), 0o644); err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("error writing journal: %w", err)
			}

			// Record that we wrote to this file
			recordFileRead(journalPath)
			recordFileWrite(journalPath)

			// Generate response
			relPath, _ := filepath.Rel(workingDir, journalPath)
			if relPath == "" {
				relPath = journalPath
			}

			response := fmt.Sprintf("Added journal entry to %s\n\nTimestamp: %s\nTitle: %s\n\nEntry has been prepended to the journal file.",
				relPath, timestamp, params.Title)

			// If this was the first entry, mention that
			if existingContent == "" {
				response += "\n\nNote: Created new journal file."
			}

			return fantasy.NewTextResponse(response), nil
		})
}

