import { TextDocument } from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Documents (integration)', () => {
    describe('workspace.textDocuments', () => {
        test('lists text documents', async () => {
            const { extensionHost } = await integrationTestContext()
            expect(extensionHost.workspace.textDocuments).toEqual([
                { uri: 'file:///f', languageId: 'l', text: 't' },
            ] as TextDocument[])
        })

        test('adds new text documents', async () => {
            const { model, extensionHost } = await integrationTestContext()
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            await extensionHost.internal.sync()
            expect(extensionHost.workspace.textDocuments).toEqual([
                { uri: 'file:///f', languageId: 'l', text: 't' },
                { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            ] as TextDocument[])
        })
    })

    describe('workspace.onDidOpenTextDocument', () => {
        test('fires when a text document is opened', async () => {
            const { model, extensionHost } = await integrationTestContext()

            const values = collectSubscribableValues(extensionHost.workspace.onDidOpenTextDocument)
            expect(values).toEqual([] as TextDocument[])

            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            await extensionHost.internal.sync()

            expect(values).toEqual([{ uri: 'file:///f2', languageId: 'l2', text: 't2' }] as TextDocument[])
        })
    })

    describe('workspace.activeTextDocument', () => {
        test('emits `null` on subscription if there is no active text document', async () => {
            const { extensionHost } = await integrationTestContext(undefined, {
                roots: [],
                visibleViewComponents: [],
            })
            const values = collectSubscribableValues(extensionHost.workspace.activeTextDocument)
            expect(values).toEqual([null])
        })

        test('emits the active text document on subscription if there is one', async () => {
            const { model, extensionHost } = await integrationTestContext(undefined, {
                roots: [],
                visibleViewComponents: [],
            })
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: { uri: 'file:///f1', languageId: 'l1', text: 't1' },
                        selections: [],
                        isActive: false,
                    },
                    {
                        type: 'textEditor',
                        item: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            await extensionHost.internal.sync()
            const values = collectSubscribableValues(extensionHost.workspace.activeTextDocument)
            expect(values.pop()).toEqual({ uri: 'file:///f2', languageId: 'l2', text: 't2' })
        })

        test('emits null when all viewComponents are closed', async () => {
            const { model, extensionHost } = await integrationTestContext(undefined, {
                roots: [],
                visibleViewComponents: [],
            })
            const values = collectSubscribableValues(extensionHost.workspace.activeTextDocument)
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: { uri: 'file:///f1', languageId: 'l1', text: 't1' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            await extensionHost.internal.sync()
            model.next({
                ...model.value,
                visibleViewComponents: [],
            })
            await extensionHost.internal.sync()
            expect(values).toEqual([null, { uri: 'file:///f1', languageId: 'l1', text: 't1' }, null])
        })
    })
})
