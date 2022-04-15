import * as Comlink from 'comlink'
import { EMPTY, of } from 'rxjs'
import { debounceTime, first, switchMap } from 'rxjs/operators'
import * as vscode from 'vscode'

import { finallyReleaseProxy, wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { makeRepoURI, parseRepoURI } from '@sourcegraph/shared/src/util/url'

import { SearchSidebarAPI } from '../contract'
import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'

export class SourcegraphReferenceProvider implements vscode.ReferenceProvider {
    constructor(
        private readonly fs: SourcegraphFileSystemProvider,
        private readonly sourcegraphExtensionHostAPI: Comlink.Remote<SearchSidebarAPI>
    ) {}
    public async provideReferences(
        document: vscode.TextDocument,
        position: vscode.Position,
        referenceContext: vscode.ReferenceContext,
        token: vscode.CancellationToken
    ): Promise<vscode.Location[] | undefined> {
        const uri = this.fs.sourcegraphUri(document.uri)
        const extensionHostUri = makeRepoURI({
            repoName: uri.repositoryName,
            revision: uri.revision,
            filePath: uri.path,
        })

        const definitions = wrapRemoteObservable(
            this.sourcegraphExtensionHostAPI.getReferences(
                {
                    textDocument: {
                        uri: extensionHostUri,
                    },
                    position: {
                        line: position.line,
                        character: position.character,
                    },
                },
                referenceContext
            )
        )
            .pipe(
                finallyReleaseProxy(),
                switchMap(({ isLoading, result }) => {
                    if (isLoading) {
                        return EMPTY
                    }

                    const locations = result.map(location => {
                        // Create a sourcegraph URI from this git URI (so we need both fromGitURI and toGitURI.)`
                        const uri = parseRepoURI(location.uri)

                        return this.fs.toVscodeLocation({
                            resource: {
                                path: uri.filePath ?? '',
                                repositoryName: uri.repoName,
                                revision: uri.commitID ?? uri.revision ?? '',
                            },
                            range: location.range,
                        })
                    })

                    return of(locations)
                }),
                debounceTime(1000),
                first()
            )
            .toPromise()

        token.onCancellationRequested(() => {
            // Debt: manually create promise so we can cancel request.
        })

        return definitions
    }
}
