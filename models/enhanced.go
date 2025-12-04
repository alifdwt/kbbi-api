package models

import "strings"

type EnhancedDictionary struct {
	Word         string              `json:"word"`
	WordInfo     string              `json:"word_info,omitempty"`
	Definitions  []Definition        `json:"definitions"`
	WordTypes    []WordTypeEntry     `json:"word_types,omitempty"`
	Examples     []string            `json:"examples,omitempty"`
	Related      []string            `json:"related,omitempty"`
	Original     string              `json:"original,omitempty"`
}

type WordTypeEntry struct {
	Type        string       `json:"type"`
	Definitions []Definition `json:"definitions"`
}

type Definition struct {
	ID         int      `json:"id"`
	Text       string   `json:"text"`
	Examples   []string `json:"examples,omitempty"`
	CrossRef   []string `json:"cross_ref,omitempty"`
}

type ParsedResult struct {
	Word     string
	WordType string
	WordInfo string
	RestText string
}

func ParseDictionaryText(text string) EnhancedDictionary {
	result := EnhancedDictionary{Original: text}

	// Split by newlines
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return result
	}

	// Parse word and first line
	firstLine := lines[0]
	parts := strings.Fields(firstLine)

	// Extract word (first part before any formatting)
	if len(parts) > 0 {
		// Remove dots from word (e.g., "ma.in" -> "main")
		result.Word = strings.ReplaceAll(parts[0], ".", "")
	}

	// Parse word types and definitions
	result = parseWordTypes(text, result)

	return result
}

func parseWordType(text string) ParsedResult {
	result := ParsedResult{}

	// Look for common word type patterns
	patterns := []string{
		"Nomina (kata benda)",
		"Verba (kata kerja)",
		"Adjektiva (kata sifat)",
		"Adverbia (kata keterangan)",
		"Numeralia (kata bilangan)",
		"Pronomina (kata ganti)",
		"Preposisi (kata depan)",
		"Konjungsi (kata sambung)",
		"Interjeksi (kata seru)",
		"Ark",
		"Akl",
		"Bahas",
		"Bio",
		"Dok",
		"Geo",
		"Huk",
		"Isl",
		"Ling",
		"Kim",
		"Man",
		"Mat",
		"Psik",
		"Sas",
		"Sen",
		"Tern",
		"Ar",
		"Bl",
		"Jk",
		"Jw",
		"n",
		"v",
		"a",
		"adv",
	}

	// Find word type in text
	for _, pattern := range patterns {
		if idx := strings.Index(text, pattern); idx != -1 {
			result.WordType = pattern

			// Extract the rest of the text after word type
			afterType := text[idx+len(pattern):]
			result.RestText = strings.TrimSpace(afterType)
			break
		}
	}

	return result
}

func parseNumberedDefinitions(text string) []Definition {
	var definitions []Definition

	// Split by numbered patterns like (1), (2), etc.
	parts := strings.Split(text, "\n")

	var currentDef strings.Builder
	defID := 1

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if this starts a new numbered definition
		if strings.HasPrefix(part, "(") && len(part) > 3 {
			// Save previous definition if exists
			if currentDef.Len() > 0 {
				defText := strings.TrimSpace(currentDef.String())
				if defText != "" {
					def, examples := extractExamples(cleanWordTypeFromDefinition(defText))
					definitions = append(definitions, Definition{
						ID:       defID - 1,
						Text:     def,
						Examples: examples,
					})
					}
			}

			// Start new definition
			currentDef.Reset()
			defID++

			// Remove the number prefix
			if endParen := strings.Index(part, ")"); endParen != -1 {
				cleanPart := strings.TrimSpace(part[endParen+1:])
				currentDef.WriteString(cleanWordTypeFromDefinition(cleanPart))
			}
		} else {
			// Continue current definition
			if currentDef.Len() > 0 {
				currentDef.WriteString(" ")
			}
			currentDef.WriteString(part)
		}
	}

	// Add last definition
	if currentDef.Len() > 0 {
		defText := strings.TrimSpace(currentDef.String())
		if defText != "" {
			def, examples := extractExamples(cleanWordTypeFromDefinition(defText))
			definitions = append(definitions, Definition{
				ID:       defID - 1,
				Text:     def,
				Examples: examples,
			})
			}
	}

	// If no numbered definitions found but we have text, try to extract examples
	if len(definitions) == 0 && text != "" {
		def, examples := extractExamples(cleanWordTypeFromDefinition(text))
		if def != text { // Only create definition if we actually found examples
			definitions = append(definitions, Definition{
				ID:       1,
				Text:     def,
				Examples: examples,
			})
		}
	}

	return definitions
}

