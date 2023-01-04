import * as vscode from "vscode";
import { TextDocument, Position } from "vscode";

import { ReferenceInfo } from "@sourcegraph/cody-common";

// Get definitions of likely referenced symbols
export async function getReferences(
	document: TextDocument,
	position: Position,
	excludeRanges: vscode.Location[]
): Promise<ReferenceInfo[]> {
	const start = document.validatePosition(
		new vscode.Position(Math.max(position.line - 3, 0), 0)
	);
	const startOffset = document.offsetAt(start);
	const range = document.validateRange(
		new vscode.Range(start.line, start.character, position.line + 1, 0)
	);

	const searchText = document.getText(range);
	const matches = [...searchText.matchAll(/\w+/g)];
	const refPositions: Position[] = [];
	const ignoreWords = ["string", "Promise", "console", "log"]; // TODO(beyang): formalize
	for (const match of matches) {
		if (match.index === undefined) {
			continue;
		}
		if (ignoreWords.indexOf(match[0]) !== -1) {
			continue;
		}
		refPositions.push(document.positionAt(startOffset + match.index));
	}
	const refLocations = (
		await Promise.all(
			refPositions.map(async (pos) => {
				try {
					const res = await vscode.commands.executeCommand(
						"vscode.executeReferenceProvider",
						document.uri,
						pos
					);
					let locations = res as vscode.Location[];
					return locations.length > 3 ? locations.slice(0, 3) : locations;
				} catch (error) {
					console.error(`failed to fetch references: ${error}`);
					return [];
				}
			})
		)
	).flatMap((d) => d);
	const filteredLocations = refLocations.filter(
		(location) =>
			!excludeRanges
				.map(
					(er) => er.uri === location.uri && er.range.contains(location.range)
				)
				.some((c) => c)
	);
	return Promise.all(
		filteredLocations.map(async (l) => {
			const doc = await vscode.workspace.openTextDocument(l.uri);
			return {
				filename: doc.uri.path,
				text: doc.getText(surroundingLine(doc, l.range, 2)),
			};
		})
	);
}

function surroundingLine(
	doc: TextDocument,
	range: vscode.Range,
	numContextLines: number = 0
): vscode.Range {
	const r = new vscode.Range(
		Math.max(0, range.start.line - numContextLines),
		0,
		range.end.line + 1 + numContextLines,
		0
	);
	return doc.validateRange(r);
}
