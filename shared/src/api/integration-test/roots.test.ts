import * as assert from 'assert'
import { WorkspaceRoot } from 'sourcegraph'
import { URI } from '../extension/types/uri'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Workspace roots (integration)', () => {
    describe('workspace.roots', () => {
        it('lists roots', async () => {
            const { extensionHost, ready } = await integrationTestContext()

            await ready
            assert.deepStrictEqual(extensionHost.workspace.roots, [{ uri: new URI('file:///') }] as WorkspaceRoot[])
        })

        it('adds new text documents', async () => {
            const { clientController, extensionHost, getEnvironment, ready } = await integrationTestContext()

            const prevEnvironment = getEnvironment()
            clientController.setEnvironment({
                ...prevEnvironment,
                roots: [{ uri: 'file:///a' }, { uri: 'file:///b' }],
            })

            await ready
            assert.deepStrictEqual(extensionHost.workspace.roots, [
                { uri: new URI('file:///a') },
                { uri: new URI('file:///b') },
            ] as WorkspaceRoot[])
        })
    })

    describe('workspace.onDidChangeRoots', () => {
        it('fires when a root is added or removed', async () => {
            const { clientController, extensionHost, getEnvironment, ready } = await integrationTestContext()

            await ready
            const values = collectSubscribableValues(extensionHost.workspace.onDidChangeRoots)
            assert.deepStrictEqual(values, [] as void[])

            const prevEnvironment = getEnvironment()
            clientController.setEnvironment({
                ...prevEnvironment,
                roots: [{ uri: 'file:///a' }],
            })
            await extensionHost.internal.sync()

            assert.deepStrictEqual(values, [void 0])
        })
    })
})
