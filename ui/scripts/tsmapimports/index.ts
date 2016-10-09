import {readFileSync, writeFileSync} from "fs";
import * as ts from "typescript";

interface AddImport {
	module: string;
	fromName?: string;
	name: string;
}

type Candidates = {
	[identOrSelector: string]: AddImport;
};

interface Replacement {
	start: number;
	end: number;
	newText: string;
}

function createIdent(name: string): ts.Identifier {
	const node = ts.createNode(ts.SyntaxKind.Identifier, 0, 0) as ts.Identifier;
	node.text = name;
	return node;
}

function getDefaultOptions(): ts.FormatCodeOptions {
	return {
		IndentSize: 4,
		TabSize: 4,
		NewLineCharacter: "\n",
		ConvertTabsToSpaces: true,
		InsertSpaceAfterCommaDelimiter: true,
		InsertSpaceAfterSemicolonInForStatements: true,
		InsertSpaceBeforeAndAfterBinaryOperators: true,
		InsertSpaceAfterKeywordsInControlFlowStatements: true,
		InsertSpaceAfterFunctionKeywordForAnonymousFunctions: false,
		InsertSpaceAfterOpeningAndBeforeClosingNonemptyParenthesis: false,
		PlaceOpenBraceOnNewLineForFunctions: false,
		PlaceOpenBraceOnNewLineForControlBlocks: false,
	} as any;
}
const defaultOptions = getDefaultOptions();

function getRuleProvider(options: ts.FormatCodeOptions): any {
	// Share this between multiple formatters using the same options.
	// This represents the bulk of the space the formatter uses.
	let ruleProvider = new (ts as any).formatting.RulesProvider();
	ruleProvider.ensureUpToDate(options);
	return ruleProvider;
}
const ruleProvider = getRuleProvider(defaultOptions);

export function transform(sourceFile: ts.SourceFile, map: Candidates): string {
	// First pass to remove comments (otherwise the positions are
	// incorrect for some reason).
	(ts as any).formatting.formatDocument(sourceFile, ruleProvider, defaultOptions);
	sourceFile = ts.createSourceFile(sourceFile.fileName, sourceFile.getText(), ts.ScriptTarget.ES6, true);

	function span(node: ts.Node): ts.TextSpan {
		return ts.createTextSpanFromBounds(node.getStart(sourceFile, true), node.getEnd());
	}

	const replacements: ts.TextChange[] = [];
	const replacements2: ts.TextChange[] = [];
	const addImports: AddImport[] = [];

	transformNode(sourceFile);

	function transformNode(node: ts.Node): void {
		if (map.hasOwnProperty(node.getText())) {
			replacements.push({span: span(node), newText: map[node.getText()].name});
			addImports.push(map[node.getText()]);
			report(node, `NODE: ${node.getText()}`);
			return;
		}
		ts.forEachChild(node, transformNode);
	}

	function report(node: ts.Node, message: string): void {
		let start = sourceFile.getLineAndCharacterOfPosition(node.getStart());
		let end = sourceFile.getLineAndCharacterOfPosition(node.getEnd());
		// console.error(`${sourceFile.fileName} (${start.line + 1}:${start.character + 1}-${end.line + 1}:${end.character + 1}): ${message}`); // tslint:disable-line no-console
	}

	function applyEdits(text: string, edits: ts.TextChange[]): string {
        // Apply edits in reverse on the existing text
		let result = text;
		for (let i = edits.length - 1; i >= 0; i--) {
			let change = edits[i];
			let head = result.slice(0, change.span.start);
			let tail = result.slice(change.span.start + change.span.length);
			// console.log("DELETE", result.slice(change.span.start, change.span.start + change.span.length), "ADD", change.newText);
			result = head + change.newText + tail;
		}
		return result;
	}

	const addImportsByModule: {[mod: string]: string[]} = {};
	const modsSorted: string[] = [];
	for (let imp of addImports) {
		if (!addImportsByModule[imp.module]) {
			addImportsByModule[imp.module] = [];
		}
		const spec = imp.fromName ? `${imp.fromName} as ${imp.name}` : imp.name;
		if (addImportsByModule[imp.module].indexOf(spec) === -1) {
			addImportsByModule[imp.module].push(spec);
			addImportsByModule[imp.module].sort();
		}
		if (modsSorted.indexOf(imp.module) === -1) {
			modsSorted.push(imp.module);
			modsSorted.sort();
		}
	}

	function moduleSpecifier(n: ts.ImportDeclaration): string {
		if (n.moduleSpecifier.kind === ts.SyntaxKind.StringLiteral) {
			return (n.moduleSpecifier as ts.StringLiteral).text;
		}
		throw new Error(`unknown import module specifier: ${n.getText()}`);
	}
	function processImports(node: ts.Node): void {
		switch (node.kind) {
		case ts.SyntaxKind.ImportDeclaration:
			let imp = (node as ts.ImportDeclaration);
			const ai = addImportsByModule[moduleSpecifier(imp)];
			if (ai) {
				if (imp.importClause && imp.importClause.namedBindings && imp.importClause.namedBindings.kind === ts.SyntaxKind.NamedImports) {
					const namedImports = imp.importClause.namedBindings as ts.NamedImports;
					const names: string[] = [];
					namedImports.elements.forEach((ni) => {
						let name: string = "";
						if (ni.propertyName) {
							name = `${ni.propertyName.text} as ${ni.name.getText()}`;
						} else {
							name = ni.name.getText();
						}
						names.push(name);
					});
					ai.forEach((newName) => {
						if (names.indexOf(newName) === -1) {
							names.push(newName);
							names.sort();
						}
					});
					replacements2.push({
						span: span(node),
						newText: `import {${names.join(", ")}} from ${JSON.stringify(moduleSpecifier(imp))};`,
					});
				}
				delete addImportsByModule[moduleSpecifier(imp)];
			}
			break;
		}
		ts.forEachChild(node, processImports);
	}
	processImports(sourceFile);

	modsSorted.forEach((mod) => {
		if (addImportsByModule[mod]) {
			replacements2.push({span: {start: 0, length: 0}, newText: `import {${addImportsByModule[mod].join(", ")}} from ${JSON.stringify(mod)};\n`});
		}
	});

	const allRepl = replacements2.concat(replacements);
	const newSrc = applyEdits(sourceFile.getText(), allRepl);
	const newSrcFile = ts.createSourceFile(sourceFile.fileName, newSrc, ts.ScriptTarget.ES6, true);
	(ts as any).formatting.formatDocument(newSrcFile, ruleProvider, defaultOptions);
	return newSrcFile.getText();
}

