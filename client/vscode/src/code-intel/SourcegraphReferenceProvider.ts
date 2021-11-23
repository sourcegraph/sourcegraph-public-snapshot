import * as Comlink from 'comlink'
import { EMPTY, of } from 'rxjs'
import { debounceTime, first, switchMap } from 'rxjs/operators'
import * as vscode from 'vscode'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { makeRepoURI, parseRepoURI } from '@sourcegraph/shared/src/util/url'

import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'
import { SourcegraphVSCodeExtensionHostAPI } from '../webview/contract'

export class SourcegraphReferenceProvider implements vscode.ReferenceProvider {
    constructor(
        private readonly fs: SourcegraphFileSystemProvider,
        private readonly sourcegraphExtensionHostAPI: Comlink.Remote<SourcegraphVSCodeExtensionHostAPI>
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
                // TODO can share this code w/ definition provider.
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
                // TODO validate that this is OK, and actually unsubscribe when token emits event.
                debounceTime(1000),
                first()
            )
            .toPromise()

        token.onCancellationRequested(() => {
            // TODO unsubscribe, manually create promise
        })

        return definitions
    }
}
