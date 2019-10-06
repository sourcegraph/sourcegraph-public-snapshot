import { from } from 'rxjs'
import { take } from 'rxjs/operators'
import { TextDocument } from 'sourcegraph'
import { assertToJSON, collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Documents (integration)', () => {
    describe('workspace.textDocuments', () => {
        test('lists text documents', async () => {
            const { extensionAPI } = await integrationTestContext()
            assertToJSON(extensionAPI.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
            ] as TextDocument[])
        })

        test('adds new text documents', async () => {
            const {
                services: { model: modelService },
                extensionAPI,
            } = await integrationTestContext()
            modelService.addModel({ uri: 'file:///f2', languageId: 'l2', text: 't2' })
            await from(extensionAPI.workspace.openedTextDocuments)
                .pipe(take(1))
                .toPromise()
            assertToJSON(extensionAPI.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
                { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            ] as TextDocument[])
        })
    })

    describe('workspace.openedTextDocuments', () => {
        test('fires when a text document is opened', async () => {
            const {
                services: { model: modelService },
                extensionAPI,
            } = await integrationTestContext()

            const values = collectSubscribableValues(extensionAPI.workspace.openedTextDocuments)
            expect(values).toEqual([] as TextDocument[])

            modelService.addModel({ uri: 'file:///f2', languageId: 'l2', text: 't2' })
            await from(extensionAPI.workspace.openedTextDocuments)
                .pipe(take(1))
                .toPromise()

            assertToJSON(values, [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }] as TextDocument[])
        })
    })

    describe('workspace.openTextDocument', () => {
        test('opens a document that was not already open', async () => {
            const {
                services: { model: modelService, fileSystem },
                extensionAPI,
            } = await integrationTestContext(undefined, { editors: [], roots: [] })
            fileSystem.setProvider(async () => 't')
            const values = collectSubscribableValues(extensionAPI.workspace.openedTextDocuments)
            expect(modelService.hasModel('file:///f')).toBeFalsy()
            const doc = await extensionAPI.workspace.openTextDocument(new URL('file:///f'))
            expect(doc).toMatchObject<Pick<TextDocument, 'uri' | 'text'>>({ uri: 'file:///f', text: 't' })
            expect(modelService.hasModel('file:///f')).toBeTruthy()
            expect(values).toEqual([doc])
        })

        test('returns a document that was already open', async () => {
            const {
                services: { model: modelService, fileSystem },
                extensionAPI,
            } = await integrationTestContext()
            fileSystem.setProvider(async () => 't')
            const values = collectSubscribableValues(extensionAPI.workspace.openedTextDocuments)
            expect(modelService.hasModel('file:///f')).toBeTruthy()
            const doc = await extensionAPI.workspace.openTextDocument(new URL('file:///f'))
            expect(doc).toMatchObject<Pick<TextDocument, 'uri' | 'text'>>({ uri: 'file:///f', text: 't' })
            expect(modelService.hasModel('file:///f')).toBeTruthy()
            expect(values).toEqual([])
        })
    })
})
