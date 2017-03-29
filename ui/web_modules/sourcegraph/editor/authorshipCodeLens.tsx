import URI from "vs/base/common/uri";
import { getLanguages, onLanguage, registerCodeLensProvider } from "vs/editor/browser/standalone/standaloneLanguages";
import { Range } from "vs/editor/common/core/range";
import { IReadOnlyModel } from "vs/editor/common/editorCommon";
import { Command, ICodeLensSymbol } from "vs/editor/common/modes";
import * as modes from "vs/editor/common/modes";

import { URIUtils } from "sourcegraph/core/uri";
import { timeFromNow } from "sourcegraph/util/dateFormatterUtil";
import { getModes } from "sourcegraph/util/features";
import { fetchGQL } from "sourcegraph/util/gqlClient";

export class AuthorshipCodeLens implements modes.CodeLensProvider {
	resolveCodeLens(model: IReadOnlyModel, codeLens: ICodeLensSymbol): ICodeLensSymbol | Thenable<ICodeLensSymbol> {
		return codeLens;
	}

	provideCodeLenses(model: IReadOnlyModel): ICodeLensSymbol[] | Thenable<ICodeLensSymbol[]> {
		return this.getBlameData(model.uri).then((blameData) => {
			let codeLenses: ICodeLensSymbol[] = [];
			if (!blameData || blameData.length === 0) {
				return codeLenses;
			}
			for (let i = 0; i < blameData.length; i++) {
				const blameLine = blameData[i];
				if (!blameLine.author || !blameLine.author.person) {
					return codeLenses;
				}
				const timeSince = timeFromNow(blameLine.author.date);
				codeLenses.push({
					id: `${blameLine.rev}${blameLine.startLine}-${blameLine.endLine}`,
					range: new Range(blameLine.startLine, 0, blameLine.endLine, Infinity),
					command: {
						id: "codelens.authorship.commit",
						title: `${blameLine.author.person.name} - ${timeSince}`,
						arguments: [Object.assign({}, blameLine)], // allow modification
					} as Command,
				});
			}
			return codeLenses;
		});
	}

	private getBlameData(resource: URI): Thenable<GQL.IHunk[]> {
		const { repo, rev, path } = URIUtils.repoParams(resource);
		return fetchGQL(`query getBlameData($repo: String, $rev: String, $path: String) {
			root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						commit {
							file(path: $path) {
								blame(startLine: 0, endLine: 0) {
									rev
									startLine
									endLine
									startByte
									endByte
									message
									author {
										person {
											name
											email
											gravatarHash
										}
										date
									}
								}
							}
						}
					}
				}
			}
		}`, { repo, rev, path }).then(query => {
				const root = query.data.root;
				const commit = root.repository && root.repository.commit.commit;
				if (!commit || !commit.file) {
					return;
				}
				return commit.file.blame;
			});
	}

}

getLanguages().forEach(({ id }) => {
	// id should just be plaintext
	onLanguage(id, () => {
		registerCodeLensProvider(id, new AuthorshipCodeLens());
	});
});
getModes().forEach(mode => {
	onLanguage(mode, () => {
		registerCodeLensProvider(mode, new AuthorshipCodeLens());
	});
});
registerCodeLensProvider("markdown", new AuthorshipCodeLens());
