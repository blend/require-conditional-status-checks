// Copyright 2021 Blend Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitignore

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	// NOTE: Using `os.PathSeparator` on Windows can cause conflicts with the
	//       escape character for literal characters, e.g. `\?` and `\*`.
	anyPathsSuffix      = "/**"
	osPathSeparatorRune = '/'
	osPathSeparator     = "/"
	twoAsterisk         = "**"
)

// GitignoreMatch checks if a `.gitignore` pattern matches a given path.
//
// The implementation may be **incomplete**; the goal here is to do paths
// matching for the `on.<push|pull_request>.paths` fields.
//
// See: https://git-scm.com/docs/gitignore#_pattern_format
func GitignoreMatch(filename, pattern string) bool {
	// Special case negation
	if strings.HasPrefix(pattern, "!") {
		return !GitignoreMatch(filename, pattern[1:])
	}

	normalized := NormalizeGitignore(pattern)
	if normalized == "" {
		return false
	}

	// Before converting to a regular expression, remove any leading `/`
	// since our regular expression will use a `^`.
	normalized = strings.TrimPrefix(normalized, osPathSeparator)
	r := asRegexp(normalized)
	return r.MatchString(filename)
}

// NormalizeGitignore normalizes a `.gitignore` pattern by removing any
// extraneous or repeated globs and appending or prepending `**` to eliminate
// the number of rules that may apply in a given context.
//
// See: https://git-scm.com/docs/gitignore#_pattern_format
func NormalizeGitignore(pattern string) string {
	// 1. A blank line matches no files, so it can serve as a separator for
	//    readability.
	if pattern == "" {
		return ""
	}

	// 2. A line starting with # serves as a comment. Put a backslash ("\") in
	//    front of the first hash for patterns that begin with a hash.
	if strings.HasPrefix(pattern, "#") {
		return ""
	}

	// 3. Trailing spaces are ignored unless they are quoted with backslash ("\").
	pattern = strings.TrimRightFunc(pattern, unicode.IsSpace)

	// 4. An optional prefix "!" which negates the pattern; any matching file
	//    excluded by a previous pattern will become included again. It is not
	//    possible to re-include a file if a parent directory of that file is
	//    excluded. Git doesn't list excluded directories for performance
	//    reasons, so any patterns on contained files have no effect, no matter
	//    where they are defined. Put a backslash ("\") in front of the first
	//    "!" for patterns that begin with a literal "!", for example,
	//    "\!important!.txt".
	if strings.HasPrefix(pattern, "!") {
		return "!" + NormalizeGitignore(pattern[1:])
	}

	// 5. The slash / is used as the directory separator. Separators may occur
	//    at the beginning, middle or end of the .gitignore search pattern.
	// 6. If there is a separator at the beginning or middle (or both) of the
	//    pattern, then the pattern is relative to the directory level of the
	//    particular .gitignore file itself. Otherwise the pattern may also
	//    match at any level below the .gitignore level.
	pattern = earlySlash(pattern)

	// 7. If there is a separator at the end of the pattern then the pattern
	//    will only match directories, otherwise the pattern can match both
	//    files and directories.
	// 8. For example, a pattern doc/frotz/ matches doc/frotz directory, but
	//    not a/doc/frotz directory; however frotz/ matches frotz and a/frotz
	//    that is a directory (all paths are relative from the .gitignore file).
	pattern = lateSlash(pattern)

	// 9. An asterisk "*" matches anything except a slash. The character "?"
	//    matches any one character except "/". The range notation, e.g.
	//    [a-zA-Z], can be used to match one of the characters in a range.
	//    See fnmatch(3) and the FNM_PATHNAME flag for a more detailed
	//    description.
	// 10-13. See godoc for `transformTwoAsterisk()`.
	pattern = transformTwoAsterisk(pattern)

	return pattern
}

// earlySlash transforms a pattern according to rules (5) and (6). I.e. if
// a pattern has a separator at the beginning or middle (or both), it will
// match absolute paths, otherwise will match any relative path.
func earlySlash(pattern string) string {
	if strings.HasPrefix(pattern, osPathSeparator) {
		return pattern
	}

	i := strings.Index(pattern, osPathSeparator)
	if i != -1 && i < len(pattern)-1 {
		return osPathSeparator + pattern
	}

	return twoAsterisk + osPathSeparator + pattern
}

// lateSlash appends `/**` glob to the end of a pattern ending in `/`.
func lateSlash(pattern string) string {
	if strings.HasSuffix(pattern, osPathSeparator) {
		return pattern + twoAsterisk
	}

	return pattern
}

