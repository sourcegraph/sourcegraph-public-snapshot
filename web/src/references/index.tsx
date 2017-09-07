import { fetchDependencyReferences } from 'sourcegraph/backend'
import { fetchReferences, fetchXdefinition, fetchXreferences } from 'sourcegraph/backend/lsp'
import { addReferences, setReferences, setReferencesLoad, setXReferencesLoad, store as referencesStore } from 'sourcegraph/references/store'
import { AbsoluteRepoPosition, makeRepoURI } from 'sourcegraph/repo'

const contextFetches = new Set<string>()

export function triggerReferences(ctx: AbsoluteRepoPosition): void {
    setReferences({ ...referencesStore.getValue(), context: ctx })

    // HACK(john): prevent double fetching (as this will add duplicate references to our store).
    const fetchKey = makeRepoURI(ctx)
    if (!contextFetches.has(fetchKey)) {
        setReferencesLoad(ctx, 'pending')
        setXReferencesLoad(ctx, 'pending')
        fetchReferences(ctx)
            .then(references => {
                if (references) {
                    addReferences(ctx, references)
                }
                setReferencesLoad(ctx, 'completed')
            })
            .catch(() => {
                setReferencesLoad(ctx, 'completed')
            })
        fetchXdefinition(ctx)
            .then(defInfo => {
                if (!defInfo) { throw new Error('no xrefs') }

                fetchDependencyReferences(ctx.repoPath, ctx.commitID, ctx.filePath, ctx.position.line - 1, ctx.position.char! - 1).then(data => {
                    if (!data || !data.repoData.repos) { throw new Error('no xrefs') }
                    const idToRepo = (id: number): any => {
                        const i = data.repoData.repoIds.indexOf(id)
                        if (i === -1) { throw new Error('repo id not found') }
                        return data.repoData.repos[i]
                    }

                    const retVal = data.dependencyReferenceData.references.map(ref => {
                        const repo = idToRepo(ref.repoId)
                        const commit = repo.lastIndexedRevOrLatest.commit
                        const workspace = commit ? { uri: repo.uri, rev: repo.lastIndexedRevOrLatest.commit.sha1 } : undefined
                        return {
                            workspace,
                            hints: ref.hints ? JSON.parse(ref.hints) : {}
                        }
                    }).filter(dep => dep.workspace) // possibly slice to MAX_DEPENDENT_REPOS (10)

                    return retVal
                }).then(dependents => {
                    if (!dependents) {
                        throw new Error('no xrefs') // no results, map below would fail.
                    }
                    return Promise.all(dependents.map(dependent => {
                        if (!dependent.workspace) {
                            return undefined
                        }
                        // const refs2Locations = (references: ReferenceInformation[]): vscode.Location[] => {
                        //     return references.map(r => this.currentWorkspaceClient.protocol2CodeConverter.asLocation(r.reference));
                        // };
                        // const params: WorkspaceReferencesParams = { query: defInfo.symbol, hints: dependent.hints, limit: 50 };
                        return fetchXreferences(dependent.workspace, ctx.filePath, defInfo.symbol, dependent.hints, 50).then(refs => {
                            if (refs) {
                                addReferences(ctx, refs)
                            }
                        })
                    })).then(() => {
                        setXReferencesLoad(ctx, 'completed')
                    })
                }).catch(() => {
                    setXReferencesLoad(ctx, 'completed')
                })
            }).catch(() => {
                setXReferencesLoad(ctx, 'completed')
            })
    }
    contextFetches.add(fetchKey)
}
