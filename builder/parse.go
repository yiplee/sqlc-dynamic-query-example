package builder

import (
	"regexp"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func Parse(stmt string) (table string, columns []string) {
	stmt = stripComments(stmt)

	re := regexp.MustCompile(`(?is)^\s*select\s+(.*?)\s+from\s+([^\s;]+)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 3 {
		return "", nil
	}

	for col := range strings.SplitSeq(matches[1], ",") {
		col = strings.TrimSpace(col)
		col = unquote(col)
		if col == "" {
			continue
		}
		columns = append(columns, col)
	}

	table = strings.TrimSpace(matches[2])
	table = strings.TrimSuffix(table, ";")
	return table, columns
}

func SelectBuilderFromStmt(stmt string) sq.SelectBuilder {
	table, columns := Parse(stmt)
	return sq.Select(columns...).From(table)
}

// stripComments removes all SQL comments from the statement:
// - Single-line comments: `-- comment`
// - Multi-line comments: `/* comment */`
func stripComments(stmt string) string {
	// First, remove multi-line comments /* ... */
	reMultiLine := regexp.MustCompile(`/\*.*?\*/`)
	stmt = reMultiLine.ReplaceAllString(stmt, "")

	// Then, remove single-line comments -- ...
	lines := strings.Split(stmt, "\n")
	var result []string
	for _, line := range lines {
		// Find the position of -- that's not inside a string
		commentIdx := -1
		inSingleQuote := false
		inDoubleQuote := false
		inBacktick := false

		for i := 0; i < len(line)-1; i++ {
			char := line[i]
			nextChar := line[i+1]

			if char == '\'' && !inDoubleQuote && !inBacktick {
				inSingleQuote = !inSingleQuote
			} else if char == '"' && !inSingleQuote && !inBacktick {
				inDoubleQuote = !inDoubleQuote
			} else if char == '`' && !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
			} else if char == '-' && nextChar == '-' && !inSingleQuote && !inDoubleQuote && !inBacktick {
				commentIdx = i
				break
			}
		}

		if commentIdx >= 0 {
			line = line[:commentIdx]
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func unquote(col string) string {
	if len(col) >= 2 {
		if (col[0] == '`' && col[len(col)-1] == '`') ||
			(col[0] == '"' && col[len(col)-1] == '"') ||
			(col[0] == '\'' && col[len(col)-1] == '\'') {
			return col[1 : len(col)-1]
		}
	}
	return col
}
