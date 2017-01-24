import * as hash from "object-hash";
import { context } from "sourcegraph/app/context";
import { URIUtils } from "sourcegraph/core/uri";
import { Features } from "sourcegraph/util/features";
import { defaultFetch } from "sourcegraph/util/xhr";
import URI from "vs/base/common/uri";
import { Range } from "vs/editor/common/core/range";
import { IPosition, IRange, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Location as VSCLocation } from "vs/editor/common/modes";

interface LSPPosition {
	line: number;
	character: number;
}

export function toPosition(pos: IPosition): LSPPosition {
	return { line: pos.lineNumber - 1, character: pos.column - 1 };
}

interface LSPRange {
	start: LSPPosition;
	end: LSPPosition;
}

export interface Location {
	uri: string;
	range: LSPRange;
}

export function toMonacoLocation(loc: Location): VSCLocation {
	if (!loc.range) {
		throw new Error(`location range is not defined: ${JSON.stringify(loc)}`);
	}
	return {
		uri: URI.parse(loc.uri),
		range: toMonacoRange(loc.range),
	};
}

export function toMonacoRange(r: LSPRange): IRange {
	if (!r) {
		throw new Error("range is not defined");
	}
	return new Range(r.start.line + 1, r.start.character + 1, r.end.line + 1, r.end.character + 1);
}

type LSPResponse = {
	method: string;
	result?: any;
	error?: { code: number, message: string };
};

// send sends an LSP request with the given method and params. Because
// it's sending it statelessly over HTTP, it bundles the LSP
// "initialize" params into each request. The server is responsible
// for managing the lifecycle of the LSP servers; this client treats
// it as a stateless service.
export function send(model: IReadOnlyModel, method: string, params: any): Promise<LSPResponse> {
	return sendExt(URIUtils.withoutFilePath(model.uri).toString(true), model.getModeId(), method, params);
}

const LSPCache = new Map<string, Promise<Response>>();

// WARNING: Caches responses that are errors.
async function cachingFetch(url: string | Request, init?: RequestInit): Promise<Response> {
	const key = hash([url, init]);

	let promise = LSPCache.get(key);
	if (!promise) {
		promise = defaultFetch(url, init);
	}

	LSPCache.set(key, promise);
	const response = await promise;
	return response.clone();
}

// sendExt mirrors the functionality of send, but is intended to be used by callers outside of Monaco.
export async function sendExt(uri: string, modeID: string, method: string, params: any): Promise<LSPResponse> {
	const t0 = Date.now();

	const request = [
		{
			id: 0,
			method: "initialize",
			params: {
				rootPath: uri,
				initializationOptions: { mode: modeID },
				mode: modeID, // DEPRECATED: use old mode field if new one is not set
			},
		},
		{ id: 1, method, params },
		{ id: 2, method: "shutdown" },
		{ method: "exit" },
	];

	// We duplicate the method in the URL and in the request body for
	// easier browser network tab debugging.
	const serverResponse = await cachingFetch(`/.api/xlang/${method}`, {
		method: "POST",
		body: JSON.stringify(request),
	});
	const traceURL = serverResponse.headers.get("x-trace");

	let response;
	if (serverResponse.status >= 200 && serverResponse.status < 300) {
		const resps = await serverResponse.json();
		response = resps[1];
	} else {
		// Synthesize an LSP response object from the HTTP error,
		// so that we can deal with errors from any source in a
		// consistent way.
		const text = await serverResponse.text();
		response = {
			error: {
				code: serverResponse.status,
				message: `HTTP error ${serverResponse.status} (${serverResponse.statusText}): ${text || "(empty HTTP response body)"}`,
			},
		};
	}
	logLSPResponse(URI.parse(uri), modeID, method, params, request, response, traceURL, Date.now() - t0);

	return response;
}

const modeToIssueAssignee = {
	go: "keegancsmith",
	typescript: "beyang",
	javascript: "beyang",
	java: "akhleung",
	python: "renfredxh",
};

function logLSPResponse(uri: URI, modeID: string, method: string, params: any, reqBody: any, resp: LSPResponse, traceURL: string, rttMsec: number): void {
	if (!(global as any).console || !(global as any).console.group || !(global as any).console.debug) {
		return;
	}

	// tslint:disable: no-console

	const console = (global as any).console;
	const err = resp.error;
	console.groupCollapsed(
		"%c%s%c LSP %s %s %c%sms",
		`background-color:${err ? "red" : "green"};color:white`, err ? "ERR" : "OK",
		`background-color:inherit;color:inherit;`,
		modeID,
		describeRequest(method, params),
		"color:#999;font-weight:normal;font-style:italic",
		rttMsec
	);
	if (Features.trace.isEnabled()) {
		console.debug("Trace: %s", traceURL);
	}
	console.debug("LSP request params: %O", params);
	console.debug("LSP response: %O", resp);

	const reproCmd = `curl --data '${JSON.stringify(reqBody).replace(/'/g, "\\'")}' ${window.location.protocol}//${window.location.host}/.api/xlang/${method}`;
	console.debug("Repro command (public repos only):\n%c%s", "background-color:#333;color:#aaa", reproCmd);

	const locHash = params.textDocument && params.textDocument.uri && params.position ? `#L${params.position.line + 1}:${params.position.character + 1}` : "";
	const pageURL = window.location.href.replace(/(#L[\d:-]*)?$/, locHash);

	const { repo } = URIUtils.repoParams(uri);
	const issueTitle = `${err ? "Error in" : "Unexpected behavior from"} ${method} in ${repo}`;
	const assignee = modeToIssueAssignee[modeID];
	const issueBody = `I saw ${err ? "an error in" : "unexpected behavior from"} from LSP ${method} on a ${modeID} file at [${pageURL}](${pageURL}).

**Repro:**
\`\`\`bash
${reproCmd}
\`\`\`

**Actual:**
\`\`\`json
${truncate(JSON.stringify(resp, null, 2), 300)}
\`\`\`

**Expected:** ______________________________

---

* Location: [${pageURL}](${pageURL})
* User: ${context.user ? context.user.Login : "(anonymous)"}
* User agent: \`${window.navigator.userAgent}\`
* Deployed site version: ${context.buildVars.Version} (${context.buildVars.Date})
* [Lightstep trace](${traceURL})
* Round-trip time: ${rttMsec}ms`;
	console.debug(`Copy and send this URL to support@sourcegraph.com to report an issue:\nhttps://github.com/sourcegraph/sourcegraph/issues/new?title=${encodeURIComponent(issueTitle)}&body=${encodeURIComponent(issueBody)}&labels[]=lang-${modeID}&labels[]=${encodeURIComponent("Component: xlang")}&labels[]=${encodeURIComponent("Type: Bug")}${assignee ? `&assignee=${assignee}` : ""}`);
	console.groupEnd();

	// tslint:enable: no-console
}

function describeRequest(method: string, params: any): string {
	if (params.textDocument && params.textDocument.uri && params.position) {
		return `${method} @ ${params.position.line + 1}:${params.position.character + 1}`;
	}
	if (typeof params.query !== "undefined") {
		return `${method} with query ${JSON.stringify(params.query)}`;
	}
	return method;
}

function truncate(s: string, max: number): string {
	if (s.length <= max) {
		return s;
	}
	return `${s.slice(0, max)}... (truncated, ${s.length - max} characters remaining)`;
}
