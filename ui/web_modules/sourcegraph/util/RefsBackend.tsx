import { IRange, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Location } from "vs/editor/common/modes";

import { URIUtils } from "sourcegraph/core/uri";
import * as lsp from "sourcegraph/editor/lsp";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { TimeFromNowUntil } from "sourcegraph/util/dateFormatterUtil";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";
import { ReferencesModel } from "sourcegraph/workbench/info/referencesModel";

import * as _ from "lodash";

export interface RefData {
	language: string;
	repo: string;
	version: string;
	file: string;
	line: number;
	column: number;
}

export interface DefinitionData {
	definition: {
		uri: string;
		range: IRange;
	};
	docString: string;
	funcName: string;
}

export interface ReferenceCommitInfo {
	loc: Location;
	hunk: GQL.IHunk;
}

export async function provideDefinition(model: IReadOnlyModel, pos: { line: number, character: number }): Promise<DefinitionData | null> {
	const hoverInfo = await lsp.send(model, "textDocument/hover", {
		textDocument: { uri: model.uri.toString(true) },
		position: pos,
		context: { includeDeclaration: false },
	});

	if (!hoverInfo || !hoverInfo.result || !hoverInfo.result.contents) {
		return null;
	}
	const [{value: funcName}, docs] = hoverInfo.result.contents;
	const docString = docs ? docs.value : "";

	const defResponse = await lsp.send(model, "textDocument/definition", {
		textDocument: { uri: model.uri.toString(true) },
		position: pos,
		context: { includeDeclaration: false },
	});

	if (!defResponse.result || !defResponse.result[0]) {
		return null;
	}

	const defFirst = defResponse.result[0];
	let definition = {
		uri: defFirst.uri,
		range: {
			startLineNumber: defFirst.range.start.line,
			startColumn: defFirst.range.start.character,
			endLineNumber: defFirst.range.end.line,
			endColumn: defFirst.range.end.character,
		}
	};
	return { funcName, docString, definition };
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

// TODO/Checkpoint: @Kingy to refine implementation for Project WOW
export async function provideGlobalReferences(object: any): Promise<any> {
	const references = object.Data.References;
	const repoData = object.RepoData;
	if (!references) {
		return;
	}

	let promises = references.map(reference => {
		let newModel = repoData[reference.RepoID];
		let repoURI = URIUtils.pathInRepo(newModel.URI, newModel.DefaultBranch, "");
		return lsp.sendExt(repoURI.toString(), newModel.Language.toLowerCase(), "workspace/xreferences", {
			query: object.Data.Location.symbol,
			hints: reference.Hints,
		}).then(resp => {
			return resp.result.map(ref => {
				let loc: lsp.Location = ref.reference;
				return loc;
			});
		});
	});

	const allReferences = await Promise.all(promises);
	return _.flatten(allReferences).map(lsp.toMonacoLocation);
}

// TODO/Checkpoint: @Kingy to refine implementation for Project WOW
export async function fetchDependencyReferencesReferences(model: IReadOnlyModel, pos: { line: number, character: number }): Promise<any> {
	let refModelQuery =
		`repository(uri: "${model.uri.authority}${model.uri.path}") {
			commit(rev: "${model.uri.query}") {
				commit {
					file(path: "${model.uri.fragment}") {
						dependencyReferences(Language: "${model.getModeId()}", Line: ${pos.line}, Character: ${pos.character}) {
							data
						}
					}
				}
			}
		}`;

	const query =
		`query {
			root {
				${refModelQuery}
			}
		}`;

	let data = await fetchGraphQLQuery(query);
	if (!data.root.repository || !data.root.repository.commit.commit || !data.root.repository.commit.commit.file ||
		!data.root.repository.commit.commit.file.dependencyReferences || !data.root.repository.commit.commit.file.dependencyReferences.data.length) {
		return null;
	}
	let object = JSON.parse(data.root.repository.commit.commit.file.dependencyReferences.data);
	return object;
}
