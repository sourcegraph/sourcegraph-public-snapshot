import { WorkspaceRoot } from 'sourcegraph'
import { URI } from '../extension/types/uri'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Workspace roots (integration)', () => {
    describe('workspace.roots', () => {
        test('lists roots', async () => {
            const { extensionAPI } = await integrationTestContext()
            expect(extensionAPI.workspace.roots).toEqual([{ uri: new URI('file:///') }] as WorkspaceRoot[])
        })

        test('adds new text documents', async () => {
            const { model, extensionAPI } = await integrationTestContext()

            model.next({
                ...model.value,
                roots: [{ uri: 'file:///a' }, { uri: 'file:///b' }],
            })
            await extensionAPI.internal.sync()

            expect(extensionAPI.workspace.roots).toEqual([
                { uri: new URI('file:///a') },
                { uri: new URI('file:///b') },
            ] as WorkspaceRoot[])
        })
    })

    describe('workspace.rootChanges', () => {
        test('fires when a root is added or removed', async () => {
            const { model, extensionAPI } = await integrationTestContext()

            const values = collectSubscribableValues(extensionAPI.workspace.rootChanges)
            expect(values).toEqual([] as void[])

            model.next({
                ...model.value,
                roots: [{ uri: 'file:///a' }],
            })
            await extensionAPI.internal.sync()

            expect(values).toEqual([void 0])
        })
    })
})
