/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Sourcegraph. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
'use strict'

import * as vscode from 'vscode'
import { newClient } from './client'
import * as net from 'net'
import * as child_process from 'child_process'
import { LanguageClient, ServerOptions, StreamInfo } from 'vscode-languageclient/lib/main'
import * as lodash from 'lodash'
import { Executable } from 'vscode-languageclient/lib/client'
import * as request from 'request-promise-native'
import _ = require('lodash')
import * as JSONC from 'jsonc-parser'
import * as path from 'path'
// TODO Figure out how to import GraphQL types. Last time I tried this, I had to
// set `rootDir: '../..'` in tsconfig.json, which seemed to break the extension.
// import * as GQL from '../../../web/src/backend/graphqlschema'

// I'm not 100% sure why, but TypeScript doesn't seem to believe that this is a
// module, thus the awkward syntax.
import deepmerge = require('deepmerge')

// Causes .ts files to show up in stack traces
require('source-map-support').install()

import GitUrlParse = require('git-url-parse')

interface PlatformStdio {
    type: 'stdio'
    command: string
    arguments: string[]
}

interface PlatformTcp {
    type: 'tcp'
    address: string
}

interface PlatformDocker {
    type: 'docker'
    image: string
    container: string
    args: string[]
}

interface PlatformSourcegraph {
    type: 'sourcegraph'
    languageId: string
}

interface ExtensionManifest {
    activationEvents: string[]
    platform: PlatformStdio | PlatformTcp | PlatformDocker | PlatformSourcegraph
}

interface Extension {
    manifest: ExtensionManifest
    settings: any
}

const runGitCommand = (command: string): string =>
    child_process
        .execSync(command, { cwd: vscode.workspace.workspaceFolders![0].uri.fsPath })
        .toString()
        .trim()

const getRemoteUrl = (): vscode.Uri | undefined => {
    const result = runGitCommand('git remote --verbose')
    const regex = /^([^\s]+)\s+([^\s]+)\s/
    const rawRemotes = result
        .split('\n')
        .filter(b => !!b)
        .map(line => regex.exec(line))
        .filter((g: RegExpExecArray | null): g is RegExpExecArray => !!g)
        .map(groups => ({ name: groups[1], url: groups[2] }))

    const url = GitUrlParse(rawRemotes[0].url)

    if (!url) {
        return undefined
    }

    return vscode.Uri.parse('/').with({
        scheme: 'git',
        authority: url.resource,
        path: '/' + url.owner + '/' + url.name,
    })
}

const getGitCommit = (): string => runGitCommand('git rev-parse HEAD')

const systemSync = (cmd: string): { stdout: string; status?: number; error?: any } => {
    try {
        return {
            stdout: child_process
                .execSync(cmd, { cwd: vscode.workspace.rootPath })
                .toString()
                .trim(),
        }
    } catch (error) {
        return { stdout: '', error }
    }
}

process.on('unhandledRejection', (error: any) => {
    console.log('unhandledRejection', error)
})

