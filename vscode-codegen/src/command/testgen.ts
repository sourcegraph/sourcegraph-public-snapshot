// import * as vscode from "vscode";
// import { DocumentSymbol, SymbolKind } from "vscode";
// import * as openai from "openai";
// import { CompletionsDocumentProvider } from "../docprovider";
// import { Utils } from "vscode-uri";
// import { getSymbols } from "../vscode-utils";

// export async function generateTestFromSelection(
// 	documentProvider: CompletionsDocumentProvider
// ) {
// 	const editor = vscode.window.activeTextEditor;
// 	const ext = editor?.document.fileName.split(".").pop();
// 	const snippet = editor?.document.getText(editor.selection);
// 	const prompt = `Here is the code I wish to test:
//   \`\`\`
//   ${snippet}
//   \`\`\`

//   Write the code for the unit test:
//   \`\`\`
//   `;

// 	const fileUri = vscode.window.activeTextEditor?.document.uri;
// 	if (fileUri) {
// 		getSymbols(fileUri);
// 	}

// 	const completions = getCompletions(prompt);

// 	const generatedUri = vscode.Uri.parse(`codegen:unittests.${ext}`);
// 	documentProvider.setDocument(generatedUri, completions);

// 	const doc = await vscode.workspace.openTextDocument(generatedUri);
// 	await vscode.window.showTextDocument(doc, {
// 		preview: false,
// 		viewColumn: 2,
// 	});
// }

// function getCompletions(prompt: string, docstring?: string) {
// 	const config = new openai.Configuration({
// 		apiKey: vscode.workspace
// 			.getConfiguration()
// 			.get("conf.codebot.openai.apiKey"),
// 	});
// 	const oa = new openai.OpenAIApi(config);
// 	const completions = oa
// 		.createCompletion({
// 			model: "code-davinci-002",
// 			prompt: prompt,
// 			temperature: 0.2,
// 			max_tokens: 700,
// 			stop: "```",
// 			n: 1,
// 		})
// 		.then((response) =>
// 			response.data.choices.map((choice) =>
// 				docstring ? `// ${docstring}\n${choice.text}` : choice.text
// 			)
// 		)
// 		.then((choiceText) => choiceText.filter((c) => c).join("\n\n"));
// 	return completions;
// }

// export async function generateTest(
// 	documentProvider: CompletionsDocumentProvider
// ) {
// 	const editor = vscode.window.activeTextEditor;
// 	const ext = editor?.document.fileName.split(".").pop();
// 	if (!editor?.document.uri) {
// 		return;
// 	}
// 	const symbols = await getSymbols(editor.document.uri);
// 	symbols.sort((a, b) => {
// 		let diff = 0;
// 		diff += 4 * ((isFunctionLike(a) ? 0 : 1) - (isFunctionLike(b) ? 0 : 1));
// 		diff +=
// 			2 *
// 			((isProbablyTestSymbol(a) ? 0 : 1) - (isProbablyTestSymbol(b) ? 0 : 1));
// 		diff += 1 * ((isClose(a) ? 0 : 1) - (isClose(b) ? 0 : 1));
// 		return diff;
// 	});

// 	const qp = await vscode.window.createQuickPick();
// 	qp.show();
// 	const selected = await new Promise(async (resolve) => {
// 		qp.title = "select the function to test";
// 		qp.items = symbols.map((s) => ({ label: s.name }));
// 		qp.onDidChangeSelection((s) => resolve(s.length > 0 && s[0].label));
// 		qp.onDidChangeActive(() => {
// 			if (qp.activeItems.length === 0) {
// 				return;
// 			}

// 			const symbolName = qp.activeItems[0].label;
// 			const range = symbols.find((s) => s.name === symbolName)?.selectionRange;
// 			if (range) {
// 				editor.revealRange(range, vscode.TextEditorRevealType.AtTop);
// 			}
// 		});
// 	});
// 	if (!selected) {
// 		console.log("aborted (no selection made)");
// 		return; // aborted
// 	}
// 	const selectedSymbol = symbols.find((s) => s.name === selected);
// 	const symbolCode = editor.document.getText(selectedSymbol?.range);

// 	// Get test function to mimic
// 	const dirname = Utils.dirname(editor.document.uri);
// 	const dir = await vscode.workspace.fs.readDirectory(
// 		Utils.dirname(editor.document.uri)
// 	);
// 	const testFiles = [...dir.entries()]
// 		.map(([_, info]) => info[1] === vscode.FileType.File && info[0])
// 		.filter((e) => e && isProbablyTestFile(e)) as string[];

