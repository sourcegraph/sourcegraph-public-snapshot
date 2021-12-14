import * as Comlink from 'comlink'
import { EMPTY, of } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import * as vscode from 'vscode'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'
import { SourcegraphVSCodeExtensionHostAPI } from '../webview/contract'

export class SourcegraphHoverProvider implements vscode.HoverProvider {
    constructor(
        private readonly fs: SourcegraphFileSystemProvider,
        private readonly sourcegraphExtensionHostAPI: Comlink.Remote<SourcegraphVSCodeExtensionHostAPI>
    ) {}
    public async provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Hover | undefined> {
        const uri = this.fs.sourcegraphUri(document.uri)
        const extensionHostUri = makeRepoURI({
            repoName: uri.repositoryName,
            revision: uri.revision,
            filePath: uri.path,
        })

        const definitions = wrapRemoteObservable(
            this.sourcegraphExtensionHostAPI.getHover({
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

                    const prefix = result?.aggregatedBadges?.reduce((prefix, badge) => {
                        if (badge.linkURL) {
                            return prefix + `[${badge.text}](${badge.linkURL})\n`
                        }
                        return prefix + `${badge.text}\n`
                    }, '')

                    return of<vscode.Hover>({
                        contents: [
                            new vscode.MarkdownString(prefix),
                            ...(result?.contents ?? []).map(content => new vscode.MarkdownString(content.value)),
                        ],
                    })
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
