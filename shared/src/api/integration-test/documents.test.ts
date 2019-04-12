import { from } from 'rxjs'
import { take } from 'rxjs/operators'
import { TextDocument } from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Documents (integration)', () => {
    describe('workspace.textDocuments', () => {
        test('lists text documents', async () => {
            const { extensionAPI } = await integrationTestContext()
            expect(extensionAPI.workspace.textDocuments).toEqual([
                { uri: 'file:///f', languageId: 'l', text: 't' },
            ] as TextDocument[])
        })

        test('adds new text documents', async () => {
            const {
                services: { editor: editorService },
                extensionAPI,
            } = await integrationTestContext()
            editorService.editors.next([
                {
                    type: 'CodeEditor',
                    item: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                    selections: [],
                    isActive: true,
                },
            ])
            await from(extensionAPI.workspace.openedTextDocuments)
                .pipe(take(1))
                .toPromise()
            expect(extensionAPI.workspace.textDocuments).toEqual([
                { uri: 'file:///f', languageId: 'l', text: 't' },
                { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            ] as TextDocument[])
        })
    })

    describe('workspace.openedTextDocuments', () => {
        test('fires when a text document is opened', async () => {
            const {
                services: { editor: editorService },
                extensionAPI,
            } = await integrationTestContext()

            const values = collectSubscribableValues(extensionAPI.workspace.openedTextDocuments)
            expect(values).toEqual([] as TextDocument[])

            editorService.editors.next([
                {
                    type: 'CodeEditor',
                    item: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                    selections: [],
                    isActive: true,
                },
            ])
            await from(extensionAPI.workspace.openedTextDocuments)
                .pipe(take(1))
                .toPromise()

            expect(values).toEqual([{ uri: 'file:///f2', languageId: 'l2', text: 't2' }] as TextDocument[])
        })
    })
})
