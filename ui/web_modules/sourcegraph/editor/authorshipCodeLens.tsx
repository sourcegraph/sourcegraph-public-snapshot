import * as moment from "moment";
import URI from "vs/base/common/uri";
import { Range } from "vs/editor/common/core/range";
import { IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Command, ICodeLensSymbol } from "vs/editor/common/modes";
import * as modes from "vs/editor/common/modes";

import { URIUtils } from "sourcegraph/core/uri";
import { codeLensCache } from "sourcegraph/editor/EditorService";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

export class AuthorshipCodeLens implements modes.CodeLensProvider {
	resolveCodeLens(model: IReadOnlyModel, codeLens: ICodeLensSymbol): ICodeLensSymbol | Thenable<ICodeLensSymbol> {
		return codeLens;
	}

	provideCodeLenses(model: IReadOnlyModel): ICodeLensSymbol[] | Thenable<ICodeLensSymbol[]> {
		return this.getBlameData(model.uri).then((blameData) => {
			let codeLenses: ICodeLensSymbol[] = [];
			for (let i = 0; i < blameData.length; i++) {
				const blameLine = blameData[i];
				const timeSince = moment(blameLine.date, "YYYY-MM-DDThh:mmTZD").fromNow();
				codeLenses.push({
					id: `${blameLine.rev}${blameLine.startLine}-${blameLine.endLine}`,
					range: new Range(blameLine.startLine, 0, blameLine.endLine, Infinity),
					command: {
						id: "codelens.authorship.commit",
						title: `${blameLine.name} - ${timeSince}`,
						arguments: [blameLine],
					} as Command,
				});
			}
			return codeLenses;
		});
	}

	private getBlameData(resource: URI): Thenable<GQL.IHunk[]> {
		const key = resource.toString(true);
		const { repo, rev, path } = URIUtils.repoParams(resource);
		let cachedLens = codeLensCache.get(key);
		if (cachedLens) {
			return Promise.resolve(cachedLens);
		}
		return fetchGraphQLQuery(`query Content($repo: String, $rev: String, $path: String) {
			root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						commit {
							file(path: $path) {
								blame(startLine: 0, endLine: 0) {
									name
									email
									rev
									date
									startLine
									endLine
									startByte
									endByte
									message
									gravatarHash
								}
							}
						}
					}
				}
			}
		}`, { repo, rev, path }).then((query) => {
				const commit = query.root.repository && query.root.repository.commit.commit;
				if (!commit || !commit.file) {
					return;
				}
				return commit.file.blame;
			});
	}

}
