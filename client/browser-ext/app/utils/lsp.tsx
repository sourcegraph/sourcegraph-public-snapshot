import { doFetch as fetch } from "../backend/xhr";
import { getEventLogger } from "../utils/context";
import { RepoRevSpec } from "./annotations";
import * as utils from "./index";
import * as tooltips from "./tooltips";

// TODO(uforic): Use consts. Also, these caches don't eject over time, which could be a problem.
let tooltipCache: { [key: string]: tooltips.TooltipData } = {};
let j2dCache = {};

function wrapLSP(req: { method: string, params: Object }, repoRevSpec: RepoRevSpec, path: string): Object[] {
	// TODO(uforic): Do { ...req, id: 1 } instead of casting above
	(req as any).id = 1;
	return [
		{
			id: 0,
			method: "initialize",
			params: {
				rootPath: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}`,
				mode: `${utils.getModeFromExtension(utils.getPathExtension(path))}`,
			},
		},
		req,
		{
			id: 2,
			method: "shutdown",
		},
		{
			method: "exit",
		},
	];
}

const prewarmCache = new Set<string>();
export function prewarmLSP(path: string, repoRevSpec: RepoRevSpec): void {
	const uri = `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`;
	if (prewarmCache.has(uri)) {
		return;
	}
	prewarmCache.add(uri);

	const body = wrapLSP({
		method: "textDocument/hover?prepare",
		params: {
			textDocument: { uri },
			position: {
				character: 0,
				line: 0,
			},
		},
	}, repoRevSpec, path);

	fetch("https://sourcegraph.com/.api/xlang/textDocument/hover?prepare", { method: "POST", body: JSON.stringify(body) });
}

export function getTooltip(target: HTMLElement, path: string, line: number, repoRevSpec: RepoRevSpec): Promise<tooltips.TooltipData> {
	const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${target.dataset["byteoffset"]}`;

	//TODO(uforic): Can this just be if(tooltipCache[cacheKey])?
	if (typeof tooltipCache[cacheKey] !== "undefined") {
		return Promise.resolve(tooltipCache[cacheKey]);
	}

	const body = wrapLSP({
		method: "textDocument/hover",
		params: {
			textDocument: {
				uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
			},
			position: {
				character: parseInt(target.dataset["byteoffset"], 10) - 1,
				line: line - 1,
			},
		},
	}, repoRevSpec, path);

	return fetch("https://sourcegraph.com/.api/xlang/textDocument/hover", { method: "POST", body: JSON.stringify(body) })
		.then((resp) => resp.json()).then((json) => {
			if (json[1].result && json[1].result.contents && json[1].result.contents.length > 0) {
				const title = json[1].result.contents[0].value;
				let doc;
				// TODO(uforic): apparently you are only interested in one result, so use array.find() What about the case of a plain string
				json[1].result.contents.filter((markedString) => markedString.language === "markdown").forEach((content) => {
					// TODO(john): what if there is more than 1?
					doc = content.value;
				});
				tooltipCache[cacheKey] = { title, doc };
			} else {
				tooltipCache[cacheKey] = null;
			}
			return tooltipCache[cacheKey];
		});
}

export function fetchJumpURL(col: string, path: string, line: number, repoRevSpec: RepoRevSpec): Promise<string | null> {
	const cacheKey = `${path}@${repoRevSpec.rev}:${line}@${col}`;
	// TODO(uforic): can you do this better?
	if (typeof j2dCache[cacheKey] !== "undefined") {
		return Promise.resolve(j2dCache[cacheKey]);
	}

	const body = wrapLSP({
		method: "textDocument/definition",
		params: {
			textDocument: {
				uri: `git://${repoRevSpec.repoURI}?${repoRevSpec.rev}#${path}`,
			},
			position: {
				character: parseInt(col, 10) - 1,
				line: line - 1,
			},
		},
	}, repoRevSpec, path);

	return fetch("https://sourcegraph.com/.api/xlang/textDocument/definition", { method: "POST", body: JSON.stringify(body) })
		.then((resp) => resp.json()).then((json) => {
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

			j2dCache[cacheKey] = `https://sourcegraph.com/${repoUri}@${frevUri}/-/blob/${pathUri}${lineAndCharEnding}`;
			return j2dCache[cacheKey];
		});
}
