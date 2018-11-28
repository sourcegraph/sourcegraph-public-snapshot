import assert from 'assert'
import { map } from 'rxjs/operators'
import { Window } from 'sourcegraph'
import { MessageType } from '../client/services/notifications'
import { assertToJSON } from '../extension/types/common.test'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Windows (integration)', () => {
    describe('app.activeWindow', () => {
        it('returns the active window', async () => {
            const { extensionHost } = await integrationTestContext()
            assertToJSON(extensionHost.app.activeWindow, {
                visibleViewComponents: [
                    {
                        type: 'CodeEditor' as 'CodeEditor',
                        document: { uri: 'file:///f', languageId: 'l', text: 't' },
                    },
                ],
            } as Window)
        })
    })

    describe('app.windows', () => {
        it('lists windows', async () => {
            const { extensionHost } = await integrationTestContext()
            assertToJSON(extensionHost.app.windows, [
                {
                    visibleViewComponents: [
                        {
                            type: 'CodeEditor' as 'CodeEditor',
                            document: { uri: 'file:///f', languageId: 'l', text: 't' },
                        },
                    ],
                },
            ] as Window[])
        })

        it('adds new text documents', async () => {
            const { model, extensionHost } = await integrationTestContext()

            model.next({
                ...model.value,
                visibleTextDocuments: [{ uri: 'file:///f2', languageId: 'l2', text: 't2' }],
            })
            await extensionHost.internal.sync()

            assertToJSON(extensionHost.app.windows, [
                {
                    visibleViewComponents: [
                        {
                            type: 'CodeEditor' as 'CodeEditor',
                            document: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                        },
                    ],
                },
            ] as Window[])
        })
    })

    describe('Window', () => {
        it('Window#showNotification', async () => {
            const { extensionHost, services } = await integrationTestContext()
            const values = collectSubscribableValues(services.notifications.showMessages)
            extensionHost.app.activeWindow!.showNotification('a') // tslint:disable-line deprecation
            await extensionHost.internal.sync()
            assert.deepStrictEqual(values, [{ message: 'a', type: MessageType.Info }] as typeof values)
        })

        it('Window#showMessage', async () => {
            const { extensionHost, services } = await integrationTestContext()
            services.notifications.showMessageRequests.subscribe(({ resolve }) => resolve(Promise.resolve(null)))
            const values = collectSubscribableValues(
                services.notifications.showMessageRequests.pipe(map(({ message, type }) => ({ message, type })))
            )
            assert.strictEqual(await extensionHost.app.activeWindow!.showMessage('a'), null)
            assert.deepStrictEqual(values, [{ message: 'a', type: MessageType.Info }] as typeof values)
        })

        it('Window#showInputBox', async () => {
            const { extensionHost, services } = await integrationTestContext()
            services.notifications.showInputs.subscribe(({ resolve }) => resolve(Promise.resolve('c')))
            const values = collectSubscribableValues(
                services.notifications.showInputs.pipe(map(({ message, defaultValue }) => ({ message, defaultValue })))
            )
            assert.strictEqual(await extensionHost.app.activeWindow!.showInputBox({ prompt: 'a', value: 'b' }), 'c')
            assert.deepStrictEqual(values, [{ message: 'a', defaultValue: 'b' }] as typeof values)
        })
    })
})
