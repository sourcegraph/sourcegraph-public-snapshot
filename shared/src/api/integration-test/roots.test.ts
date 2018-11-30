import * as assert from 'assert'
import { WorkspaceRoot } from 'sourcegraph'
import { URI } from '../extension/types/uri'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Workspace roots (integration)', () => {
    describe('workspace.roots', () => {
        it('lists roots', async () => {
            const { extensionHost } = await integrationTestContext()
            assert.deepStrictEqual(extensionHost.workspace.roots, [{ uri: new URI('file:///') }] as WorkspaceRoot[])
        })

        it('adds new text documents', async () => {
            const { model, extensionHost } = await integrationTestContext()

            model.next({
                ...model.value,
                roots: [{ uri: 'file:///a' }, { uri: 'file:///b' }],
            })
            await extensionHost.internal.sync()

            assert.deepStrictEqual(extensionHost.workspace.roots, [
                { uri: new URI('file:///a') },
                { uri: new URI('file:///b') },
            ] as WorkspaceRoot[])
        })
    })

    describe('workspace.onDidChangeRoots', () => {
        it('fires when a root is added or removed', async () => {
            const { model, extensionHost } = await integrationTestContext()

            const values = collectSubscribableValues(extensionHost.workspace.onDidChangeRoots)
            assert.deepStrictEqual(values, [] as void[])

            model.next({
                ...model.value,
                roots: [{ uri: 'file:///a' }],
            })
            await extensionHost.internal.sync()

            assert.deepStrictEqual(values, [void 0])
        })
    })
})
