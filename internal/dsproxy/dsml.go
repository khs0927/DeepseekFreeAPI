package dsproxy

import (
	"encoding/json"
	"regexp"
	"strings"
)

// DSML (DeepSeek Markup Language) regexes.
// DeepSeek uses fullwidth vertical bars (｜, U+FF5C, UTF-8 \xef\xbd\x9c)
// or standard ASCII pipes (|).
var (
	// dsmlCallsStartRe matches <||DSML|| calls> or <｜｜DSML｜｜ calls>
	dsmlCallsStartRe = regexp.MustCompile(`(?i)<[|｜]{2}\s*DSML\s*[|｜]{2}\s*calls\s*>`)
	// dsmlCallsEndRe matches </||DSML|| calls> or </｜｜DSML｜｜ calls>
	dsmlCallsEndRe = regexp.MustCompile(`(?i)</[|｜]{2}\s*DSML\s*[|｜]{2}\s*calls\s*>`)

	// dsmlInvokeStartRe matches <||DSML|| invoke name="..."> or <｜｜DSML｜｜ invoke name="...">
	dsmlInvokeStartRe = regexp.MustCompile(`(?i)<[|｜]{2}\s*DSML\s*[|｜]{2}\s*invoke\s+name=["']?([^"'>\s]+)["']?[^>]*>`)
	// dsmlInvokeEndRe matches </||DSML|| invoke> or </｜｜DSML｜｜ invoke>
	dsmlInvokeEndRe = regexp.MustCompile(`(?i)</[|｜]{2}\s*DSML\s*[|｜]{2}\s*invoke\s*>`)

	// dsmlParamRe matches <||DSML|| parameter name="..." ...> ... </||DSML|| parameter>
	dsmlParamRe = regexp.MustCompile(`(?i)<[|｜]{2}\s*DSML\s*[|｜]{2}\s*parameter\s+name=["']?([^"'>\s]+)["']?(?:\s+string=["']?([^"'>\s]+)["']?)?[^>]*>([\s\S]*?)</[|｜]{2}\s*DSML\s*[|｜]{2}\s*parameter\s*>`)

	// dsmlInvokeFullRe matches a complete <invoke ...> ... </invoke> block
	dsmlInvokeFullRe = regexp.MustCompile(`(?i)<[|｜]{2}\s*DSML\s*[|｜]{2}\s*invoke\s+name=["']?([^"'>\s]+)["']?[^>]*>([\s\S]*?)</[|｜]{2}\s*DSML\s*[|｜]{2}\s*invoke\s*>`)
)

// dsmlParsedCall holds a tool call parsed from DSML.
type dsmlParsedCall struct {
	Name      string
	Arguments string // JSON string
}

// parseDSMLInvokes parses all <invoke>...</invoke> blocks within text.
func parseDSMLInvokes(text string) []dsmlParsedCall {
	var calls []dsmlParsedCall
	matches := dsmlInvokeFullRe.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		toolName := strings.TrimSpace(m[1])
		body := m[2]
		args := make(map[string]any)

		paramMatches := dsmlParamRe.FindAllStringSubmatch(body, -1)
		for _, pm := range paramMatches {
			paramName := strings.TrimSpace(pm[1])
			isString := strings.ToLower(strings.TrimSpace(pm[2])) == "true"
			val := pm[3]

			if isString {
				args[paramName] = val
			} else {
				t := strings.TrimSpace(val)
				var parsed any
				if err := json.Unmarshal([]byte(t), &parsed); err == nil {
					args[paramName] = parsed
				} else {
					args[paramName] = val
				}
			}
		}

		argsBytes, err := json.Marshal(args)
		var argsJSON string
		if err == nil {
			argsJSON = string(argsBytes)
		} else {
			argsJSON = "{}"
		}

		calls = append(calls, dsmlParsedCall{
			Name:      toolName,
			Arguments: argsJSON,
		})
	}
	return calls
}

// dsmlSpan represents the location of a DSML block in text.
type dsmlSpan struct {
	start, end int
	calls      []dsmlParsedCall
}

// findDSMLSpans locates all complete DSML blocks in text.
func findDSMLSpans(text string) []dsmlSpan {
	var spans []dsmlSpan

	// First look for <calls>...</calls> blocks
	callsStarts := dsmlCallsStartRe.FindAllStringIndex(text, -1)
	for _, cs := range callsStarts {
		sub := text[cs[1]:]
		ce := dsmlCallsEndRe.FindStringIndex(sub)
		if ce == nil {
			continue
		}
		blockEnd := cs[1] + ce[1]
		body := text[cs[1] : cs[1]+ce[0]]
		parsed := parseDSMLInvokes(body)
		spans = append(spans, dsmlSpan{
			start: cs[0],
			end:   blockEnd,
			calls: parsed,
		})
	}

	// If no <calls> wrapper was found, check for isolated <invoke>...</invoke> blocks
	if len(spans) == 0 {
		invokeMatches := dsmlInvokeFullRe.FindAllStringIndex(text, -1)
		for _, im := range invokeMatches {
			blockText := text[im[0]:im[1]]
			parsed := parseDSMLInvokes(blockText)
			spans = append(spans, dsmlSpan{
				start: im[0],
				end:   im[1],
				calls: parsed,
			})
		}
	}

	return spans
}

// findDSMLStart finds the first DSML opening marker in s.
// Returns start index, marker length, isCalls (true for calls, false for invoke),
// or (-1, 0, false) if not found.
func findDSMLStart(s string) (int, int, bool) {
	callLoc := dsmlCallsStartRe.FindStringIndex(s)
	invokeLoc := dsmlInvokeStartRe.FindStringIndex(s)

	if callLoc == nil && invokeLoc == nil {
		return -1, 0, false
	}
	if callLoc != nil && (invokeLoc == nil || callLoc[0] <= invokeLoc[0]) {
		return callLoc[0], callLoc[1] - callLoc[0], true
	}
	return invokeLoc[0], invokeLoc[1] - invokeLoc[0], false
}

// findDSMLEnd finds the closing marker corresponding to the start marker.
// If isCalls is true, looks for </...calls...>.
// If isCalls is false, looks for </...invoke...>.
func findDSMLEnd(s string, isCalls bool) (int, int) {
	var loc []int
	if isCalls {
		loc = dsmlCallsEndRe.FindStringIndex(s)
	} else {
		loc = dsmlInvokeEndRe.FindStringIndex(s)
	}
	if loc == nil {
		return -1, 0
	}
	return loc[0], loc[1] - loc[0]
}
