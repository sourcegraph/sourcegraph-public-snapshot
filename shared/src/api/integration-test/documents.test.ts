import * as assert from 'assert'
import { TextDocument } from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Documents (integration)', () => {
    describe('workspace.textDocuments', () => {
        it('lists text documents', async () => {
            const { extensionHost } = await integrationTestContext()
            assert.deepStrictEqual(extensionHost.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
            ] as TextDocument[])
        })

        it('adds new text documents', async () => {
            const { model, extensionHost } = await integrationTestContext()
            model.next({
                ...model.value,
                visibleTextDocuments: [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }],
            })
            await extensionHost.internal.sync()
            assert.deepStrictEqual(extensionHost.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
                { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            ] as TextDocument[])
        })
    })

    describe('workspace.onDidOpenTextDocument', () => {
        it('fires when a text document is opened', async () => {
            const { model, extensionHost } = await integrationTestContext()

            const values = collectSubscribableValues(extensionHost.workspace.onDidOpenTextDocument)
            assert.deepStrictEqual(values, [] as TextDocument[])

            model.next({
                ...model.value,
                visibleTextDocuments: [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }],
            })
            await extensionHost.internal.sync()

            assert.deepStrictEqual(values, [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }] as TextDocument[])
        })
    })
})
