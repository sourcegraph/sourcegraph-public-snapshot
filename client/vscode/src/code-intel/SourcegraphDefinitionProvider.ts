import * as Comlink from 'comlink'
import { EMPTY, of } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import * as vscode from 'vscode'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { makeRepoURI, parseRepoURI } from '@sourcegraph/shared/src/util/url'

import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'
import { SourcegraphVSCodeExtensionHostAPI } from '../webview/contract'

export class SourcegraphDefinitionProvider implements vscode.DefinitionProvider {
    constructor(
        private readonly fs: SourcegraphFileSystemProvider,
        private readonly sourcegraphExtensionHostAPI: Comlink.Remote<SourcegraphVSCodeExtensionHostAPI>
    ) {}
    public async provideDefinition(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Definition | undefined> {
        const uri = this.fs.sourcegraphUri(document.uri)
        const extensionHostUri = makeRepoURI({
            repoName: uri.repositoryName,
            revision: uri.revision,
            filePath: uri.path,
        })

        const definitions = wrapRemoteObservable(
            this.sourcegraphExtensionHostAPI.getDefinition({
                textDocument: {
                    uri: extensionHostUri,
                },
                position: {
                    line: position.line,
                    character: position.character,
                },
            })
        )
            .pipe(
                switchMap(({ isLoading, result }) => {
                    if (isLoading) {
                        return EMPTY
                    }

                    const locations = result.map(location => {
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
                first()
            )
            .toPromise()

        token.onCancellationRequested(() => {
            // TODO unsubscribe, manually create promise
        })

        return definitions
    }
}
