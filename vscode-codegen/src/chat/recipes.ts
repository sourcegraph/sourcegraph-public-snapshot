import path from 'path'
import * as vscode from 'vscode'

import { ContextSearchOptions } from './context-search-options'

export interface RecipeInput {
	fileName: string
	precedingText: string
	selectedText: string
	followingText: string
}

function getFileExtension(fileName: string): string {
	return path.extname(fileName).slice(1).toLowerCase()
}

const RECIPE_INPUTS: { [key: string]: () => RecipeInput | null } = {
	explainCode: getActiveEditorSelection,
	explainCodeHighLevel: getActiveEditorSelection,
	generateUnitTest: getActiveEditorSelection,
	generateDocstring: getActiveEditorSelection,
}

const RECIPE_RESPONSE_PREFIXES: { [key: string]: (input: RecipeInput) => string } = {
	explainCode: () => '',
	explainCodeHighLevel: () => '',
	generateUnitTest: (input: RecipeInput) => {
		const extension = getFileExtension(input.fileName)
		return `Here is the generated unit test:\n\`\`\`${extension}\n`
	},
	generateDocstring: (input: RecipeInput) => {
		const extension = getFileExtension(input.fileName)

		let docStart = ''
		if (extension === 'java' || extension.startsWith('js') || extension.startsWith('ts')) {
			docStart = '/*'
		} else if (extension === 'py') {
			docStart = '"""\n'
		} else if (extension === 'go') {
			docStart = '// '
		}

		return `Here is the generated documentation:\n\`\`\`${extension}\n${docStart}`
	},
}

const RECIPE_DISPLAY_TEXTS: { [key: string]: (input: RecipeInput) => string } = {
	explainCode: (input: RecipeInput) => `Explain the following code:\n\`\`\`\n${input.selectedText}\n\`\`\``,
	explainCodeHighLevel: (input: RecipeInput) =>
		`Explain the following code at a high level:\n\`\`\`\n${input.selectedText}\n\`\`\``,
	generateUnitTest: (input: RecipeInput) =>
		`Generate a unit test for the following code:\n\`\`\`\n${input.selectedText}\n\`\`\``,
	generateDocstring: (input: RecipeInput) =>
		`Generate documentation for the following code:\n\`\`\`\n${input.selectedText}\n\`\`\``,
}

const MARKDOWN_FORMAT_PROMPT = `Enclose code snippets with three backticks like so: \`\`\`.`

const EXTENSION_TO_LANGUAGE: { [key: string]: string } = {
	py: 'Python',
	rb: 'Ruby',
	md: 'Markdown',
	php: 'PHP',
	js: 'Javascript',
	ts: 'Typescript',
	jsx: 'JSX',
	tsx: 'TSX',
}

function getNormalizedLanguageName(extension: string): string {
	if (!extension) {
		return ''
	}
	const language = EXTENSION_TO_LANGUAGE[extension]
	if (language) {
		return language
	}
	return extension.charAt(0).toUpperCase() + extension.slice(1)
}

const RECIPE_PROMPTS: { [key: string]: (input: RecipeInput) => string } = {
	explainCode: (input: RecipeInput): string => {
		const languageName = getNormalizedLanguageName(input.fileName)
		return `Please explain the following ${languageName} code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.\n\`\`\`\n${input.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
	},
	explainCodeHighLevel: (input: RecipeInput): string => {
		const languageName = getNormalizedLanguageName(input.fileName)
		return `Explain the following ${languageName} code at a high level. Only include details that are essential to an overal understanding of what's happening in the code.\n\`\`\`\n${input.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
	},
	generateUnitTest: (input: RecipeInput): string => {
		const languageName = getNormalizedLanguageName(input.fileName)
		return `Generate a unit test in ${languageName} for the following code:\n\`\`\`\n${input.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`
	},
	generateDocstring: (input: RecipeInput): string => {
		const extension = getFileExtension(input.fileName)
		const languageName = getNormalizedLanguageName(input.fileName)
		const promptPrefix = `Generate a comment documenting the parameters and functionality of the following ${languageName} code:`
		let additionalInstructions = `Use the ${languageName} documentation style to generate a ${languageName} comment.`
		if (extension === 'java') {
			additionalInstructions = `Use the JavaDoc documentation style to generate a Java comment.`
		} else if (extension === 'py') {
			additionalInstructions = `Use a Python docstring to generate a Python multi-line string.`
		}
		return `${promptPrefix}\n\`\`\`\n${input.selectedText}\n\`\`\`\n Only generate the documentation, do not generate the code. ${additionalInstructions} ${MARKDOWN_FORMAT_PROMPT}`
	},
}

const RECIPE_CONTEXT_OPTIONS: { [key: string]: ContextSearchOptions } = {
	explainCode: { numCodeResults: 2, numMarkdownResults: 0 },
	explainCodeHighLevel: { numCodeResults: 2, numMarkdownResults: 0 },
	// Request more results hoping that we get test files in the results.
	generateUnitTest: { numCodeResults: 5, numMarkdownResults: 0 },
	generateDocstring: { numCodeResults: 2, numMarkdownResults: 0 },
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

export function getRecipeResponsePrefix(recipe: string, input: RecipeInput): string {
	return RECIPE_RESPONSE_PREFIXES[recipe](input)
}
