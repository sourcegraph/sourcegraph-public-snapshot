import path from 'path'

import * as vscode from 'vscode'

import { Message } from '@sourcegraph/cody-common'

import { getActiveEditorSelection } from './helpers'
import { languageMarkdownID, languageNames } from './langs'
import { Recipe, RecipePrompt } from './recipe'

export class TranslateToLanguage implements Recipe {
	getID(): string {
		return 'translateToLanguage'
	}
	async getPrompt(maxTokens: number): Promise<RecipePrompt | null> {
		const maxInputTokens = Math.round(0.8 * maxTokens)
		const maxSurroundingTokens = Math.round(0.2 * maxTokens)

		// Inputs
		const selection = await getActiveEditorSelection()
		if (!selection) {
			return null
		}

		const qp = await vscode.window.createQuickPick()
		const origItems = languageNames.map(l => ({ label: l }))
		qp.title = 'Select a language to translate to'
		qp.items = origItems
		qp.show()
		const toLanguage = await new Promise<string | undefined>(async resolve => {
			qp.onDidChangeValue(() => {
				if (!languageNames.map(lang => lang.toLocaleLowerCase()).includes(qp.value)) {
					qp.items = [{ label: qp.value }, ...origItems]
				} else {
					qp.items = origItems
				}
			})
			qp.onDidChangeSelection(s => resolve(s[0].label))
		})
		if (!toLanguage) {
			vscode.window.showErrorMessage('Must pick a language to translate to')
			return null
		}

		// Context messages
		const contextMessages: Message[] = []

		// Get query message
		const promptMessage: Message = {
			speaker: 'you',
			text: `Translate the following code into ${toLanguage}\n\`\`\`\n${selection.selectedText}\n\`\`\``,
		}

		// Response prefix
		const markdownID = languageMarkdownID[toLanguage] || ''
		const botResponsePrefix = `Here is the code translated to ${toLanguage}:\n\`\`\`${markdownID}\n`

		return {
			displayText: promptMessage.text,
			promptMessage,
			botResponsePrefix,
			contextMessages,
		}
	}
}

export function getFileExtension(fileName: string): string {
	return path.extname(fileName).slice(1).toLowerCase()
}