func cleanWordTypeFromDefinition(text string) string {
	// Remove word type patterns from definition text
	patterns := []string{
		"Verba (kata kerja) ",
		"Nomina (kata benda) ",
		"Adjektiva (kata sifat) ",
		"Adverbia (kata keterangan) ",
		"Numeralia (kata bilangan) ",
		"Pronomina (kata ganti) ",
		"Preposisi (kata depan) ",
		"Konjungsi (kata sambung) ",
		"Interjeksi (kata seru) ",
		"Ark ",
		"Akl ",
		"Bahas ",
		"Bio ",
		"Dok ",
		"Geo ",
		"Huk ",
		"Isl ",
		"Ling ",
		"Kim ",
		"Man ",
		"Mat ",
		"Psik ",
		"Sas ",
		"Sen ",
		"Tern ",
		"Ar ",
		"Bl ",
		"Jk ",
		"Jw ",
	}

	cleaned := text
	for _, pattern := range patterns {
		cleaned = strings.ReplaceAll(cleaned, pattern, "")
	}

	// Also handle patterns without trailing space
	patternsNoSpace := []string{
		"Verba (kata kerja)",
		"Nomina (kata benda)",
		"Adjektiva (kata sifat)",
		"Adverbia (kata keterangan)",
		"Numeralia (kata bilangan)",
		"Pronomina (kata ganti)",
		"Preposisi (kata depan)",
		"Konjungsi (kata sambung)",
		"Interjeksi (kata seru)",
	}

	for _, pattern := range patternsNoSpace {
		cleaned = strings.ReplaceAll(cleaned, pattern, "")
	}

	return cleaned
}

func containsWordType(text string) bool {
	patterns := []string{
		"Nomina (kata benda)",
		"Verba (kata kerja)",
		"Adjektiva (kata sifat)",
		"Adverbia (kata keterangan)",
		"Numeralia (kata bilangan)",
		"Pronomina (kata ganti)",
		"Preposisi (kata depan)",
		"Konjungsi (kata sambung)",
		"Interjeksi (kata seru)",
		"Ark",
		"Akl",
		"Bahas",
		"Bio",
		"Dok",
		"Geo",
		"Huk",
		"Isl",
		"Ling",
		"Kim",
		"Man",
		"Mat",
		"Psik",
		"Sas",
		"Sen",
		"Tern",
		"Ar",
		"Bl",
		"Jk",
		"Jw",
	}

	// Check if any word type pattern exists in text
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}

	return false
}

func extractWordTypeFromText(text string) string {
	// Look for common word type patterns
	patterns := []string{
		"Nomina (kata benda)",
		"Verba (kata kerja)",
		"Adjektiva (kata sifat)",
		"Adverbia (kata keterangan)",
		"Numeralia (kata bilangan)",
		"Pronomina (kata ganti)",
		"Preposisi (kata depan)",
		"Konjungsi (kata sambung)",
		"Interjeksi (kata seru)",
		"Ark",
		"Akl",
		"Bahas",
		"Bio",
		"Dok",
		"Geo",
		"Huk",
		"Isl",
		"Ling",
		"Kim",
		"Man",
		"Mat",
		"Psik",
		"Sas",
		"Sen",
		"Tern",
		"Ar",
		"Bl",
		"Jk",
		"Jw",
	}

	// Find word type in text
	for _, pattern := range patterns {
		if idx := strings.Index(text, pattern); idx != -1 {
			return pattern
		}
	}

	return ""
}

