package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fumiama/go-docx"
	"github.com/ledongthuc/pdf"
)

// readPDFFile extracts text content from a PDF file
func readPDFFile(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	var content strings.Builder

	content.WriteString(fmt.Sprintf("PDF Document: %s\n", filepath.Base(filePath)))
	content.WriteString(fmt.Sprintf("Total Pages: %d\n", totalPages))
	content.WriteString("---\n\n")

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Log error but continue with other pages
			content.WriteString(fmt.Sprintf("Page %d: [Error extracting text: %v]\n\n", pageNum, err))
			continue
		}

		if strings.TrimSpace(text) != "" {
			content.WriteString(fmt.Sprintf("Page %d:\n", pageNum))
			content.WriteString(text)
			content.WriteString("\n\n")
		}
	}

	if content.Len() == 0 {
		return "", fmt.Errorf("no text content could be extracted from PDF")
	}

	return content.String(), nil
}

// readWordFile extracts text content from a Word document (DOCX format)
func readWordFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open Word file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	doc, err := docx.Parse(file, fileInfo.Size())
	if err != nil {
		return "", fmt.Errorf("failed to parse Word file: %w", err)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Word Document: %s\n", filepath.Base(filePath)))
	content.WriteString("---\n\n")

	// Simplified extraction - iterate through body items
	for _, item := range doc.Document.Body.Items {
		// For now, just extract paragraphs
		if p, ok := item.(*docx.Paragraph); ok {
			text := extractParagraphText(p)
			if strings.TrimSpace(text) != "" {
				content.WriteString(text)
				content.WriteString("\n\n")
			}
		}
		// Tables require more complex extraction, skip for now
		// Can be enhanced later when we understand the library structure better
	}

	if content.Len() == 0 {
		return "", fmt.Errorf("no text content could be extracted from Word document")
	}

	return content.String(), nil
}

// extractParagraphText extracts text from a docx paragraph
func extractParagraphText(p *docx.Paragraph) string {
	if p == nil {
		return ""
	}
	var text strings.Builder
	// Access paragraph children if they exist
	for _, child := range p.Children {
		if run, ok := child.(*docx.Run); ok {
			for _, runChild := range run.Children {
				if txt, ok := runChild.(*docx.Text); ok {
					text.WriteString(txt.Text)
				}
			}
		}
	}
	return text.String()
}

// readPlainTextFile reads a plain text file
func readPlainTextFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read text file: %w", err)
	}
	return string(content), nil
}

// readMarkdownFile reads and processes a markdown file with structure awareness
func readMarkdownFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read markdown file: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Markdown Document: %s\n", filepath.Base(filePath)))
	result.WriteString("---\n\n")

	// Extract and display document structure
	text := string(content)

	// Extract headers
	headerPattern := regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)
	headers := headerPattern.FindAllStringSubmatch(text, -1)
	if len(headers) > 0 {
		result.WriteString("## Document Structure:\n")
		for _, match := range headers {
			level := strings.Count(match[0], "#")
			indent := strings.Repeat("  ", level-1)
			result.WriteString(fmt.Sprintf("%s- %s\n", indent, strings.TrimSpace(match[1])))
		}
		result.WriteString("\n---\n\n")
	}

	// Extract code blocks for summary
	codeBlockPattern := regexp.MustCompile("(?s)```(\\w*)\\n(.*?)```")
	codeBlocks := codeBlockPattern.FindAllStringSubmatch(text, -1)
	if len(codeBlocks) > 0 {
		result.WriteString("## Code Blocks Found:\n")
		for i, block := range codeBlocks {
			lang := block[1]
			if lang == "" {
				lang = "plaintext"
			}
			lines := strings.Split(block[2], "\n")
			result.WriteString(fmt.Sprintf("- Block %d: %s (%d lines)\n", i+1, lang, len(lines)))
		}
		result.WriteString("\n---\n\n")
	}

	// Extract links
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	links := linkPattern.FindAllStringSubmatch(text, -1)
	if len(links) > 0 {
		result.WriteString("## Links Found:\n")
		for _, link := range links {
			result.WriteString(fmt.Sprintf("- [%s](%s)\n", link[1], link[2]))
		}
		result.WriteString("\n---\n\n")
	}

	// Add the full content
	result.WriteString("## Full Content:\n\n")
	result.WriteString(text)

	return result.String(), nil
}

// extractDocumentMetadata attempts to extract metadata from a document
func extractDocumentMetadata(filePath string) (map[string]string, error) {
	metadata := make(map[string]string)
	metadata["filename"] = filepath.Base(filePath)
	metadata["extension"] = strings.ToLower(filepath.Ext(filePath))

	// For PDFs, we could extract author, title, etc. if the library supports it
	// For now, just return basic metadata
	return metadata, nil
}

// detectDocumentStructure analyzes document structure (headings, sections, etc.)
func detectDocumentStructure(content string) map[string]interface{} {
	structure := make(map[string]interface{})

	lines := strings.Split(content, "\n")
	var headings []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Simple heuristic: lines in all caps or starting with numbers might be headings
		if len(trimmed) > 0 && len(trimmed) < 100 {
			if strings.ToUpper(trimmed) == trimmed && len(trimmed) > 3 {
				headings = append(headings, trimmed)
			} else if len(trimmed) > 0 && (trimmed[0] >= '1' && trimmed[0] <= '9') {
				headings = append(headings, trimmed)
			}
		}
	}

	structure["headings"] = headings
	structure["line_count"] = len(lines)
	structure["word_count"] = len(strings.Fields(content))

	return structure
}

// isDocumentFile checks if a file is a supported document format
func isDocumentFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	documentExts := []string{
		".pdf", ".docx", ".doc", ".txt", ".md", ".rtf",
		".odt", ".tex", ".log", ".csv", ".tsv",
	}

	for _, docExt := range documentExts {
		if ext == docExt {
			return true
		}
	}
	return false
}

// getDocumentReader returns the appropriate reader function for a document type
func getDocumentReader(filePath string) (func(string) (string, error), error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return readPDFFile, nil
	case ".docx", ".doc":
		return readWordFile, nil
	case ".md":
		return readMarkdownFile, nil
	case ".txt", ".log", ".csv", ".tsv":
		return readPlainTextFile, nil
	default:
		return nil, fmt.Errorf("unsupported document format: %s", ext)
	}
}