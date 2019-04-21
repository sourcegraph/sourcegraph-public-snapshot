import { from } from 'rxjs'
import { take } from 'rxjs/operators'
import { TextDocument } from 'sourcegraph'
import { assertToJSON } from '../extension/types/testHelpers'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

jest.setTimeout(400)

describe('Documents (integration)', () => {
    describe('workspace.textDocuments', () => {
        test('lists text documents', async () => {
            const { extensionAPI } = await integrationTestContext()
            assertToJSON(extensionAPI.workspace.textDocuments, [
                { uri: 'file:///f', languageId: 'l', text: 't' },
            ] as TextDocument[])
        })

        test.only('adds new text documents', async () => {
            const {
                services: { model: modelService },
                extensionAPI,
            } = await integrationTestContext(undefined, { roots: [], editors: [] })
            console.log('test: call addModel')
            // setTimeout(() =>
            modelService.addModel({ uri: 'file:///f2', languageId: 'l2', text: 't2' }) // , 50)
            console.log('test: watch openedTextDocuments')
            await from(extensionAPI.workspace.openedTextDocuments)
                .pipe(take(1))
                .toPromise()
            assertToJSON(extensionAPI.workspace.textDocuments, [
                // { uri: 'file:///f', languageId: 'l', text: 't' },
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
})