const registerLSPExtensions = ({ client, context }: { client: LanguageClient; context: vscode.ExtensionContext }) => {
    const decorate = async () => {
        if (!lodash.get(client, 'initializeResult.capabilities.experimental.decorationsProvider', false)) {
            return
        }

        const editor = vscode.window.activeTextEditor

        if (!editor) {
            return
        }

        const path = editor.document.uri.fsPath
        const isGitDir = systemSync('git rev-parse --is-inside-work-tree').stdout === 'true'
        const isTrackedFile = [0, undefined].indexOf(systemSync('git ls-files --error-unmatch ' + path).status) !== -1

        if (isGitDir && isTrackedFile) {
            const decorations: any = await client.sendRequest('textDocument/decorations', {
                textDocument: {
                    uri: vscode.Uri.file(vscode.workspace.asRelativePath(path)).toString(),
                },
            })

            if (!decorations) {
                return
            }

            for (const decoration of decorations) {
                const decorationType = vscode.window.createTextEditorDecorationType(decoration)
                editor.setDecorations(decorationType, [decoration.range])
                context.subscriptions.push(decorationType)
            }
        }
    }

    // Decorate each text editor that the user switches to.
    vscode.window.onDidChangeActiveTextEditor(decorate)

    // Don't wait for the user to switch editors - immediately decorate the
    // active one.
    decorate()

    client.onRequest('workspace/exec', params => {
        // TODO handle spaces in filenames
        return {
            stdout: systemSync(params.command + ' ' + params.arguments.join(' ')).stdout,
            stderr: '', // TODO capture real stderr
            exitCode: 0, // TODO capture real exit code
        }
    })

    client.onRequest('textDocument/xcontent', async params => {
        let doc: vscode.TextDocument | undefined = undefined
        for (const folder of vscode.workspace.workspaceFolders || []) {
            try {
                doc = await vscode.workspace.openTextDocument(
                    folder.uri.fsPath + vscode.Uri.parse(params.textDocument.uri).fsPath
                )
                break
            } catch (e) {
                // The file doesn't exist in this workspace.
            }
        }

        if (doc) {
            return {
                uri: params.textDocument.uri,
                languageId: doc.languageId,
                version: doc.version,
                text: doc.getText(),
            }
        } else {
            throw new Error('File ' + params.textDocument.uri + ' not found in any workspace')
        }
    })

    client.onRequest('workspace/xfiles', async params => {
        return (await vscode.workspace.findFiles(params.rootPath ? params.rootPath + '/**/*' : '**/*')).map(uri => ({
            uri: vscode.Uri.file(
                path.relative(vscode.workspace.getWorkspaceFolder(uri)!.uri.fsPath, uri.fsPath)
            ).toString(),
        }))
    })
}

export async function activate({
    context,
    logger,
}: {
    context: vscode.ExtensionContext
    logger: vscode.OutputChannel
}): Promise<void> {
    if (!vscode.workspace.workspaceFolders || vscode.workspace.workspaceFolders.length == 0) {
        throw new Error('No workspace is currently open - not registering any Sourcegraph extensions.')
    }

    const stdio = (cmd: Executable): ServerOptions => ({ run: cmd, debug: cmd })

    const tcp = (options: { port: number; host?: string }) => (): Promise<StreamInfo> => {
        const socket = net.connect(options)
        return Promise.resolve(<StreamInfo>{
            writer: socket,
            reader: socket,
        })
    }

    const registerExtension = async (extension: Extension, name: string): Promise<void> => {
        logger.appendLine(name + ' Connecting')
        let client: LanguageClient

        const manifest = extension.manifest

        const clientOptions = {
            documentSelector: lodash.get(extension, 'manifest.activationEvents', ['*']),
        }

        switch (manifest.platform.type) {
            case 'stdio':
                client = new LanguageClient(
                    name,
                    stdio({ command: manifest.platform.command, args: manifest.platform.arguments }),
                    clientOptions
                )
                break
            case 'tcp':
                client = new LanguageClient(
                    name,
                    tcp({
                        host: manifest.platform.address.split(':')[0],
                        port: parseInt(manifest.platform.address.split(':')[1]),
                    }),
                    clientOptions
                )
                break
            case 'docker':
                const containerName = manifest.platform.container || manifest.platform.image.replace(/[:/]/g, '-')
                systemSync('docker rm -f ' + containerName)
                client = new LanguageClient(
                    name,
                    stdio({
                        command: 'docker',
                        args: [
                            'run',
                            '--name',
                            containerName,
                            '--rm',
                            '-i',
                            '-w',
                            vscode.workspace.workspaceFolders![0].uri.fsPath,
                            '--volume',
                            vscode.workspace.rootPath + ':' + vscode.workspace.rootPath,
                            ...(manifest.platform.args || []),
                            manifest.platform.image,
                        ],
                    }),
                    clientOptions
                )
                break
            case 'sourcegraph':
                const token = vscode.workspace.getConfiguration('sourcegraph').get<string>('token')
                const url = vscode.workspace.getConfiguration('sourcegraph').get<string>('URL')
                if (!url) {
                    return
                }
                const endpointAuthority = vscode.Uri.parse(url).authority
                const scheme = vscode.Uri.parse(url).scheme
                const rootWithoutCommit = await getRemoteUrl()
                const commitID = getGitCommit()
                if (!token) {
                    return
                }
                if (!rootWithoutCommit) {
                    return
                }
                client = newClient({
                    mode: name,
                    languageIds: [manifest.platform.languageId],
                    rootWithoutCommit,
                    commitID,
                    scheme,
                    endpointAuthority,
                    token,
                })
                break

            default:
                return
        }

        client.registerFeature({
            fillClientCapabilities: capabilities => {
                lodash.set(capabilities, 'experimental.exec', true)
                lodash.set(capabilities, 'experimental.decorations', true)
            },
            initialize: () => {},
            fillInitializeParams: params => {
                const remoteUrl = getRemoteUrl()
                if (remoteUrl) {
                    lodash.set(params, 'originalRootUri', remoteUrl.with({ query: getGitCommit() }).toString())
                }
                lodash.set(params, 'initializationOptions.settings.merged', lodash.get(extension, 'settings', {}))
            },
        })

        context.subscriptions.push(client.start())
        client.onReady().then(() => {
            registerLSPExtensions({ client, context })
            client.outputChannel.appendLine('Ready')
            client.outputChannel.appendLine('Capabilities: ' + JSON.stringify(client.initializeResult))
        })
    }

    const sourcegraphToken = vscode.workspace.getConfiguration('sourcegraph').get<string>('token')
    const sourcegraphURL = vscode.workspace.getConfiguration('sourcegraph').get<string>('URL')

    let sourcegraphExtensions: { [id: string]: Extension }
    if (sourcegraphURL) {
        try {
            sourcegraphExtensions = await fetchExtensionsFromSourcegraph(sourcegraphURL, sourcegraphToken)
        } catch (e) {
            vscode.window.showErrorMessage(
                'Unable to connect to Sourcegraph instance ' + sourcegraphURL + ': ' + e.toString()
            )
            sourcegraphExtensions = {}
        }
    } else {
        console.log('sourcegraph.url is not set, only running extensions configured in user settings')
        sourcegraphExtensions = {}
    }

    const localExtensions = _.pickBy(
        vscode.workspace.getConfiguration('sourcegraph').get<{ [name: string]: Extension }>('extensions'),
        override => !lodash.get(override, 'settings.disabled')
    )

    // TODO consider using a Map<string, Extension> instead of an object.
    const extensions: { [id: string]: Extension } = _.pickBy(
        deepmerge(sourcegraphExtensions, localExtensions, {
            arrayMerge: (_, source) => source,
        }),
        extension => lodash.get(extension, 'manifest.platform')
    ) as { [id: string]: Extension }

    if (_.keys(extensions).length > 0) {
        logger.appendLine('Starting ' + _.keys(extensions).length + ' extensions')
    } else {
        logger.appendLine(
            'No extensions are enabled. Visit ' +
                (sourcegraphURL || 'https://about.sourcegraph.com/docs/') +
                ' to enable extensions.'
        )
    }

    lodash.forEach(extensions, registerExtension)
}

