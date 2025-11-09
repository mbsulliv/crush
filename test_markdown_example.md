# Research Document Example

## Introduction

This is a test markdown document to demonstrate Crush's enhanced markdown parsing capabilities in research mode.

## Features

### Document Structure Analysis
Crush can now extract and analyze the hierarchical structure of markdown documents, making it easier to navigate large research papers and documentation.

### Code Block Detection

Here's an example Python code block:

```python
def analyze_documents(path):
    """Analyze all documents in the given path."""
    for doc in os.listdir(path):
        if doc.endswith('.md'):
            process_markdown(doc)
```

And a JavaScript example:

```javascript
function fetchResearchData() {
    return fetch('/api/research')
        .then(response => response.json());
}
```

### Link Extraction

Crush can extract all links from markdown documents:
- [OpenAI Research](https://openai.com/research)
- [Anthropic Papers](https://anthropic.com/papers)
- [ArXiv ML Papers](https://arxiv.org/list/cs.LG/recent)

## Tables Support

| Feature | Status | Priority |
|---------|--------|----------|
| PDF Support | ✓ Completed | High |
| Word Support | ✓ Completed | High |
| Markdown Support | ✓ Enhanced | High |
| OCR Support | Planned | Medium |

## Conclusion

With enhanced markdown support, Crush can now:
1. Parse document structure
2. Extract code blocks with language identification
3. Find all links in documents
4. Preserve formatting while providing analysis

This makes it an excellent tool for research document management and analysis.