func parseMixedFormatDefinitions(text string) []Definition {
	var definitions []Definition

	// Split by semicolons first to separate definitions
	parts := strings.Split(text, ";")

	var currentDef strings.Builder
	defID := 1

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if this starts a numbered definition
		if strings.HasPrefix(part, "(") && len(part) > 3 {
			// Save previous definition if exists
			if currentDef.Len() > 0 {
				defText := strings.TrimSpace(currentDef.String())
				if defText != "" {
					def, examples := extractExamples(defText)
					definitions = append(definitions, Definition{
						ID:       defID - 1,
						Text:     def,
						Examples: examples,
					})
				}
			}

			// Start new definition
			currentDef.Reset()
			defID++

			// Remove the number prefix
			if endParen := strings.Index(part, ")"); endParen != -1 {
				currentDef.WriteString(strings.TrimSpace(part[endParen+1:]))
			}
		} else {
			// Check if this has examples (contains :)
			if strings.Contains(part, ":") {
				// This might be a definition with examples
				if currentDef.Len() > 0 {
					currentDef.WriteString(" ")
				}
				currentDef.WriteString(part)
			} else if !strings.Contains(part, "(") {
				// Regular definition text
				if currentDef.Len() > 0 {
					currentDef.WriteString(" ")
				}
				currentDef.WriteString(part)
			}
		}
	}

	// Add last definition
	if currentDef.Len() > 0 {
		defText := strings.TrimSpace(currentDef.String())
		if defText != "" {
			def, examples := extractExamples(defText)
			definitions = append(definitions, Definition{
				ID:       defID - 1,
				Text:     def,
				Examples: examples,
			})
		}
	}

	// If no structured definitions found, try to extract single definition with examples
	if len(definitions) == 0 {
		defText, examples := extractExamples(text)
		if defText != "" {
			definitions = append(definitions, Definition{
				ID:       1,
				Text:     defText,
				Examples: examples,
			})
		}
	}

	return definitions
}

func parseWordTypes(text string, result EnhancedDictionary) EnhancedDictionary {
	// Split by double newlines to separate word type sections
	sections := strings.Split(text, "\n\n")

	// Check if this is a single word type with reference (like "keras" with "lihat buah keras")
	// by checking if the last section doesn't start with a word type pattern or numbered definitions
	isSingleWordTypeWithReference := false
	if len(sections) >= 2 {
		lastSection := strings.TrimSpace(sections[len(sections)-1])
		// If last section doesn't contain a word type or numbered definitions, it's probably a reference
		if !containsWordType(lastSection) && !strings.Contains(lastSection, "(") && !strings.HasPrefix(lastSection, "(") {
			isSingleWordTypeWithReference = true
		}
	}

	// Handle single section case
	if len(sections) <= 2 || isSingleWordTypeWithReference {
		var remainingText string

		if len(sections) == 1 {
			// Split by single newlines and skip first line (word)
			lines := strings.Split(text, "\n")
			if len(lines) > 1 {
				remainingText = strings.Join(lines[1:], "\n")
			}
		} else if isSingleWordTypeWithReference && len(sections) > 2 {
			// Join all sections except the first and last one (exclude reference)
			remainingText = strings.Join(sections[1:len(sections)-1], "\n")
		} else {
			// For cases like "keras" where we have 2 sections, use section[0] (the one with all definitions)
			// but process it to remove the word line
			if isSingleWordTypeWithReference && len(sections) == 2 {
				lines := strings.Split(sections[0], "\n")
				if len(lines) > 1 {
					remainingText = strings.Join(lines[1:], "\n")
				}
			} else {
				// Join all sections except the first one
				remainingText = strings.Join(sections[1:], "\n")
			}
		}

		
		if remainingText != "" {
			// Try to extract word type and definition from single section
			wordType, definitions := extractWordTypeAndDefinitions(remainingText)

			if wordType != "" && len(definitions) > 0 {
				result.WordTypes = append(result.WordTypes, WordTypeEntry{
					Type:        wordType,
					Definitions: definitions,
				})
				result.Definitions = definitions
				return result
			}

			// Fallback: advanced parsing for single sections with numbered definitions
			if wordType == "" {
				wordType = extractWordTypeFromText(remainingText)
			}

			// Clean word type from the beginning of remaining text if it exists
			cleanedText := remainingText
			if wordType != "" && strings.HasPrefix(remainingText, wordType) {
				cleanedText = strings.TrimSpace(strings.TrimPrefix(remainingText, wordType))
			}

			// Try numbered definitions first
			definitions = parseNumberedDefinitions(cleanedText)

			if len(definitions) > 0 && definitions[0].Text != "" {
				result.Definitions = definitions
				if wordType != "" {
					result.WordTypes = append(result.WordTypes, WordTypeEntry{
						Type:        wordType,
						Definitions: definitions,
					})
				}
				return result
			}

			// Try mixed format as fallback
			definitions = parseMixedFormatDefinitions(cleanedText)

			if len(definitions) > 0 && definitions[0].Text != "" {
				result.Definitions = definitions
				if wordType != "" {
					result.WordTypes = append(result.WordTypes, WordTypeEntry{
						Type:        wordType,
						Definitions: definitions,
					})
				}
				return result
			}
		}
	}

	// Handle multi-section parsing (for words like 'lari')
	var allDefinitions []Definition
	var wordTypes []WordTypeEntry

	for i, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		// Skip the first section (word definition line)
		if i == 0 {
			continue
		}

		// Extract word type from this section
		wordType, definitions := extractWordTypeAndDefinitions(section)

		if wordType != "" {
			wordTypes = append(wordTypes, WordTypeEntry{
				Type:        wordType,
				Definitions: definitions,
			})
		}

		// Also add to general definitions array
		allDefinitions = append(allDefinitions, definitions...)
	}

	// If no structured word types found, fall back to simple parsing
	if len(wordTypes) == 0 {
		definitions := parseNumberedDefinitions(strings.Join(sections[1:], "\n"))
		result.Definitions = definitions
	} else {
		result.WordTypes = wordTypes
		result.Definitions = allDefinitions
	}

	return result
}

