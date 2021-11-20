import * as vscode from 'vscode'

import { Range } from '@sourcegraph/extension-api-types'

import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'

// TODO interact with Sourcegraph extension host
export class SourcegraphDefinitionProvider implements vscode.DefinitionProvider {
    constructor(private readonly fs: SourcegraphFileSystemProvider) {}
    public async provideDefinition(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): Promise<vscode.Definition | undefined> {
        const uri = this.fs.sourcegraphUri(document.uri)
        // const blob = await this.fs.fetchBlob(uri)

        console.log({ uri })

        // Array of (SG) Locations, map over them and parse each URI into LocationNode, pass that and Range
        // to fs.toVSCodeLocation.

        token.onCancellationRequested(() => {
            // TODO unsubscribe
        })

        return Promise.resolve(undefined)
    }
}

// TODO this will probably change during extension host implementation
export interface LocationNode {
    resource: {
        path: string
        repository: {
            name: string
        }
        commit: {
            oid: string
        }
    }
    range: Range
}
