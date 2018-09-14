import * as assert from 'assert'
import { TextDocument } from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Documents (integration)', () => {
    describe('workspace.textDocuments', () => {
        it('lists text documents', async () => {
            const { extensionHost, ready } = await integrationTestContext()

            await ready
            assert.deepStrictEqual(extensionHost.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
            ] as TextDocument[])
        })

        it('adds new text documents', async () => {
            const { clientController, extensionHost, getEnvironment, ready } = await integrationTestContext()

            const prevEnvironment = getEnvironment()
            clientController.setEnvironment({
                ...prevEnvironment,
                visibleTextDocuments: [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }],
            })

            await ready
            assert.deepStrictEqual(extensionHost.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
                { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            ] as TextDocument[])
        })
    })

    describe('workspace.onDidOpenTextDocument', () => {
        it('fires when a text document is opened', async () => {
            const { clientController, extensionHost, getEnvironment, ready } = await integrationTestContext()

            await ready
            const values = collectSubscribableValues(extensionHost.workspace.onDidOpenTextDocument)
            assert.deepStrictEqual(values, [] as TextDocument[])

            const prevEnvironment = getEnvironment()
            clientController.setEnvironment({
                ...prevEnvironment,
                visibleTextDocuments: [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }],
            })
            await extensionHost.internal.sync()

            assert.deepStrictEqual(values, [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }] as TextDocument[])
        })
    })
})
