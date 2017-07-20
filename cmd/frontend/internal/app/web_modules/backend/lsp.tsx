import { doFetch as fetch } from "app/backend/xhr";
import { getModeFromExtension, getPathExtension, supportedExtensions } from "app/util";
import { RepoRevSpec } from "app/util/types";
import { Reference, TooltipData } from "app/util/types";

const tooltipCache: { [key: string]: TooltipData } = {};
const j2dCache = {};
const referencesCache = {};

const sourcegraphUrl = "https://sourcegraph.com";

interface LSPRequest {
	method: string;
	params: any;
}

function wrapLSP(req: LSPRequest, repoRevSpec: RepoRevSpec, path: string): any[] {
	return [
		{
			id: 0,
			method: "initialize",
			params: {
				rootPath: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}`,
				mode: `${getModeFromExtension(getPathExtension(path))}`,
			},
		},
		{
			id: 1,
			...req,
		},
		{
			id: 2,
			method: "shutdown",
		},
		{
			// id not included on 'exit' requests
			method: "exit",
		},
	];
}

export function getTooltip(path: string, line: number, char: number, repoRevSpec: RepoRevSpec): Promise<TooltipData> {
	const ext = getPathExtension(path);
	if (!supportedExtensions.has(ext)) {
		return Promise.resolve({});
	}

	const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${char}`;

	if (tooltipCache[cacheKey]) {
		return Promise.resolve(tooltipCache[cacheKey]!);
	}

	const body = wrapLSP({
		method: "textDocument/hover",
		params: {
			textDocument: {
				uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
			},
			position: {
				character: char - 1,
				line: line - 1,
			},
		},
	}, repoRevSpec, path);

	return fetch(`${sourcegraphUrl}/.api/xlang/textDocument/hover`, { method: "POST", body: JSON.stringify(body) })
		.then((resp) => resp.json()).then((json) => {
			if (json[1].result && json[1].result.contents && json[1].result.contents.length > 0) {
				const title = json[1].result.contents[0].value;
				let doc;
				json[1].result.contents.forEach(markedString => {
					if (typeof markedString === "string") {
						doc = markedString;
					} else if (markedString.language === "markdown") {
						doc = markedString.value;
					}
				});
				tooltipCache[cacheKey] = { title, doc };
			} else {
				tooltipCache[cacheKey] = {};
			}
			return tooltipCache[cacheKey]!;
		});
}

export function fetchJumpURL(col: number, path: string, line: number, repoRevSpec: RepoRevSpec): Promise<string | null> {
	const ext = getPathExtension(path);
	if (!supportedExtensions.has(ext)) {
		return Promise.resolve(null);
	}
	const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${col}`;
	if (j2dCache[cacheKey]) {
		return Promise.resolve(j2dCache[cacheKey]);
	}

	const body = wrapLSP({
		method: "textDocument/definition",
		params: {
			textDocument: {
				uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
			},
			position: {
				character: col - 1,
				line: line - 1,
			},
		},
	}, repoRevSpec, path);

	return fetch(`${sourcegraphUrl}/.api/xlang/textDocument/definition`, { method: "POST", body: JSON.stringify(body) })
		.then((resp) => resp.json()).then((json) => {
			if (!json || !json[1] || !json[1].result || !json[1].result[0] || !json[1].result[0].uri) {
				// TODO(john): better error handling.
				return null;
			}
			const respUri = json[1].result[0].uri.split("git://")[1];
			const prt0Uri = respUri.split("?");
			const prt1Uri = prt0Uri[1].split("#");

			const repoUri = prt0Uri[0];
			const frevUri = (repoUri === repoRevSpec.repoURI ? repoRevSpec.rev : prt1Uri[0]) || "master"; // TODO(john): preserve rev branch
			const pathUri = prt1Uri[1];
			const startLine = parseInt(json[1].result[0].range.start.line, 10) + 1;
			const startChar = parseInt(json[1].result[0].range.start.character, 10) + 1;

			let lineAndCharEnding = "";
			if (startLine && startChar) {
				lineAndCharEnding = `#L${startLine}:${startChar}`;
			} else if (startLine) {
				lineAndCharEnding = `#L${startLine}`;
			}

			j2dCache[cacheKey] = `${sourcegraphUrl}/${repoUri}@${frevUri}/-/blob/${pathUri}${lineAndCharEnding}`;
			return j2dCache[cacheKey];
		});
}

export function fetchReferences(col: number, path: string, line: number, repoRevSpec: RepoRevSpec): Promise<Reference[] | null> {
	const ext = getPathExtension(path);
	if (!supportedExtensions.has(ext)) {
		return Promise.resolve(null);
	}
	const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${col}`;
	if (referencesCache[cacheKey]) {
		return Promise.resolve(referencesCache[cacheKey]);
	}

	const body = wrapLSP({
		method: "textDocument/references",
		params: {
			textDocument: {
				uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
			},
			position: {
				character: col - 1,
				line: line - 1,
			},
		},
		context: {
			includeDeclaration: true,
		},
	} as any, repoRevSpec, path);

	return fetch(`${sourcegraphUrl}/.api/xlang/textDocument/references`, { method: "POST", body: JSON.stringify(body) })
		.then((resp) => resp.json()).then((json) => {
			if (!json || !json[1] || !json[1].result) {
				// TODO(john): better error handling.
				return null;
			}

			referencesCache[cacheKey] = json[1].result;
			return referencesCache[cacheKey];
		});
}

export function fetchXreferences(col: number, path: string, line: number, repoRevSpec: RepoRevSpec, hints: any, limit: number, query: number): Promise<Reference[] | null> {
	const ext = getPathExtension(path);
	if (!supportedExtensions.has(ext)) {
		return Promise.resolve(null);
	}
	const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${col}`;
	// if (referencesCache[cacheKey]) {
	// 	return Promise.resolve(referencesCache[cacheKey]);
	// }

	const body = wrapLSP({
		method: "textDocument/xreferences",
		params: {
			hints,
			query,
			limit,
		},
	}, repoRevSpec, path);

	return fetch(`${sourcegraphUrl}/.api/xlang/textDocument/xreferences`, { method: "POST", body: JSON.stringify(body) })
		.then((resp) => resp.json()).then((json) => {
			if (!json || !json[1] || !json[1].result) {
				// TODO(john): better error handling.
				return null;
			}

			referencesCache[cacheKey] = json[1].result;
			return referencesCache[cacheKey];
		});
}
