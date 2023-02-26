import * as vscode from 'vscode'

export const MARKDOWN_FORMAT_PROMPT = 'Enclose code snippets with three backticks like so: ```.'

export interface RecipeInput {
	fileName: string
	precedingText: string
	selectedText: string
	followingText: string
}

const SURROUNDING_LINES = 50

export function getActiveEditorSelection(): RecipeInput | null {
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

export function getNormalizedLanguageName(extension: string): string {
	if (!extension) {
		return ''
	}
	const language = EXTENSION_TO_LANGUAGE[extension]
	if (language) {
		return language
	}
	return extension.charAt(0).toUpperCase() + extension.slice(1)
}
