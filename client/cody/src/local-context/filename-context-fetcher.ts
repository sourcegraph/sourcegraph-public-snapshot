import { execFile } from 'child_process'
import * as path from 'path'

import { uniq } from 'lodash'
import * as vscode from 'vscode'

import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { ContextResult } from '@sourcegraph/cody-shared/src/local-context'

import { debug } from '../log'

/**
 * A local context fetcher that uses a LLM to generate filename fragments, which are then used to
 * find files that are relevant based on their path or name.
 */
export class FilenameContextFetcher {
    constructor(private rgPath: string, private editor: Editor, private chatClient: ChatClient) {}

    /**
     * Returns pieces of context relevant for the given query. Uses a filename search approach
     * @param query user query
     * @param numResults the number of context results to return
     * @returns a list of context results, sorted in *reverse* order (that is,
     * the most important result appears at the bottom)
     */
    public async getContext(query: string, numResults: number): Promise<ContextResult[]> {
        const time0 = performance.now()

        const rootPath = this.editor.getWorkspaceRootPath()
        if (!rootPath) {
            return []
        }
        const time1 = performance.now()
        const filenameFragments = await this.queryToFileFragments(query)
        const time2 = performance.now()
        const unsortedMatchingFiles = await this.getFilenames(rootPath, filenameFragments, 3)
        const time3 = performance.now()

        const specialFragments = ['readme']
        const allBoostedFiles = []
        let remainingFiles = unsortedMatchingFiles
        let nextRemainingFiles = []
        for (const specialFragment of specialFragments) {
            const boostedFiles = []
            for (const fileName of remainingFiles) {
                const fileNameLower = fileName.toLocaleLowerCase()
                if (fileNameLower.includes(specialFragment)) {
                    boostedFiles.push(fileName)
                } else {
                    nextRemainingFiles.push(fileName)
                }
            }
            remainingFiles = nextRemainingFiles
            nextRemainingFiles = []
            allBoostedFiles.push(...boostedFiles.sort((a, b) => a.length - b.length))
        }

        const sortedMatchingFiles = allBoostedFiles.concat(remainingFiles).slice(0, numResults)

        const results = await Promise.all(
            sortedMatchingFiles
                .map(async fileName => {
                    const uri = vscode.Uri.file(path.join(rootPath, fileName))
                    const content = (await vscode.workspace.openTextDocument(uri)).getText()
                    return {
                        fileName,
                        content,
                    }
                })
                .reverse()
        )

        const time4 = performance.now()
        debug(
            'FilenameContextFetcher:getContext',
            JSON.stringify({
                duration: time4 - time0,
                queryToFileFragments: { duration: time2 - time1, fragments: filenameFragments },
                getFilenames: { duration: time3 - time2 },
            }),
            { verbose: { matchingFiles: unsortedMatchingFiles, results: results.map(r => r.fileName) } }
        )

        return results
    }

    private async queryToFileFragments(query: string): Promise<string[]> {
        const filenameFragments = await new Promise<string[]>((resolve, reject) => {
            let responseText = ''
            this.chatClient.chat(
                [
                    {
                        speaker: 'human',
                        text: `Write 3 filename fragments that would be contained by files in a git repository that are relevant to answering the following user query: <query>${query}</query> Your response should be only a space-delimited list of filename fragments and nothing else.`,
                    },
                ],
                {
                    onChange: (text: string) => {
                        responseText = text
                    },
                    onComplete: () => {
                        resolve(responseText.split(/\s+/).filter(e => e.length > 0))
                    },
                    onError: (message: string, statusCode?: number) => {
                        reject(new Error(message))
                    },
                },
                {
                    temperature: 0,
                    fast: true,
                }
            )
        })
        const uniqueFragments = uniq(filenameFragments.map(e => e.toLocaleLowerCase()))
        return uniqueFragments
    }

    private async getFilenames(rootPath: string, filenameFragments: string[], maxDepth: number): Promise<string[]> {
        const searchPattern = '{' + filenameFragments.map(fragment => `**${fragment}**`).join(',') + '}'
        const rgArgs = [
            '--files',
            '--iglob',
            searchPattern,
            '--crlf',
            '--fixed-strings',
            '--no-config',
            '--no-ignore-global',
            `--max-depth=${maxDepth}`,
        ]
        const results = await new Promise<string>((resolve, reject) => {
            execFile(
                this.rgPath,
                rgArgs,
                {
                    cwd: rootPath,
                    maxBuffer: 1024 * 1024 * 1024,
                },
                (error, stdout, stderr) => {
                    if (error?.code === 2) {
                        reject(new Error(`${error.message}: ${stderr}`))
                    } else {
                        resolve(stdout)
                    }
                }
            )
        })
        return results
            .split('\n')
            .map(r => r.trim())
            .filter(r => r.length > 0)
            .sort((a, b) => a.length - b.length)
    }
}
