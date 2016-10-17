import {context} from "sourcegraph/app/context";
import {URIUtils} from "sourcegraph/core/uri";
import {defaultFetch as fetch} from "sourcegraph/util/xhr";
import URI from "vs/base/common/uri";
import {Range} from "vs/editor/common/core/range";
import {IPosition, IRange, IReadOnlyModel} from "vs/editor/common/editorCommon";
import {Location as VSCLocation} from "vs/editor/common/modes";

interface LSPPosition {
	line: number;
	character: number;
}

export function toPosition(pos: IPosition): LSPPosition {
	return {line: pos.lineNumber - 1, character: pos.column - 1};
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
	return {
		uri: URI.parse(loc.uri),
		range: toMonacoRange(loc.range),
	};
}

export function toMonacoRange(r: LSPRange): IRange {
	return new Range(r.start.line + 1, r.start.character + 1, r.end.line + 1, r.end.character + 1);
}

type LSPResponse = {
	method: string;
	result: any;
	error: {code: number, message: string};
};

// send sends an LSP request with the given method and params. Because
// it's sending it statelessly over HTTP, it bundles the LSP
// "initialize" params into each request. The server is responsible
// for managing the lifecycle of the LSP servers; this client treats
// it as a stateless service.
export function send(model: IReadOnlyModel, method: string, params: any): Promise<LSPResponse> {
	return sendExt(URIUtils.withoutFilePath(URIUtils.fromRefsDisplayURIMaybe(model.uri)).toString(true), model.getModeId(), method, params);
}

// sendExt mirrors the functionality of send, but is intended to be used by callers outside of Monaco.
export function sendExt(uri: string, modeID: string, method: string, params: any): Promise<LSPResponse> {
	const t0 = Date.now();

	const body = [
		{
			id: 0,
			method: "initialize",
			params: {
				rootPath: uri,
				mode: modeID,
			},
		},
		{id: 1, method, params},
		{id: 2, method: "shutdown"},
		{method: "exit"},
	];

	type LSPResponseWithTraceURL = {
		resp: LSPResponse;
		traceURL: string; // Lightstep trace URL
	};

	// We duplicate the method in the URL and in the request body for
	// easier browser network tab debugging.
	return fetch(`/.api/xlang/${method}`, {
		method: "POST",
		body: JSON.stringify(body),
	})
		.then((resp: Response): Promise<LSPResponseWithTraceURL> => {
			const traceURL = resp.headers.get("x-trace");

			if (resp.status >= 200 && resp.status <= 299) {
				// Pass along the main request's response (not the
				// initialize/shutdown responses).
				return resp.json().then((resps: LSPResponse[]) => ({resp: resps[1], traceURL}));
			}

			// Synthesize an LSP response object from the HTTP error,
			// so that we can deal with errors from any source in a
			// consistent way.
			return resp.text().then((text: string) => ({
				resp: {
					error: {
						code: resp.status,
						message: `HTTP error ${resp.status} (${resp.statusText}): ${text || "(empty HTTP response body)"}`,
					},
				},
				traceURL,
			}));
		})
		.then((respWithTrace: LSPResponseWithTraceURL): LSPResponse | null => {
			logLSPResponse(URI.parse(uri), modeID, method, params, body, respWithTrace.resp, respWithTrace.traceURL, Date.now() - t0);

			// Report LSP error.
			if (respWithTrace.resp.error) {
				return null;
			}
			return respWithTrace.resp;
		});
}

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
	console.debug("LSP request params: %o", params);
	console.debug("LSP response: %o", resp);
	console.debug("Trace: %s", traceURL);

	const reproCmd = `curl --data '${JSON.stringify(reqBody).replace(/'/g, "\\'")}' ${window.location.protocol}//${window.location.host}/.api/xlang/${method}`;
	console.debug("Repro command (public repos only):\n%c%s", "background-color:#333;color:#aaa", reproCmd);

	const locHash = params.textDocument && params.textDocument.uri && params.position ? `#L${params.position.line + 1}:${params.position.character + 1}` : "";
	const pageURL = window.location.href.replace(/(#L[\d:-]*)?$/, locHash);

	const {repo} = URIUtils.repoParams(uri);
	const issueTitle = `${err ? "Error in" : "Unexpected behavior from"} ${method} in ${repo}`;
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
* Feature flags: \`${JSON.stringify(context.features)}\`
* [Lightstep trace](${traceURL})
* Round-trip time: ${rttMsec}ms`;
	console.debug(`Post a GitHub issue\nhttps://github.com/sourcegraph/sourcegraph/issues/new?title=${encodeURIComponent(issueTitle)}&body=${encodeURIComponent(issueBody)}&labels=lang-${modeID}`);
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
	return `${method} with ${JSON.stringify(params)}`;
}

function truncate(s: string, max: number): string {
	if (s.length <= max) {
		return s;
	}
	return `${s.slice(0, max)}... (truncated, ${s.length - max} characters remaining)`;
}
