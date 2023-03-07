import * as vscode from 'vscode'
import { DocumentSymbol, SymbolKind } from 'vscode'
import { Utils } from 'vscode-uri'

import { CompletionsDocumentProvider } from '../docprovider'
import { getSymbols } from '../vscode-utils'

export async function explainCode(): Promise<string | null> {
	const activeEditor = vscode.window.activeTextEditor
	if (!activeEditor) {
		return null
	}
	const selection = activeEditor.selection
	if (!selection || selection?.start.isEqual(selection.end)) {
		vscode.window.showErrorMessage('No code selected. Please select some code and try again')
	}
	const code = activeEditor.document.getText(selection)
	return `Please explain the following code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.\n\`\`\`\n${code}\n\`\`\``
}

export async function explainCodeHighLevel(): Promise<string | null> {
	const activeEditor = vscode.window.activeTextEditor
	if (!activeEditor) {
		return null
	}
	const selection = activeEditor.selection
	if (!selection || selection?.start.isEqual(selection.end)) {
		vscode.window.showErrorMessage('No code selected. Please select some code and try again')
	}
	const code = activeEditor.document.getText(selection)
	return `Explain the following code at a high level. Only include details that are essential to an overal understanding of what's happening in the code.\n\`\`\`\n${code}\n\`\`\``
}

/**
 * Returns a prompt to use to generate unit test code.
 *
 * @param documentProvider
 * @returns
 */
export async function generateTest(documentProvider: CompletionsDocumentProvider): Promise<string | null> {
	const editor = vscode.window.activeTextEditor
	const ext = editor?.document.fileName.split('.').pop()
	if (!editor?.document.uri) {
		return null
	}
	const symbols = await getSymbols(editor.document.uri)
	symbols.sort((a, b) => {
		let diff = 0
		diff += 4 * ((isFunctionLike(a) ? 0 : 1) - (isFunctionLike(b) ? 0 : 1))
		diff += 2 * ((isProbablyTestSymbol(a) ? 0 : 1) - (isProbablyTestSymbol(b) ? 0 : 1))
		diff += 1 * ((isClose(a) ? 0 : 1) - (isClose(b) ? 0 : 1))
		return diff
	})

	const qp = await vscode.window.createQuickPick()
	qp.show()
	const selected = await new Promise(async resolve => {
		qp.title = 'select the function to test'
		qp.items = symbols.map(s => ({ label: s.name }))
		qp.onDidChangeSelection(s => resolve(s.length > 0 && s[0].label))
		qp.onDidChangeActive(() => {
			if (qp.activeItems.length === 0) {
				return
			}

			const symbolName = qp.activeItems[0].label
			const range = symbols.find(s => s.name === symbolName)?.selectionRange
			if (range) {
				editor.revealRange(range, vscode.TextEditorRevealType.AtTop)
			}
		})
	})
	if (!selected) {
		console.log('aborted (no selection made)')
		return null // aborted
	}
	const selectedSymbol = symbols.find(s => s.name === selected)
	const symbolCode = editor.document.getText(selectedSymbol?.range)

	// Get test function to mimic
	const dirname = Utils.dirname(editor.document.uri)
	const dir = await vscode.workspace.fs.readDirectory(Utils.dirname(editor.document.uri))
	const testFiles = [...dir.entries()]
		.map(([_, info]) => info[1] === vscode.FileType.File && info[0])
		.filter(e => e && isProbablyTestFile(e)) as string[]

	const testSymbols = (
		await Promise.all(
			testFiles.map(testFile => {
				const uri = Utils.joinPath(dirname, testFile)
				return getSymbols(uri).then(symbols =>
					symbols.map(symbol => ({
						uri,
						symbol,
					}))
				)
			})
		)
	).flat()
	testSymbols.sort(({ symbol: a }, { symbol: b }) => {
		let diff = 0
		diff += 4 * ((isFunctionLike(a) ? 0 : 1) - (isFunctionLike(b) ? 0 : 1))
		diff += 2 * ((isProbablyTestSymbol(b) ? 0 : 1) - (isProbablyTestSymbol(a) ? 0 : 1))
		diff += 1 * ((isClose(a) ? 0 : 1) - (isClose(b) ? 0 : 1))
		return diff
	})

	const tqp = await vscode.window.createQuickPick()
	tqp.title = "(Optional) Is there an existing test whose structure you'd like copy?"
	const noneSentinel = '[NONE]'
	tqp.items = [{ label: noneSentinel }, ...testSymbols.map(s => ({ label: s.symbol.name }))]
	tqp.onDidChangeActive(async () => {
		if (tqp.activeItems.length === 0) {
			return
		}
		const symbolName = tqp.activeItems[0].label
		const symbol = testSymbols.find(s => s.symbol.name === symbolName)
		if (symbol) {
			const doc = await vscode.workspace.openTextDocument(symbol.uri)
			const editor = await vscode.window.showTextDocument(doc, 1, true)
			editor.revealRange(symbol.symbol.range, vscode.TextEditorRevealType.AtTop)
		}
	})
	tqp.show()
	const selectedTestSymbolName = await new Promise(async resolve => {
		tqp.onDidChangeSelection(s => resolve(s.length > 0 && s[0].label))
	})
	const selectedTestSymbol = testSymbols.find(ts => ts.symbol.name === selectedTestSymbolName)

	let testCode
	if (selectedTestSymbol) {
		const testfileDoc = await vscode.workspace.openTextDocument(selectedTestSymbol.uri)
		// TODO: need to close testfileDoc?
		testCode = await testfileDoc.getText(selectedTestSymbol.symbol.range)
	}

	const testCodeString = testCode
		? `I'd like to write a unit test. Here is an example unit test that is similar to the test I would like to write:\n\`\`\`\n${testCode}\n\`\`\`\n`
		: ''
	const prompt = `${testCodeString}Here is the code I want to test:
\`\`\`
${symbolCode}
\`\`\`
Write the unit test:
\`\`\`
`

	return prompt
}

function isProbablyTestFile(filename: string): boolean {
	return filename.toLowerCase().includes('test')
}

function isProbablyTestSymbol(s: DocumentSymbol): boolean {
	if (s.name.toLowerCase().includes('test')) {
		return true
	}
	if (s.name === 'describe') {
		return true
	}
	return false
}

function isFunctionLike(s: DocumentSymbol): boolean {
	return [SymbolKind.Function, SymbolKind.Method].includes(s.kind)
}

function isClose(s: DocumentSymbol): boolean {
	const curPoint = vscode.window.activeTextEditor?.selection.end
	if (!curPoint) {
		return false
	}
	return s.range.contains(curPoint)
}