// transformTwoAsterisk modifies `pattern` by replacing all "invalid" `**`
// pairs with a single `*`. Does so according to the matching rules for
// `.gitignore`.
//
// 10. A leading "**" followed by a slash means match in all directories. For
//     example, "**/foo" matches file or directory "foo" anywhere, the same as
//     pattern "foo". "**/foo/bar" matches file or directory "bar" anywhere
//     that is directly under directory "foo".
// 11. A trailing "/**" matches everything inside. For example, "abc/**"
//     matches all files inside directory "abc", relative to the location of
//     the .gitignore file, with infinite depth.
// 12. A slash followed by two consecutive asterisks then a slash matches zero
//     or more directories. For example, "a/**/b" matches "a/b", "a/x/b",
//     "a/x/y/b" and so on.
// 13. Other consecutive asterisks are considered regular asterisks and will
//     match according to the previous rules.
func transformTwoAsterisk(pattern string) string {
	parts := strings.Split(pattern, osPathSeparator)
	for i, part := range parts {
		if part == twoAsterisk || !strings.Contains(part, twoAsterisk) {
			continue
		}
		// **Multiple** consecutive (e.g. `***`) may need more than one pass.
		// Use a "bounded while loop" to avoid an accidental infinite loop.
		for j := 0; j < 1000; j++ {
			if !strings.Contains(part, twoAsterisk) {
				break
			}
			part = strings.Replace(part, twoAsterisk, "*", -1)
		}
		parts[i] = part
	}

	return strings.Join(parts, osPathSeparator)
}

// asRegexp converts a normalized `pattern` to a regular expression. This
// assumes the caller has stripped a leading `!` and will be handling negation
// **outside of** the regular expression.
//
// It uses the following mapping for `.gitignore` matching patterns to
// regular expression matching patterns:
// - `**` --> `.*` (match anything)
// - `*` --> `[^\/]*`
// - `?` --> `[^\/]`
// - `[a-zA-Z]` and similar range notation carries over as-is
//
// Also note that a `\*`, `\?` and `\[` should be handled differently than
// a bare `*`, `?` and `[`. (Confounding this even more is the fact that the
// escape character `\` can be used as a windows path separator.)
func asRegexp(pattern string) *regexp.Regexp {
	rewritten := make([]rune, 0, len(pattern)+2)
	rewritten = append(rewritten, '^')

	// NOTE: The index `i` can **also** be incremented within the loop, e.g.
	//       if all the runes in a sequence `[a-z]` are consumed at the same
	//       time.
	p := []rune(pattern)
	for i := 0; i < len(p); i++ {
		r := p[i]
		// Always consume the "pair" when a character is escaped.
		if r == '\\' {
			rewritten = append(rewritten, []rune("\\\\\\")...)
			// Don't read beyond the end.
			if i+1 >= len(p) {
				continue
			}
			i++
			rewritten = append(rewritten, p[i])
			continue
		}

		// Always check if a single or double glob.
		if r == '*' {
			// Don't read beyond the end.
			if i+1 >= len(p) || p[i+1] != '*' {
				rewritten = append(rewritten, []rune("[^\\/]*")...)
				continue
			}

			i++
			// Also consume the next file path separator if not at the end
			// of the pattern.
			if i+1 < len(p) && p[i+1] == osPathSeparatorRune {
				i++
			}
			rewritten = append(rewritten, []rune(".*")...)
			continue
		}

		if r == '?' {
			rewritten = append(rewritten, []rune("[^\\/]")...)
			continue
		}

		if r == '[' {
			// Don't read beyond the end.
			if i+1 >= len(p) {
				rewritten = append(rewritten, '[')
				continue
			}

			extra := findClosingBracket(p[i+1:])
			rewritten = append(rewritten, p[i:i+1+extra]...)
			i += extra
			continue
		}

		// Default case, accept the rune.
		rewritten = append(rewritten, r)
	}

	rewritten = append(rewritten, '$')
	return regexp.MustCompile(string(rewritten))
}

// findClosingBracket finds the first of an (unescaped) closing bracket `]` in
// a sequence `[a-z]`. This assumes the caller has already captured a slice
// of the input starting with `a-z]...` (i.e. after the `[` was matched).
func findClosingBracket(p []rune) int {
	if len(p) == 0 {
		return 0
	}
	if p[0] == ']' {
		return 1
	}

	for i := 1; i < len(p); i++ {
		if p[i] == ']' && p[i-1] != '\\' {
			return i + 1
		}
	}

	return len(p)
}
