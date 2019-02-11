import { map } from 'rxjs/operators'
import { ViewComponent, Window } from 'sourcegraph'
import { MessageType } from '../client/services/notifications'
import { assertToJSON } from '../extension/types/testHelpers'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Windows (integration)', () => {
    describe('app.activeWindow', () => {
        test('returns the active window', async () => {
            const { extensionHost } = await integrationTestContext()
            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as 'CodeEditor',
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            }
            assertToJSON(extensionHost.app.activeWindow, {
                visibleViewComponents: [viewComponent],
                activeViewComponent: viewComponent,
            } as Window)
        })
    })

    describe('app.activeWindowChanged', () => {
        test('reflects changes to the active window', async () => {
            const { extensionHost, model } = await integrationTestContext(undefined, {
                roots: [],
                visibleViewComponents: [],
            })
            await extensionHost.internal.sync()
            const values = collectSubscribableValues(extensionHost.app.activeWindowChanges)
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: { uri: 'foo', languageId: 'l1', text: 't1' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            model.next({
                ...model.value,
                visibleViewComponents: [],
            })
            await extensionHost.internal.sync()
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: { uri: 'bar', languageId: 'l2', text: 't2' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            await extensionHost.internal.sync()
            assertToJSON(values.map(w => w && w.activeViewComponent && w.activeViewComponent.document.uri), [
                null,
                'foo',
                null,
                'bar',
            ])
        })
    })

    describe('app.windows', () => {
        test('lists windows', async () => {
            const { extensionHost } = await integrationTestContext()
            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as 'CodeEditor',
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            }
            assertToJSON(extensionHost.app.windows, [
                {
                    visibleViewComponents: [viewComponent],
                    activeViewComponent: viewComponent,
                },
            ] as Window[])
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

            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as 'CodeEditor',
                document: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            }
            assertToJSON(extensionHost.app.windows, [
                {
                    visibleViewComponents: [viewComponent],
                    activeViewComponent: viewComponent,
                },
            ] as Window[])
        })
    })

    describe('Window', () => {
        test('Window#visibleViewComponent', async () => {
            const { model, extensionHost } = await integrationTestContext()

            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: {
                            uri: 'file:///inactive',
                            languageId: 'inactive',
                            text: 'inactive',
                        },
                        selections: [],
                        isActive: false,
                    },
                    ...(model.value.visibleViewComponents || []),
                ],
            })
            await extensionHost.internal.sync()

            assertToJSON(extensionHost.app.windows[0].visibleViewComponents, [
                {
                    type: 'CodeEditor' as 'CodeEditor',
                    document: { uri: 'file:///inactive', languageId: 'inactive', text: 'inactive' },
                },
                {
                    type: 'CodeEditor' as 'CodeEditor',
                    document: { uri: 'file:///f', languageId: 'l', text: 't' },
                },
            ] as ViewComponent[])
        })

        test('Window#activeViewComponent', async () => {
            const { model, extensionHost } = await integrationTestContext()

            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: {
                            uri: 'file:///inactive',
                            languageId: 'inactive',
                            text: 'inactive',
                        },
                        selections: [],
                        isActive: false,
                    },
                    ...(model.value.visibleViewComponents || []),
                ],
            })
            await extensionHost.internal.sync()

            assertToJSON(extensionHost.app.windows[0].activeViewComponent, {
                type: 'CodeEditor' as 'CodeEditor',
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            } as ViewComponent)
        })

        test('Window#showNotification', async () => {
            const { extensionHost, services } = await integrationTestContext()
            const values = collectSubscribableValues(services.notifications.showMessages)
            extensionHost.app.activeWindow!.showNotification('a') // tslint:disable-line deprecation
            await extensionHost.internal.sync()
            expect(values).toEqual([{ message: 'a', type: MessageType.Info }] as typeof values)
        })

        test('Window#showMessage', async () => {
            const { extensionHost, services } = await integrationTestContext()
            services.notifications.showMessageRequests.subscribe(({ resolve }) => resolve(Promise.resolve(null)))
            const values = collectSubscribableValues(
                services.notifications.showMessageRequests.pipe(map(({ message, type }) => ({ message, type })))
            )
            expect(await extensionHost.app.activeWindow!.showMessage('a')).toBe(null)
            expect(values).toEqual([{ message: 'a', type: MessageType.Info }] as typeof values)
        })

        test('Window#showInputBox', async () => {
            const { extensionHost, services } = await integrationTestContext()
            services.notifications.showInputs.subscribe(({ resolve }) => resolve(Promise.resolve('c')))
            const values = collectSubscribableValues(
                services.notifications.showInputs.pipe(map(({ message, defaultValue }) => ({ message, defaultValue })))
            )
            expect(await extensionHost.app.activeWindow!.showInputBox({ prompt: 'a', value: 'b' })).toBe('c')
            expect(values).toEqual([{ message: 'a', defaultValue: 'b' }] as typeof values)
        })
    })
})
