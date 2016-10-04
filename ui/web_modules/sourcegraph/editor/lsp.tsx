import {URIUtils} from "sourcegraph/core/uri";
import {defaultFetch as fetch} from "sourcegraph/util/xhr";

interface Position {
	line: number;
	character: number;
}

export function toPosition(pos: monaco.IPosition): Position {
	return {line: pos.lineNumber - 1, character: pos.column - 1};
}

interface Range {
	start: Position;
	end: Position;
}

export interface Location {
	uri: string;
	range: Range;
}

export function toMonacoLocation(loc: Location): monaco.languages.Location {
	return {
		uri: monaco.Uri.parse(loc.uri),
		range: toMonacoRange(loc.range),
	};
}

export function toMonacoRange(r: Range): monaco.IRange {
	return new monaco.Range(r.start.line + 1, r.start.character + 1, r.end.line + 1, r.end.character + 1);
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
export function send(model: monaco.editor.IReadOnlyModel, method: string, params: any): Promise<LSPResponse> {
	return sendExt(URIUtils.withoutFilePath(URIUtils.fromRefsDisplayURIMaybe(model.uri)).toString(true), model.getModeId(), method, params);
}

// sendExt mirrors the functionality of send, but is intended to be used by callers outside of Monaco.
export function sendExt(uri: string, modeID: string, method: string, params: any): Promise<LSPResponse> {
	// We duplicate the method in the URL and in the request body for
	// easier browser network tab debugging.
	return fetch(`/.api/xlang/${method}`, {
		method: "POST",
		body: JSON.stringify([
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
		]),
	})
		.then((resp: Response): Promise<LSPResponse[]> => {
			if (resp.status >= 200 && resp.status <= 299) { return resp.json(); }
			return resp.text().then((body) => [{error: {code: `HTTP ${resp.status}`, message: body}}]);
		})
		.then(resp => {
			if (resp && resp.length >= 2) {
				return resp[1];
			}
			return null;
		})
		.catch(err => {
			// LSP error
			if (err && err.message) {
				const jsonrpcErr = JSON.parse(err.message);
				if ((global as any).console.debug) {
					console.debug("LSP %s: %s\terror: %o params: %o", method, jsonrpcErr.message, jsonrpcErr, params); // tslint:disable-line no-console
				}
			}
			return null;
		});
}
