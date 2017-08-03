import { fetchDependencyReferences } from "sourcegraph/backend";
import { fetchReferences, fetchXdefinition, fetchXreferences } from "sourcegraph/backend/lsp";
import { addReferences, ReferencesContext, setReferences, setReferencesLoad, setXReferencesLoad, store as referencesStore } from "sourcegraph/references/store";

const contextFetches = new Set<string>();

export function triggerReferences(context: ReferencesContext): void {
	const repoRevSpec = {
		repoURI: context.loc.uri,
		rev: context.loc.rev,
	};

	setReferences({ ...referencesStore.getValue(), context });

	// HACK(john): prevent double fetching (as this will add duplicate references to our store).
	const fetchKey = JSON.stringify(context);
	if (!contextFetches.has(fetchKey)) {
		setReferencesLoad(context.loc, "pending");
		setXReferencesLoad(context.loc, "pending");
		fetchReferences(context.loc.char - 1, context.loc.path, context.loc.line - 1, repoRevSpec)
			.then((references) => {
				if (references) {
					addReferences(context.loc, references);
				}
				setReferencesLoad(context.loc, "completed");
			})
			.catch(() => {
				setReferencesLoad(context.loc, "completed");
			});
		fetchXdefinition(context.loc.char - 1, context.loc.path, context.loc.line - 1, repoRevSpec)
			.then(defInfo => {
				if (!defInfo) { throw new Error("no xrefs"); }

				fetchDependencyReferences(repoRevSpec.repoURI, repoRevSpec.rev, context.loc.path, 40, 25).then((data) => {
					if (!data || !data.repoData.repos) { throw new Error("no xrefs"); }
					const idToRepo = (id: number): any => {
						const i = data.repoData.repoIds.indexOf(id);
						if (i === -1) { throw new Error("repo id not found"); }
						return data.repoData.repos[i];
					};

					const retVal = data.dependencyReferenceData.references.map(ref => {
						const repo = idToRepo(ref.repoId);
						const commit = repo.lastIndexedRevOrLatest.commit;
						const workspace = commit ? { uri: repo.uri, rev: repo.lastIndexedRevOrLatest.commit.sha1 } : undefined;
						return {
							workspace,
							hints: ref.hints ? JSON.parse(ref.hints) : {},
						};
					}).filter(dep => dep.workspace); // possibly slice to MAX_DEPENDENT_REPOS (10)

					return retVal;
				}).then(dependents => {
					if (!dependents) {
						throw new Error("no xrefs"); // no results, map below would fail.
					}
					Promise.all(dependents.map(dependent => {
						// const refs2Locations = (references: ReferenceInformation[]): vscode.Location[] => {
						// 	return references.map(r => this.currentWorkspaceClient.protocol2CodeConverter.asLocation(r.reference));
						// };
						// const params: WorkspaceReferencesParams = { query: defInfo.symbol, hints: dependent.hints, limit: 50 };
						return fetchXreferences(dependent.workspace, context.loc.path, defInfo.symbol, dependent.hints, 50).then((refs) => {
							if (refs) {
								addReferences(context.loc, refs);
							}
						});
					})).then(() => {
						setXReferencesLoad(context.loc, "completed");
					});
				}).catch(() => {
					setXReferencesLoad(context.loc, "completed");
				});
			}).catch(() => {
				setXReferencesLoad(context.loc, "completed");
			});
	}
	contextFetches.add(JSON.stringify(context));
}
