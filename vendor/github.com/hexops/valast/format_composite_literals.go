package valast

func formatCompositeLiterals(input []rune) []rune {
	var (
		inStringLiteral, inRawStringLiteral bool
		depth                               int
		breakFields                         bool
		lineWidth                           int
		result                              []rune
	)
	for i, r := range input {
		switch {
		case inStringLiteral || inRawStringLiteral:
			// Reading a string literal.
			switch {
			case inStringLiteral:
				if r == '"' && (i == 0 || input[i-1] != '\\') {
					inStringLiteral = false
				}
			case inRawStringLiteral:
				if r == '`' {
					inRawStringLiteral = false
				}
			}
			if r == '\n' {
				depth = 0
				lineWidth = 0
			} else {
				lineWidth++
			}
			result = append(result, r)
		default:
			if r == '"' {
				inStringLiteral = true
				result = append(result, r)
				break
			}
			if r == '`' {
				inRawStringLiteral = true
				result = append(result, r)
				break
			}
			if r == '\n' {
				depth = 0
				lineWidth = 0
			} else {
				lineWidth++
			}
			if lineWidth >= 50 {
				breakFields = true
			}
			if r == ',' && breakFields {
				result = append(result, r)
				result = append(result, '\n')
				break
			}
			if r == '{' {
				depth++
				if depth >= 2 {
					depth = 0
					breakFields = true
					result = append(result, r)
					result = append(result, '\n')
					break
				}
			}
			if r == '}' {
				depth--
				if depth >= 2 {
					depth = 0
					breakFields = false
					result = append(result, r)
					result = append(result, ',')
					result = append(result, '\n')
					break
				}
			}
			result = append(result, r)
		}
	}
	return result
}
