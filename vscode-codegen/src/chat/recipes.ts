import * as vscode from 'vscode'

import { ContextSearchOptions } from './context-search-options'

export interface RecipeInput {
	fileName: string
	precedingText: string
	selectedText: string
	followingText: string
}

const RECIPE_INPUTS: { [key: string]: () => RecipeInput | null } = {
	explainCode: getActiveEditorSelection,
	explainCodeHighLevel: getActiveEditorSelection,
	generateUnitTest: getActiveEditorSelection,
}

const RECIPE_DISPLAY_TEXTS: { [key: string]: (input: RecipeInput) => string } = {
	explainCode: (input: RecipeInput) => `Explain the following code:\n\`\`\`\n${input.selectedText}\n\`\`\``,
	explainCodeHighLevel: (input: RecipeInput) =>
		`Explain the following code at a high level:\n\`\`\`\n${input.selectedText}\n\`\`\``,
	generateUnitTest: (input: RecipeInput) =>
		`Generate a unit test for the following code:\n\`\`\`\n${input.selectedText}\n\`\`\``,
}

// TODO: Include languages in the prompt.
const RECIPE_PROMPTS: { [key: string]: (input: RecipeInput) => string } = {
	explainCode: (input: RecipeInput) =>
		`Please explain the following code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.\n\`\`\`\n${input.selectedText}\n\`\`\`\nFormat your response using Markdown. Code snippets should be enclosed with three backticks like so: \`\`\`.`,
	explainCodeHighLevel: (input: RecipeInput) =>
		`Explain the following code at a high level. Only include details that are essential to an overal understanding of what's happening in the code.\n\`\`\`\n${input.selectedText}\n\`\`\`\nFormat your response using Markdown. Code snippets should be enclosed with three backticks like so: \`\`\`.`,
	generateUnitTest: (input: RecipeInput) =>
		`Generate a unit test for the following code:\n\`\`\`\n${input.selectedText}\n\`\`\`\nFormat your response using Markdown. Code snippets should be enclosed with three backticks like so: \`\`\`.`,
}

const RECIPE_CONTEXT_OPTIONS: { [key: string]: ContextSearchOptions } = {
	explainCode: { numCodeResults: 2, numMarkdownResults: 0 },
	explainCodeHighLevel: { numCodeResults: 2, numMarkdownResults: 0 },
	// Request more results hoping that we get test files in the results.
	generateUnitTest: { numCodeResults: 5, numMarkdownResults: 0 },
}

const SURROUNDING_LINES = 50

function getActiveEditorSelection(): RecipeInput | null {
	const activeEditor = vscode.window.activeTextEditor
	if (!activeEditor) {
		vscode.window.showErrorMessage('No code selected. Please select some code and try again.')
		return null
	}
	const selection = activeEditor.selection
	if (!selection || selection?.start.isEqual(selection.end)) {
		vscode.window.showErrorMessage('No code selected. Please select some code and try again.')
		return null
	}

	const precedingText = activeEditor.document.getText(
		new vscode.Range(new vscode.Position(Math.max(0, selection.start.line - SURROUNDING_LINES), 0), selection.start)
	)
	const followingText = activeEditor.document.getText(
		new vscode.Range(selection.end, new vscode.Position(selection.end.line + SURROUNDING_LINES, 0))
	)

	return {
		fileName: activeEditor.document.fileName,
		selectedText: activeEditor.document.getText(selection),
		precedingText,
		followingText,
	}
}

export function getRecipeInput(recipe: string): RecipeInput | null {
	return RECIPE_INPUTS[recipe]()
}

export function getRecipePrompt(recipe: string, input: RecipeInput): string {
	return RECIPE_PROMPTS[recipe](input)
}

export function getRecipeDisplayText(recipe: string, input: RecipeInput): string {
	return RECIPE_DISPLAY_TEXTS[recipe](input)
}

export function getRecipeContextOptions(recipe: string): ContextSearchOptions {
	return RECIPE_CONTEXT_OPTIONS[recipe]
}
