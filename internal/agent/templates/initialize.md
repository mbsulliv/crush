Analyze this directory and create/update the appropriate documentation file to help future agents work effectively.

**First**: Check directory contents with `ls` to determine the project type:
- If primarily code files (.py, .js, .go, etc.) → Create/update **CRUSH.md** for coding assistance
- If primarily documents (.pdf, .docx, .txt, etc.) → Create/update **RESEARCH.md** for research assistance
- If mixed content → Create both files as appropriate
- If empty or only config files → Stop and say "Directory appears empty. Add content first."

## For Code Projects (CRUSH.md)

**Goal**: Document what an agent needs to know to work in this codebase - commands, patterns, conventions, gotchas.

**Discovery process**:

1. Look for existing rule files (`.cursor/rules/*.md`, `.cursorrules`, `.github/copilot-instructions.md`, `claude.md`, `agents.md`)
2. Identify project type from config files and directory structure
3. Find build/test/lint commands from config files, scripts, Makefiles, or CI configs
4. Read representative source files to understand code patterns
5. If CRUSH.md exists, read and improve it

**Content to include**:

- Essential commands (build, test, run, deploy, etc.)
- Code organization and structure
- Naming conventions and style patterns
- Testing approach and patterns
- Important gotchas or non-obvious patterns
- Any project-specific context from existing rule files

## For Research Projects (RESEARCH.md)

**Goal**: Document the research materials, methodology, and key findings to guide future analysis.

**Discovery process**:

1. Check for existing documentation (README.md, NOTES.md, METHODOLOGY.md)
2. Identify document types and their purposes (reports, papers, datasets, notes)
3. Look for organizational patterns (chronological, thematic, by source)
4. Sample key documents to understand content and themes
5. If RESEARCH.md exists, read and improve it

**Content to include**:

- Document inventory (types, counts, key files)
- Organization structure (how files are arranged)
- Main topics and themes across documents
- Key research questions or objectives (if apparent)
- Notable findings or patterns observed
- Methodology notes (if present)
- Data files and their formats
- Suggested analysis approaches

**Format**: Clear markdown sections. Use your judgment on structure based on what you find. Aim for completeness over brevity.

**Critical**: Only document what you actually observe. Never invent content, findings, or patterns. If you can't determine something, don't include it.
