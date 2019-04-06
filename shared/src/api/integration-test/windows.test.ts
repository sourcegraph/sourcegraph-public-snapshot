import { from } from 'rxjs'
import { map, take, toArray } from 'rxjs/operators'
import { NotificationType, ViewComponent, Window } from 'sourcegraph'
import { assertToJSON } from '../extension/types/testHelpers'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Windows (integration)', () => {
    describe('app.activeWindow', () => {
        test('returns the active window', async () => {
            const { extensionAPI } = await integrationTestContext()
            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as const,
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            }
            assertToJSON(extensionAPI.app.activeWindow, {
                visibleViewComponents: [viewComponent],
                activeViewComponent: viewComponent,
            } as Window)
        })
    })

    describe('app.activeWindowChanges', () => {
        test('reflects changes to the active window', async () => {
            const { extensionAPI, model } = await integrationTestContext(undefined, {
                roots: [],
                visibleViewComponents: [],
            })
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
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
            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
                        item: { uri: 'bar', languageId: 'l2', text: 't2' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            const values = await from(extensionAPI.app.activeWindowChanges)
                .pipe(
                    take(4),
                    toArray()
                )
                .toPromise()
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
            const { extensionAPI } = await integrationTestContext()
            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as const,
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            }
            assertToJSON(extensionAPI.app.windows, [
                {
                    visibleViewComponents: [viewComponent],
                    activeViewComponent: viewComponent,
                },
            ] as Window[])
        })

        test('adds new text documents', async () => {
            const { model, extensionAPI } = await integrationTestContext()

            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
                        item: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
            await from(extensionAPI.app.activeWindowChanges)
                .pipe(take(2))
                .toPromise()

            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as const,
                document: { uri: 'file:///f2', languageId: 'l2', text: 't2' },
            }
            assertToJSON(extensionAPI.app.windows, [
                {
                    visibleViewComponents: [viewComponent],
                    activeViewComponent: viewComponent,
                },
            ] as Window[])
        })
    })

    describe('Window', () => {
        test('Window#visibleViewComponents', async () => {
            const { model, extensionAPI } = await integrationTestContext()

            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
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
            await from(extensionAPI.app.activeWindowChanges)
                .pipe(take(2))
                .toPromise()

            assertToJSON(extensionAPI.app.windows[0].visibleViewComponents, [
                {
                    type: 'CodeEditor' as const,
                    document: { uri: 'file:///inactive', languageId: 'inactive', text: 'inactive' },
                },
                {
                    type: 'CodeEditor' as const,
                    document: { uri: 'file:///f', languageId: 'l', text: 't' },
                },
            ] as ViewComponent[])
        })

        test('Window#activeViewComponent', async () => {
            const { model, extensionAPI } = await integrationTestContext()

            model.next({
                ...model.value,
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
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
            await extensionAPI.internal.sync()

            assertToJSON(extensionAPI.app.windows[0].activeViewComponent, {
                type: 'CodeEditor' as const,
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            } as ViewComponent)
        })

        test('Window#showNotification', async () => {
            const { extensionAPI, services } = await integrationTestContext()
            const values = collectSubscribableValues(services.notifications.showMessages)
            extensionAPI.app.activeWindow!.showNotification('a', NotificationType.Info) // tslint:disable-line deprecation
            await extensionAPI.internal.sync()
            expect(values).toEqual([{ message: 'a', type: NotificationType.Info }] as typeof values)
        })

        test('Window#showMessage', async () => {
            const { extensionAPI, services } = await integrationTestContext()
            services.notifications.showMessageRequests.subscribe(({ resolve }) => resolve(Promise.resolve(null)))
            const values = collectSubscribableValues(
                services.notifications.showMessageRequests.pipe(map(({ message, type }) => ({ message, type })))
            )
            expect(await extensionAPI.app.activeWindow!.showMessage('a')).toBe(undefined)
            expect(values).toEqual([{ message: 'a', type: NotificationType.Info }] as typeof values)
        })

        test('Window#showInputBox', async () => {
            const { extensionAPI, services } = await integrationTestContext()
            services.notifications.showInputs.subscribe(({ resolve }) => resolve(Promise.resolve('c')))
            const values = collectSubscribableValues(
                services.notifications.showInputs.pipe(map(({ message, defaultValue }) => ({ message, defaultValue })))
            )
            expect(await extensionAPI.app.activeWindow!.showInputBox({ prompt: 'a', value: 'b' })).toBe('c')
            expect(values).toEqual([{ message: 'a', defaultValue: 'b' }] as typeof values)
        })
    })
})
