/**
 * Utility functions for extracting results from LLM output
 */

export function extractFromTripleBacktickBlock(text: string): string {
	const start = text.indexOf('```')
	if (start === -1) {
		throw new Error(`Didn't get an extractable response inside triple backticks, response was:\n${text}`)
	}
	let end: number | undefined = text.indexOf('```', start + 3 + 1)
	if (end === -1) {
		end = undefined
	}
	return text.substring(start + 3, end)
}

export function extractUntilTripleBacktick(text: string): string {
	const found = text.indexOf('```')
	if (found === -1) {
		return text
	}
	return text.slice(0, Math.max(0, found))
}