/**
 * Simultaneously maps and filters an array.
 */
function mapMaybe<T, U>(ts: T[], f: (t: T) => U | undefined): U[] {
    return _.flatMap(ts, t => {
        const uMaybe = f(t)
        return uMaybe ? [uMaybe] : []
    })
}

async function fetchExtensionsFromSourcegraph(
    sourcegraphURL: string,
    token: string | undefined
): Promise<{ [id: string]: Extension }> {
    const response = await request({
        method: 'POST',
        url: sourcegraphURL + '/.api/graphql',
        headers: {
            ...(token
                ? {
                      Authorization: 'token ' + token,
                  }
                : {}),
        },
        json: true,
        body: {
            query: `
                    query {
                        currentUser {
                            configuredExtensions {
                                nodes {
                                    extension {
                                    extensionID
                                        manifest {
                                            raw
                                        }
                                    }
                                    capabilities
                                    contributions
                                    mergedSettings
                                }
                                url
                            }
                        }
                    }
                `,
            variables: {},
        },
    })

    const extensions: any[] = _.get(response, 'data.currentUser.configuredExtensions.nodes')

    return extensions
        ? _.fromPairs(
              mapMaybe(extensions, e => {
                  return e.extension
                      ? [
                            e.extension.extensionID,
                            {
                                manifest: JSONC.parse(_.get(e, 'extension.manifest.raw', '')),
                                settings: e.mergedSettings,
                            },
                        ]
                      : undefined
              })
          )
        : {}
}