// 	const testSymbols = (
// 		await Promise.all(
// 			testFiles.map((testFile) => {
// 				const uri = Utils.joinPath(dirname, testFile);
// 				return getSymbols(uri).then((symbols) =>
// 					symbols.map((symbol) => ({
// 						uri,
// 						symbol,
// 					}))
// 				);
// 			})
// 		)
// 	).flatMap((arr) => arr);
// 	testSymbols.sort(({ symbol: a }, { symbol: b }) => {
// 		let diff = 0;
// 		diff += 4 * ((isFunctionLike(a) ? 0 : 1) - (isFunctionLike(b) ? 0 : 1));
// 		diff +=
// 			2 *
// 			((isProbablyTestSymbol(b) ? 0 : 1) - (isProbablyTestSymbol(a) ? 0 : 1));
// 		diff += 1 * ((isClose(a) ? 0 : 1) - (isClose(b) ? 0 : 1));
// 		return diff;
// 	});

// 	const tqp = await vscode.window.createQuickPick();
// 	tqp.title =
// 		"(Optional) Is there an existing test whose structure you'd like copy?";
// 	const noneSentinel = "[NONE]";
// 	tqp.items = [
// 		{ label: noneSentinel },
// 		...testSymbols.map((s) => ({ label: s.symbol.name })),
// 	];
// 	tqp.onDidChangeActive(async () => {
// 		if (tqp.activeItems.length === 0) {
// 			return;
// 		}
// 		const symbolName = tqp.activeItems[0].label;
// 		const symbol = testSymbols.find((s) => s.symbol.name === symbolName);
// 		if (symbol) {
// 			const doc = await vscode.workspace.openTextDocument(symbol.uri);
// 			const editor = await vscode.window.showTextDocument(doc, 1, true);
// 			editor.revealRange(
// 				symbol.symbol.range,
// 				vscode.TextEditorRevealType.AtTop
// 			);
// 		}
// 	});
// 	tqp.show();
// 	const selectedTestSymbolName = await new Promise(async (resolve) => {
// 		tqp.onDidChangeSelection((s) => resolve(s.length > 0 && s[0].label));
// 	});
// 	const selectedTestSymbol = testSymbols.find(
// 		(ts) => ts.symbol.name === selectedTestSymbolName
// 	);

// 	let testCode;
// 	if (selectedTestSymbol) {
// 		const testfileDoc = await vscode.workspace.openTextDocument(
// 			selectedTestSymbol.uri
// 		);
// 		// TODO: need to close testfileDoc?
// 		testCode = await testfileDoc.getText(selectedTestSymbol.symbol.range);
// 	}

// 	const testCodeString = testCode
// 		? `Here is an example unit test:\n\`\`\`\n${testCode}\n\`\`\`\n`
// 		: "";
// 	const prompt = `${testCodeString}Here is the code I want to test:
// \`\`\`
// ${symbolCode}
// \`\`\`
// Write the unit test for the function. It doesn't have to follow the example unit test exactly.
// \`\`\`
// `;

// 	console.log(prompt);
// 	const completions = getCompletions(prompt);
// 	const generatedUri = vscode.Uri.parse(`codegen:unittests.${ext}`);
// 	documentProvider.setDocument(generatedUri, completions);
// 	const doc = await vscode.workspace.openTextDocument(generatedUri);
// 	await vscode.window.showTextDocument(doc, {
// 		preview: false,
// 		viewColumn: 2,
// 	});
// }

// function isProbablyTestFile(filename: string): boolean {
// 	return filename.toLowerCase().indexOf("test") !== -1;
// }

// function isProbablyTestSymbol(s: DocumentSymbol): boolean {
// 	if (s.name.toLowerCase().indexOf("test") !== -1) {
// 		return true;
// 	}
// 	if (s.name === "describe") {
// 		return true;
// 	}
// 	return false;
// }

// function isFunctionLike(s: DocumentSymbol): boolean {
// 	return [SymbolKind.Function, SymbolKind.Method].indexOf(s.kind) !== -1;
// }

// function isClose(s: DocumentSymbol): boolean {
// 	const curPoint = vscode.window.activeTextEditor?.selection.end;
// 	if (!curPoint) return false;
// 	return s.range.contains(curPoint);
// }