if ((process as any).argv.length > 2) {
	const mapping = JSON.parse(readFileSync((process as any).argv[2], "utf-8"));
	const fileNames = (process as any).argv.slice(3);
	fileNames.forEach(fileName => {
		let sourceFile = ts.createSourceFile(fileName, readFileSync(fileName).toString(), ts.ScriptTarget.ES6, true);
		const newSource = transform(sourceFile, mapping);
		if (fileName === "-" || fileName === "/dev/stdin") {
			// tslint:disable: no-console
			console.log(newSource);
		} else {
			writeFileSync(fileName, newSource);
		}
	});
}

if ((process as any).env.TEST) {
	const cases = {
		"ident": {
			source: `x(1);`,
			map: {x: {module: "m", name: "x"}},
			want: `import {x} from "m";\nx(1);`,
		},
		"jsdoc": {
			source: `/* hello */\nx(1);`,
			map: {x: {module: "m", name: "x"}},
			want: `import {x} from "m";\nx(1);`,
		},
		"different length ident": {
			source: `x(1); x(2);`,
			map: {x: {module: "m", name: "xxx"}},
			want: `import {xxx} from "m";\nxxx(1); xxx(2);`,
		},
		"multiple ident": {
			source: `x(1); x(2); x(3);`,
			map: {x: {module: "m", name: "x"}},
			want: `import {x} from "m";\nx(1); x(2); x(3);`,
		},
		"import from another name": {
			source: `x(1);`,
			map: {x: {module: "m", fromName: "xx", name: "x"}},
			want: `import {xx as x} from "m";\nx(1);`,
		},
		"import aliased to another name": {
			source: `x(1);`,
			map: {x: {module: "m", fromName: "x", name: "xx"}},
			want: `import {x as xx} from "m";\nxx(1);`,
		},
		"multiple aliased imports": {
			source: `x(1); x(2);`,
			map: {x: {module: "m", fromName: "x", name: "y"}},
			want: `import {x as y} from "m";\ny(1); y(2);`,
		},
		"selector": {
			source: `x.y(1);`,
			map: {"x.y": {module: "m/x", name: "y"}},
			want: `import {y} from "m/x";\ny(1);`,
		},
		"multiple selector": {
			source: `x.y(1); x.y(2); x.y(3);`,
			map: {"x.y": {module: "m/x", name: "y"}},
			want: `import {y} from "m/x";\ny(1); y(2); y(3);`,
		},
		"multiple mappings": {
			source: `x.y(1); z(2); x.y(3);`,
			map: {"x.y": {module: "m/x", name: "y"}, "z": {module: "m/z", name: "zz"}},
			want: `import {y} from "m/x";\nimport {zz} from "m/z";\ny(1); zz(2); y(3);`,
		},
		"append import": {
			source: `import {a} from "b";\nx(1);`,
			map: {x: {module: "m", name: "x"}},
			want: `import {x} from "m";\nimport {a} from "b";\nx(1);`,
		},
		"merge imports from same module": {
			source: `import {z} from "m";\nx(1);`,
			map: {x: {module: "m", name: "x"}},
			want: `import {x, z} from "m";\nx(1);`,
		},
		"merge imports and alias imports from same module": {
			source: `import {z as zz} from "m";\nx(1);`,
			map: {x: {module: "m", name: "x"}},
			want: `import {x, z as zz} from "m";\nx(1);`,
		},
	};
	Object.keys(cases).forEach(label => {
		const c = cases[label];
		const f = ts.createSourceFile(`${label.replace(/ /g, "-")}.ts`, c.source, ts.ScriptTarget.ES6, true);
		const out = transform(f, c.map);
		// tslint:disable no-console
		if (c.want === out) {
			console.log(`# ${label}: OK`);
			console.log();
		} else {
			console.log(`# ${label}: FAIL`);
			console.log(`  want: ${JSON.stringify(c.want)}`);
			console.log(`  got:  ${JSON.stringify(out)}`);
			console.log();
		}
	});
}