func extractWordTypeAndDefinitions(section string) (string, []Definition) {
	// Look for word type patterns at the beginning
	patterns := []string{
		"Nomina (kata benda)",
		"Verba (kata kerja)",
		"Adjektiva (kata sifat)",
		"Adverbia (kata keterangan)",
		"Numeralia (kata bilangan)",
		"Pronomina (kata ganti)",
		"Preposisi (kata depan)",
		"Konjungsi (kata sambung)",
		"Interjeksi (kata seru)",
		"Ark",
		"Akl",
		"Bahas",
		"Bio",
		"Dok",
		"Geo",
		"Huk",
		"Isl",
		"Ling",
		"Kim",
		"Man",
		"Mat",
		"Psik",
		"Sas",
		"Sen",
		"Tern",
		"Ar",
		"Bl",
		"Jk",
		"Jw",
	}

	var wordType string
	var definitionsText string

	for _, pattern := range patterns {
		if strings.HasPrefix(section, pattern) {
			wordType = pattern
			definitionsText = strings.TrimSpace(strings.TrimPrefix(section, pattern))
			break
		}
	}

	// If no pattern found, try to extract from middle of text
	if wordType == "" {
		for _, pattern := range patterns {
			if idx := strings.Index(section, pattern); idx != -1 {
				wordType = pattern
				// Get text after pattern
				afterType := section[idx+len(pattern):]
				definitionsText = strings.TrimSpace(afterType)
				break
			}
		}
	}

	// Parse definitions from the text
	var definitions []Definition
	if definitionsText != "" {
		// First try numbered definitions
		definitions = parseNumberedDefinitions(definitionsText)

		// If no numbered definitions found, try to extract single definition with examples
		if len(definitions) == 0 || (len(definitions) == 1 && definitions[0].Text == "") {
			defText, examples := extractExamples(definitionsText)
			if defText != "" {
				definitions = []Definition{
					{
						ID:       1,
						Text:     defText,
						Examples: examples,
					},
				}
			}
		}
	}

	return wordType, definitions
}

func extractExamples(text string) (string, []string) {
	var examples []string

	// Look for examples separated by ":"
	if colonIndex := strings.Index(text, ":"); colonIndex != -1 {
		defText := strings.TrimSpace(text[:colonIndex])
		examplesText := strings.TrimSpace(text[colonIndex+1:])

		// Split examples by semicolon
		exampleParts := strings.Split(examplesText, ";")
		for _, example := range exampleParts {
			example = strings.TrimSpace(example)
			if example != "" {
				// Remove trailing punctuation
				example = strings.TrimRight(example, ";,.")
				if example != "" {
					examples = append(examples, example)
				}
			}
		}

		return defText, examples
	}

	return text, examples
}