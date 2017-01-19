import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { ICommonCodeEditor, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Location } from "vs/editor/common/modes";

import { URIUtils } from "sourcegraph/core/uri";
import * as lsp from "sourcegraph/editor/lsp";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { TimeFromNowUntil } from "sourcegraph/util/dateFormatterUtil";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";

export interface RefData {
	language: string;
	repo: string;
	version: string;
	file: string;
	line: number;
	column: number;
}

export interface ReferenceCommitInfo {
	loc: Location;
	hunk: GQL.IHunk;
}

export async function provideReferences(model: IReadOnlyModel, pos: { line: number, character: number }): Promise<Location[]> {
	return lsp.send(model, "textDocument/references", {
		textDocument: { uri: model.uri.toString(true) },
		position: pos,
		context: { includeDeclaration: false },
	})
		.then(resp => resp ? resp.result : null)
		.then((resp: lsp.Location | lsp.Location[] | null) => {
			if (!resp || Object.keys(resp).length === 0) {
				return null;
			}

			const {repo, rev, path} = URIUtils.repoParams(model.uri);
			AnalyticsConstants.Events.CodeReferences_Viewed.logEvent({ repo, rev: rev || "", path });

			const locs: lsp.Location[] = resp instanceof Array ? resp : [resp];
			return locs.map(lsp.toMonacoLocation);
		});
}

export async function provideReferencesCommitInfo(referencesModel: ReferencesModel): Promise<ReferencesModel> {
	let refModelQuery: string = "";
	referencesModel.references.forEach(reference => {
		refModelQuery = refModelQuery +
			`${reference.id.replace("#", "")}:repository(uri: "${reference.uri.authority}${reference.uri.path}") {
				commit(rev: "${reference.uri.query}") {
					commit {
						file(path: "${reference.uri.fragment}") {
							blame(startLine: ${reference.range.startLineNumber}, endLine: ${reference.range.endLineNumber}) {
								rev
								startLine
								endLine
								startByte
								endByte
								message
								author {
									person {
										gravatarHash
										name
										email
									}
									date
								}
							}
		  				}
					}
				}
			}`;
	});

	const query =
		`query {
			root {
				${refModelQuery}
			}
		}`;

	let data = await fetchGraphQLQuery(query);
	if (!data.root) {
		return referencesModel;
	}

	referencesModel.references.forEach(reference => {
		let dataByRefID = data.root[reference.id.replace("#", "")];
		let hunk: GQL.IHunk = dataByRefID.commit.commit.file.blame[0];
		if (!hunk || !hunk.author || !hunk.author.person) {
			return;
		}
		hunk.author.date = TimeFromNowUntil(hunk.author.date, 14);
		let commitInfo = {
			loc: {
				uri: reference.uri,
				range: reference.range,
			},
			hunk: hunk,
		} as ReferenceCommitInfo;
		reference.info = commitInfo;
	});

	return referencesModel;
}

export async function provideGlobalReferences(editor: ICommonCodeEditor): TPromise<ReferencesModel> {
	const { repo, path } = URIUtils.repoParams(editor.getModel().uri);
	const editorPosition = editor.getPosition();
	let model = editor.getModel();
	let refData: RefData = {
		language: model.getModeIdAtPosition(editorPosition.lineNumber, editorPosition.column),
		repo: repo,
		version: model.uri.query,
		file: path,
		line: editorPosition.lineNumber - 1,
		column: editorPosition.column - 1,
	};
	// Why is this not an LSP request?
	const globalRefs = await fetchGlobalReferences(refData);
	let globalRefLocs: Location[] = [];
	globalRefs.forEach(ref => {
		if (!ref.refLocation || !ref.uri) {
			return;
		}
		globalRefLocs.push({
			range: {
				startLineNumber: ref.refLocation.startLineNumber + 1,
				startColumn: ref.refLocation.startColumn + 1,
				endLineNumber: ref.refLocation.endLineNumber + 1,
				endColumn: ref.refLocation.endColumn + 1,
			},
			uri: URI.from(ref.uri),
		});
	});

	return TPromise.as(new ReferencesModel(globalRefLocs));
}

// TODO: @Kingy to implement for Project WOW
async function fetchGlobalReferences(ref: RefData): Promise<Array<GQL.IRefFields>> {
	return [];
}
