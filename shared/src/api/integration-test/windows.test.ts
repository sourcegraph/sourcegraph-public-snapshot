import { from } from 'rxjs'
import { distinctUntilChanged, map, switchMap, take, toArray } from 'rxjs/operators'
import { NotificationType, ViewComponent, Window } from 'sourcegraph'
import { assertToJSON } from '../extension/types/testHelpers'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

describe('Windows (integration)', () => {
    describe('app.activeWindow', () => {
        test('returns the active window', async () => {
            const { extensionAPI } = await integrationTestContext()
            await extensionAPI.internal.sync()
            await extensionAPI.internal.sync()
            const viewComponent: Pick<ViewComponent, 'type' | 'document'> = {
                type: 'CodeEditor' as const,
                document: { uri: 'file:///f', languageId: 'l', text: 't' },
            }
            assertToJSON(extensionAPI.app.activeWindow, {
                visibleViewComponents: [viewComponent],
                activeViewComponent: viewComponent,
            })
        })
    })

    describe('app.activeWindowChanges', () => {
        test('reflects changes to the active window', async () => {
            const {
                services: { editor: editorService, model: modelService },
                extensionAPI,
            } = await integrationTestContext(undefined, {
                roots: [],
                editors: [],
            })
            modelService.addModel({
                uri: 'u',
                languageId: 'l',
                text: 't',
            })
            editorService.addEditor({
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            })
            const values = await from(extensionAPI.app.activeWindowChanges)
                .pipe(
                    take(1),
                    toArray()
                )
                .toPromise()
            assertToJSON(values.map(w => !!w), [true])
        })
    })

    describe('app.windows', () => {
        test('lists windows', async () => {
            const { extensionAPI } = await integrationTestContext()
            await extensionAPI.internal.sync()
            await extensionAPI.internal.sync()
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
            const {
                services: { editor: editorService, model: modelService },
                extensionAPI,
            } = await integrationTestContext(undefined, { editors: [], roots: [] })

            modelService.addModel({ uri: 'file:///f2', languageId: 'l2', text: 't2' })
            editorService.addEditor({
                type: 'CodeEditor',
                resource: 'file:///f2',
                selections: [],
                isActive: true,
            })
            await from(extensionAPI.app.activeWindowChanges)
                .pipe(
                    switchMap(w => (w ? w.activeViewComponentChanges : [])),
                    take(3)
                )
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
            const {
                services: { editor: editorService, model: modelService },
                extensionAPI,
            } = await integrationTestContext()

            modelService.addModel({
                uri: 'file:///inactive',
                languageId: 'inactive',
                text: 'inactive',
            })
            editorService.addEditor({
                type: 'CodeEditor',
                resource: 'file:///inactive',
                selections: [],
                isActive: false,
            })
            await from(extensionAPI.app.activeWindowChanges)
                .pipe(
                    switchMap(w => (w ? w.activeViewComponentChanges : [])),
                    take(4)
                )
                .toPromise()

            assertToJSON(extensionAPI.app.windows[0].visibleViewComponents, [
                {
                    type: 'CodeEditor' as const,
                    document: { uri: 'file:///f', languageId: 'l', text: 't' },
                },
                {
                    type: 'CodeEditor' as const,
                    document: { uri: 'file:///inactive', languageId: 'inactive', text: 'inactive' },
                },
            ] as ViewComponent[])
        })

        describe('Window#activeViewComponent', () => {
            test('ignores inactive components', async () => {
                const {
                    services: { editor: editorService, model: modelService },
                    extensionAPI,
                } = await integrationTestContext()

                modelService.addModel({
                    uri: 'file:///inactive',
                    languageId: 'inactive',
                    text: 'inactive',
                })
                editorService.addEditor({
                    type: 'CodeEditor',
                    resource: 'file:///inactive',
                    selections: [],
                    isActive: false,
                })
                await extensionAPI.internal.sync()
                await extensionAPI.internal.sync()

                assertToJSON(extensionAPI.app.windows[0].activeViewComponent, {
                    type: 'CodeEditor' as const,
                    document: { uri: 'file:///f', languageId: 'l', text: 't' },
                })
            })
        })

        describe('Window#activeViewComponentChanges', () => {
            test('reflects changes to the active window', async () => {
                const {
                    services: { editor: editorService, model: modelService },
                    extensionAPI,
                } = await integrationTestContext(undefined, {
                    roots: [],
                    editors: [],
                })
                modelService.addModel({ uri: 'foo', languageId: 'l1', text: 't1' })
                modelService.addModel({ uri: 'bar', languageId: 'l2', text: 't2' })
                editorService.addEditor({
                    type: 'CodeEditor',
                    resource: 'foo',
                    selections: [],
                    isActive: true,
                })
                editorService.removeAllEditors()
                editorService.addEditor({
                    type: 'CodeEditor',
                    resource: 'bar',
                    selections: [],
                    isActive: true,
                })
                const values = await from(extensionAPI.app.windows[0].activeViewComponentChanges)
                    .pipe(
                        distinctUntilChanged(),
                        take(4),
                        toArray()
                    )
                    .toPromise()
                assertToJSON(values.map(c => (c ? c.document.uri : null)), [null, 'foo', null, 'bar'])
            })
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
