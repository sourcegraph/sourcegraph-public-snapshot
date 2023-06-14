/* eslint-disable no-sync */
import './mock-vscode'

import fs from 'fs'
import path from 'path'

import * as vscode from 'vscode'
import { URI } from 'vscode-uri'

import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { CodyCompletionItemProvider } from '../../completions'
import { CompletionsDocumentProvider } from '../../completions/docprovider'
import { History } from '../../completions/history'
import { getFullConfig } from '../../configuration'
import { logger } from '../../log'
import { InMemorySecretStorage } from '../../services/SecretStorageProvider'

import { CURSOR, completionsDataset } from './completions-dataset'
import { ENVIRONMENT_CONFIG } from './environment-config'
import { webviewErrorMessager, findSubstringPosition } from './utils'
import { TextDocument } from './vscode-text-document'

async function initCompletionsProvider(): Promise<CodyCompletionItemProvider> {
    const secretStorage = new InMemorySecretStorage()
    await secretStorage.store('cody.access-token', ENVIRONMENT_CONFIG.SOURCEGRAPH_ACCESS_TOKEN)

    const initialConfig = await getFullConfig(secretStorage)
    console.log('Running `initCompletionsProvider` with config:', initialConfig)

    const completionsClient = new SourcegraphNodeCompletionsClient(initialConfig, logger)

    if (!initialConfig.experimentalSuggest) {
        throw new Error('`cody.experimental.suggestions` is not true!')
    }

    const docprovider = new CompletionsDocumentProvider()
    const history = new History()

    // TODO: convert this args to an object for readability
    const completionsProvider = new CodyCompletionItemProvider(
        webviewErrorMessager,
        completionsClient,
        docprovider,
        history,
        null as any,
        null as any,
        2048,
        4,
        200,
        0.6,
        0.1,
        true
    )

    return completionsProvider
}

/**
 * Converts the code sample to a format that can be used by the VSCode completions provider.
 */
function prepareTextDocument(code: string): { textDocument: TextDocument; position: vscode.Position } {
    const position = findSubstringPosition(code, CURSOR)

    if (!position) {
        throw new Error(`No caret position found! add ${CURSOR} to the code.`)
    }

    // Remove CURSOR marks from the code before processing it further.
    const completionReadyCode = code.replaceAll(CURSOR, '')
    const textDocument = new TextDocument(URI.parse('/whatever/'), completionReadyCode)

    return { textDocument, position }
}

interface CompletionResult {
    completions: string[]
    timestamp: string
    code: string
}

// TODO: use VSCode mocked APIs to provide context for the completions provider
// See client/cody/src/completions/context.ts:10:23
async function generateCompletionsForDataset(codeSamples: string[]): Promise<void> {
    const completionsProvider = await initCompletionsProvider()

    const results: CompletionResult[] = []
    const timestamp = new Date().getTime().toString()

    await Promise.all(
        codeSamples.map(async (code, index) => {
            const { textDocument, position } = prepareTextDocument(code)

            const completionItems = await completionsProvider.provideInlineCompletionItems(textDocument, position, {
                triggerKind: 1,
                selectedCompletionInfo: undefined,
            })

            const completions = completionItems.map(item =>
                typeof item.insertText === 'string' ? item.insertText : ''
            )

            console.log(`#${index}`, completions)

            // Write completionItems, timestamp and code to a JSON file
            const data = {
                completions,
                timestamp,
                code,
            }
            results.push(data)
        })
    )

    // TODO: prettfy path management
    // Save results to a JSON file in the completions-review-tool/data folder to be used by the review tool:
    // pnpm --filter @sourcegraph/completions-review-tool run dev
    fs.mkdirSync(ENVIRONMENT_CONFIG.OUTPUT_PATH, { recursive: true })
    const filename = path.join(ENVIRONMENT_CONFIG.OUTPUT_PATH, `completions-${timestamp}.json`)
    fs.writeFileSync(filename, JSON.stringify(results, null, 2))
    console.log(
        '\n✅ Completions saved to:',
        filename.slice(filename.indexOf('/sourcegraph/') + '/sourcegraph/'.length)
    )
}

generateCompletionsForDataset(completionsDataset).catch(console.error)
