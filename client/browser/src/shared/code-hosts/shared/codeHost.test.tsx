import { nextTick } from 'process'
import { promisify } from 'util'

import { RenderResult } from '@testing-library/react'
import { Remote } from 'comlink'
import { uniqueId, noop, isEmpty, pick } from 'lodash'
import { BehaviorSubject, NEVER, of, Subject, Subscription } from 'rxjs'
import { filter, take, first } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import * as sinon from 'sinon'
import * as sourcegraph from 'sourcegraph'

import { DiffPart } from '@sourcegraph/codeintellify'
import { allOf, check, isTaggedUnionMember, resetAllMemoizationCaches, subtypeOf } from '@sourcegraph/common'
import { Range } from '@sourcegraph/extension-api-classes'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { SuccessGraphQLResult } from '@sourcegraph/http-client'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ExtensionCodeEditor } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { integrationTestContext } from '@sourcegraph/shared/src/api/integration-test/testHelpers'
import { Controller } from '@sourcegraph/shared/src/extensions/controller'
import { IQuery } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/testing/MockIntersectionObserver'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { DEFAULT_SOURCEGRAPH_URL } from '../../util/context'
import { MutationRecordLike } from '../../util/dom'

import {
    CodeIntelligenceProps,
    createGlobalDebugMount,
    getExistingOrCreateOverlayMount,
    handleCodeHost,
    observeHoverOverlayMountLocation,
    HandleCodeHostOptions,
    DiffOrBlobInfo,
} from './codeHost'
import { toCodeViewResolver } from './codeViews'
import { DEFAULT_GRAPHQL_RESPONSES, mockRequestGraphQL } from './testHelpers'

const RENDER = sinon.spy()

const notificationClassNames = {
    [NotificationType.Log]: 'log',
    [NotificationType.Success]: 'success',
    [NotificationType.Info]: 'info',
    [NotificationType.Warning]: 'warning',
    [NotificationType.Error]: 'error',
}

const elementRenderedAtMount = (mount: Element): RenderResult | undefined => {
    const call = RENDER.args.find(call => call[1] === mount)
    return call?.[0]
}

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

const createTestElement = (): HTMLElement => {
    const element = document.createElement('div')
    element.className = `test test-${uniqueId()}`
    document.body.append(element)
    return element
}

jest.mock('uuid', () => ({
    v4: () => 'uuid',
}))

const createMockController = (extensionHostAPI: Remote<FlatExtensionHostAPI>): Controller => ({
    executeCommand: () => Promise.resolve(),
    registerCommand: () => new Subscription(),
    commandErrors: NEVER,
    unsubscribe: noop,
    extHostAPI: Promise.resolve(extensionHostAPI),
})

const createMockPlatformContext = (
    partialMocks?: Partial<CodeIntelligenceProps['platformContext']>
): CodeIntelligenceProps['platformContext'] => ({
    urlToFile: toPrettyBlobURL,
    requestGraphQL: mockRequestGraphQL(),
    sideloadedExtensionURL: new Subject<string | null>(),
    settings: NEVER,
    refreshSettings: () => Promise.resolve(),
    sourcegraphURL: '',
    ...partialMocks,
})

const commonArguments = () =>
    subtypeOf<Partial<HandleCodeHostOptions>>()({
        mutations: of([{ addedNodes: [document.body], removedNodes: [] }]),
        showGlobalDebug: false,
        platformContext: createMockPlatformContext(),
        sourcegraphURL: DEFAULT_SOURCEGRAPH_URL,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        render: RENDER,
        userSignedIn: true,
        minimalUI: false,
        background: {
            notifyRepoSyncError: () => Promise.resolve(),
            openOptionsPage: () => Promise.resolve(),
        },
    })

function getEditors(
    extensionAPI: typeof sourcegraph
): Pick<ExtensionCodeEditor, 'type' | 'viewerId' | 'isActive' | 'resource' | 'selections'>[] {
    return [...extensionAPI.app.activeWindow!.visibleViewComponents]
        .filter((viewer): viewer is ExtensionCodeEditor => viewer.type === 'CodeEditor')
        .map(editor => pick(editor, 'viewerId', 'isActive', 'resource', 'selections', 'type'))
}

const tick = promisify(nextTick)

describe('codeHost', () => {
    // Mock the global IntersectionObserver constructor with an implementation that
    // will immediately signal all observed elements as intersecting.
    beforeAll(() => {
        window.IntersectionObserver = MockIntersectionObserver
    })

    beforeEach(() => {
        document.body.innerHTML = ''
    })

    describe('createOverlayMount()', () => {
        it('should create the overlay mount', () => {
            getExistingOrCreateOverlayMount('some-code-host', document.body)
            const mount = document.body.querySelector('.hover-overlay-mount')
            expect(mount).toBeDefined()
            expect(mount!.className).toBe('hover-overlay-mount hover-overlay-mount__some-code-host')
        })
    })

    describe('createGlobalDebugMount()', () => {
        it('should create the debug menu mount', () => {
            createGlobalDebugMount()
            const mount = document.body.querySelector('[data-global-debug]')
            expect(mount).toBeDefined()
        })
    })

    describe('handleCodeHost()', () => {
        let subscriptions = new Subscription()

        afterEach(() => {
            RENDER.resetHistory()
            resetAllMemoizationCaches()
            subscriptions.unsubscribe()
            subscriptions = new Subscription()
        })

        test('renders the hover overlay mount', async () => {
            const { extensionHostAPI } = await integrationTestContext()
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        codeViewResolvers: [],
                        notificationClassNames,
                    },
                    extensionsController: createMockController(extensionHostAPI),
                })
            )
            const overlayMount = document.body.querySelector('.hover-overlay-mount')
            expect(overlayMount).toBeDefined()
            expect(overlayMount!.className).toBe('hover-overlay-mount hover-overlay-mount__github')
            const renderedOverlay = elementRenderedAtMount(overlayMount!)
            expect(renderedOverlay).not.toBeUndefined()
        })

        test('renders the command palette if codeHost.getCommandPaletteMount is defined', async () => {
            const { extensionHostAPI } = await integrationTestContext()
            const commandPaletteMount = createTestElement()
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        getCommandPaletteMount: () => commandPaletteMount,
                        codeViewResolvers: [],
                        notificationClassNames,
                    },
                    extensionsController: createMockController(extensionHostAPI),
                })
            )
            const renderedCommandPalette = elementRenderedAtMount(commandPaletteMount)
            expect(renderedCommandPalette).not.toBeUndefined()
        })

        test('creates a data-global-debug element and renders the debug menu if showGlobalDebug is true', async () => {
            const { extensionHostAPI } = await integrationTestContext()
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        codeViewResolvers: [],
                        notificationClassNames,
                    },
                    extensionsController: createMockController(extensionHostAPI),
                    showGlobalDebug: true,
                })
            )
            const globalDebugMount = document.body.querySelector('[data-global-debug]')
            expect(globalDebugMount).toBeDefined()
            const renderedDebugElement = elementRenderedAtMount(globalDebugMount!)
            expect(renderedDebugElement).toBeDefined()
        })

        test('detects code views based on selectors', async () => {
            const { extensionHostAPI, extensionAPI } = await integrationTestContext(undefined, {
                roots: [],
                viewers: [],
            })
            const codeView = createTestElement()
            codeView.id = 'code'
            const toolbarMount = document.createElement('div')
            codeView.append(toolbarMount)
            const blobInfo: DiffOrBlobInfo = {
                blob: {
                    rawRepoName: 'foo',
                    filePath: '/bar.ts',
                    commitID: '1',
                },
            }
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        notificationClassNames,
                        codeViewResolvers: [
                            toCodeViewResolver('#code', {
                                dom: {
                                    getCodeElementFromTarget: sinon.spy(),
                                    getCodeElementFromLineNumber: sinon.spy(),
                                    getLineElementFromLineNumber: sinon.spy(),
                                    getLineNumberFromCodeElement: sinon.spy(),
                                },
                                resolveFileInfo: codeView => of(blobInfo),
                                getToolbarMount: () => toolbarMount,
                            }),
                        ],
                    },
                    extensionsController: createMockController(extensionHostAPI),
                    showGlobalDebug: true,
                    platformContext: createMockPlatformContext({
                        // Simulate an instance with repositoryPathPattern
                        requestGraphQL: mockRequestGraphQL({
                            ...DEFAULT_GRAPHQL_RESPONSES,
                            ResolveRepo: variables =>
                                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                                of({
                                    data: {
                                        repository: {
                                            name: `github/${variables.rawRepoName as string}`,
                                        },
                                    },
                                    errors: undefined,
                                } as SuccessGraphQLResult<IQuery>),
                        }),
                    }),
                })
            )
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()

            expect(getEditors(extensionAPI)).toEqual([
                {
                    viewerId: 'viewer#0',
                    isActive: true,
                    // The repo name exposed to extensions is affected by repositoryPathPattern
                    resource: 'git://github/foo?1#/bar.ts',
                    selections: [],
                    type: 'CodeEditor',
                },
            ])

            await tick()
            expect(codeView).toHaveClass('sg-mounted')
            const toolbar = elementRenderedAtMount(toolbarMount)
            expect(toolbar).not.toBeUndefined()
        })

        describe('Decorations', () => {
            it('decorates a code view', async () => {
                const { extensionAPI, extensionHostAPI } = await integrationTestContext(undefined, {
                    roots: [],
                    viewers: [],
                })
                const codeView = createTestElement()
                codeView.id = 'code'
                const blobInfo: DiffOrBlobInfo = {
                    blob: {
                        rawRepoName: 'foo',
                        filePath: '/bar.ts',
                        commitID: '1',
                    },
                }
                // For this test, we pretend bar.ts only has one line of code
                const line = document.createElement('div')
                codeView.append(line)
                subscriptions.add(
                    await handleCodeHost({
                        ...commonArguments(),
                        codeHost: {
                            type: 'github',
                            name: 'GitHub',
                            check: () => true,
                            notificationClassNames,
                            codeViewResolvers: [
                                toCodeViewResolver('#code', {
                                    dom: {
                                        getCodeElementFromTarget: () => line,
                                        getCodeElementFromLineNumber: () => line,
                                        getLineElementFromLineNumber: () => line,
                                        getLineNumberFromCodeElement: () => 1,
                                    },
                                    resolveFileInfo: codeView => of(blobInfo),
                                }),
                            ],
                        },
                        extensionsController: createMockController(extensionHostAPI),
                        showGlobalDebug: true,
                    })
                )
                // const activeEditor = await from(extensionAPI.app.activeWindowChanges)
                //     .pipe(
                //         filter(isDefined),
                //         switchMap(window => window.activeViewComponentChanges),
                //         filter(isDefined),
                //         take(1)
                //     )
                //     .toPromise()

                await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()
                const editor = extensionAPI.app.activeWindow!.visibleViewComponents.find(
                    (viewer): viewer is ExtensionCodeEditor => viewer.type === 'CodeEditor'
                )

                if (!editor) {
                    throw new Error('Expected editor to be defined')
                }

                const decorationType = extensionAPI.app.createDecorationType()
                const decorated = (editor: ExtensionCodeEditor): Promise<TextDocumentDecoration[] | null> =>
                    wrapRemoteObservable(extensionHostAPI.getTextDecorations({ viewerId: editor.viewerId }))
                        .pipe(first(decorations => !isEmpty(decorations)))
                        .toPromise()

                // Set decorations and verify that a decoration attachment has been added
                editor.setDecorations(decorationType, [
                    {
                        range: new Range(0, 0, 0, 0),
                        after: {
                            contentText: 'test decoration',
                        },
                    },
                ])
                await decorated(editor)
                await tick()
                expect(line.querySelectorAll('[data-line-decoration-attachment]')).toHaveLength(1)
                expect(line.querySelector('[data-line-decoration-attachment]')!).toHaveTextContent('test decoration')

                // Decorate the code view again, and verify that previous decorations
                // are cleaned up and replaced by the new decorations.
                editor.setDecorations(decorationType, [
                    {
                        range: new Range(0, 0, 0, 0),
                        after: {
                            contentText: 'test decoration 2',
                        },
                    },
                ])
                await wrapRemoteObservable(extensionHostAPI.getTextDecorations({ viewerId: editor.viewerId }))
                    .pipe(
                        filter(
                            decorations =>
                                !!decorations &&
                                !!decorations[0].after &&
                                decorations[0].after.contentText === 'test decoration 2'
                        ),
                        take(1)
                    )
                    .toPromise()
                expect(line.querySelectorAll('[data-line-decoration-attachment]').length).toBe(1)
                expect(line.querySelector('[data-line-decoration-attachment]')!).toHaveTextContent('test decoration 2')
            })

            it('decorates a diff code view', async () => {
                const { extensionAPI, extensionHostAPI } = await integrationTestContext(undefined, {
                    roots: [],
                    viewers: [],
                })
                const codeView = createTestElement()
                codeView.id = 'code'
                const diffInfo: DiffOrBlobInfo = {
                    base: {
                        rawRepoName: 'foo',
                        filePath: '/bar.ts',
                        commitID: '1',
                    },
                    head: {
                        rawRepoName: 'foo',
                        filePath: '/bar.ts',
                        commitID: '2',
                    },
                }
                codeView.innerHTML =
                    '<div line="1" part="head"><span class="code-element"></span></div>\n' +
                    '<div line="2" part="base"><span class="code-element"></span></div>\n' +
                    '<div line="2" part="head"><span class="code-element"></span></div>\n' +
                    '<div line="4" part="head"><span class="code-element"></span></div>\n' +
                    '<div line="5" part="base"><span class="code-element"></span></div>\n'
                const dom = {
                    getCodeElementFromTarget: (target: HTMLElement) => target.closest('.code-element') as HTMLElement,
                    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number, part?: DiffPart) =>
                        codeView.querySelector<HTMLElement>(`[line="${line}"][part="${String(part)}"] > .code-element`),
                    getLineElementFromLineNumber: (codeView: HTMLElement, line: number, part?: DiffPart) =>
                        codeView.querySelector<HTMLElement>(`[line="${line}"][part="${String(part)}"]`),
                    getLineNumberFromCodeElement: (codeElement: HTMLElement) =>
                        parseInt(codeElement.parentElement!.getAttribute('line')!, 10),
                }
                subscriptions.add(
                    await handleCodeHost({
                        ...commonArguments(),
                        codeHost: {
                            type: 'github',
                            name: 'GitHub',
                            check: () => true,
                            notificationClassNames,
                            codeViewResolvers: [
                                toCodeViewResolver('#code', {
                                    dom,
                                    resolveFileInfo: () => of(diffInfo),
                                }),
                            ],
                        },
                        extensionsController: createMockController(extensionHostAPI),
                        showGlobalDebug: true,
                        platformContext: createMockPlatformContext({}),
                    })
                )
                await wrapRemoteObservable(extensionHostAPI.viewerUpdates())
                    .pipe(
                        filter(({ action }) => action === 'addition'),
                        take(2)
                    )
                    .toPromise()
                const decorationType = extensionAPI.app.createDecorationType()
                const decorated = (editor: ExtensionCodeEditor): Promise<TextDocumentDecoration[] | null> =>
                    wrapRemoteObservable(extensionHostAPI.getTextDecorations({ viewerId: editor.viewerId }))
                        .pipe(first(decorations => !isEmpty(decorations)))
                        .toPromise()

                // Set decorations and verify that a decoration attachment has been added
                const viewers = extensionAPI.app.activeWindow!.visibleViewComponents
                expect(viewers).toHaveLength(2)

                const baseEditor = viewers.find(
                    allOf(
                        isTaggedUnionMember('type', 'CodeEditor' as const),
                        check(editor => editor.document.uri === 'git://foo?1#/bar.ts')
                    )
                ) as ExtensionCodeEditor

                const baseDecorations = [
                    {
                        range: new Range(0, 0, 0, 0),
                        isWholeLine: true,
                        backgroundColor: 'red',
                        after: {
                            contentText: 'test decoration base line 1',
                        },
                    },
                    {
                        range: new Range(1, 0, 1, 0),
                        isWholeLine: true,
                        backgroundColor: 'red',
                        after: {
                            contentText: 'test decoration base line 2',
                        },
                    },
                    {
                        range: new Range(4, 0, 4, 0),
                        isWholeLine: true,
                        backgroundColor: 'red',
                        after: {
                            contentText: 'test decoration base line 5',
                        },
                    },
                ]
                baseEditor.setDecorations(decorationType, baseDecorations)

                const headEditor = viewers.find(
                    allOf(
                        isTaggedUnionMember('type', 'CodeEditor' as const),
                        check(editor => editor.document.uri === 'git://foo?2#/bar.ts')
                    )
                ) as ExtensionCodeEditor
                const headDecorations = [
                    {
                        range: new Range(0, 0, 0, 0),
                        isWholeLine: true,
                        after: {
                            contentText: 'test decoration head line 1',
                        },
                        baze: true,
                    },
                    {
                        range: new Range(1, 0, 1, 0),
                        isWholeLine: true,
                        backgroundColor: 'blue',
                        after: {
                            contentText: 'test decoration head line 2',
                        },
                    },
                    {
                        range: new Range(6, 0, 6, 0),
                        isWholeLine: true,
                        after: {
                            contentText: 'test decoration not visible',
                        },
                    },
                ]
                headEditor.setDecorations(decorationType, headDecorations)

                await Promise.all([decorated(baseEditor), decorated(headEditor)])
                await tick()

                expect(codeView).toMatchSnapshot()

                // Decorate the code view again, and verify that previous decorations
                // are cleaned up and replaced by the new decorations.
                // Remove decoration in first and second line
                baseEditor.setDecorations(decorationType, baseDecorations.slice(2))

                await decorated(baseEditor)
                await tick()
                expect(codeView).toMatchSnapshot()

                // Change decoration in first line
                headEditor.setDecorations(decorationType, [
                    headDecorations[0],
                    {
                        ...headDecorations[1],
                        after: {
                            ...headDecorations[1].after,
                            contentText: 'test decoration head line 2 changed',
                        },
                    },
                    headDecorations[2],
                ])
                await decorated(headEditor)
                // console.log({
                //     headDecs: headEditor.mergedDecorations.value,
                //     baseDecs: baseEditor.mergedDecorations.value,
                // })
                await tick()
                expect(codeView).toMatchSnapshot()
            })
        })

        test('removes code views and models', async () => {
            const { extensionAPI, extensionHostAPI } = await integrationTestContext(undefined, {
                roots: [],
                viewers: [],
            })
            const codeView1 = createTestElement()
            codeView1.className = 'code'
            const codeView2 = createTestElement()
            codeView2.className = 'code'
            const blobInfo: DiffOrBlobInfo = {
                blob: {
                    rawRepoName: 'foo',
                    filePath: '/bar.ts',
                    commitID: '1',
                },
            }
            const mutations = new BehaviorSubject<MutationRecordLike[]>([
                { addedNodes: [document.body], removedNodes: [] },
            ])
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    mutations,
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        notificationClassNames,
                        codeViewResolvers: [
                            toCodeViewResolver('.code', {
                                dom: {
                                    getCodeElementFromTarget: sinon.spy(),
                                    getCodeElementFromLineNumber: sinon.spy(),
                                    getLineElementFromLineNumber: sinon.spy(),
                                    getLineNumberFromCodeElement: sinon.spy(),
                                },
                                resolveFileInfo: codeView =>
                                    codeView === codeView1
                                        ? of(blobInfo)
                                        : of({
                                              blob: {
                                                  ...blobInfo.blob,
                                                  filePath: '/bar2.ts',
                                              },
                                          }),
                            }),
                        ],
                    },
                    extensionsController: createMockController(extensionHostAPI),
                    showGlobalDebug: true,
                    platformContext: createMockPlatformContext(),
                })
            )
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(take(2)).toPromise()

            expect(getEditors(extensionAPI)).toEqual([
                {
                    viewerId: 'viewer#0',
                    isActive: true,
                    resource: 'git://foo?1#/bar.ts',
                    selections: [],
                    type: 'CodeEditor',
                },
                {
                    viewerId: 'viewer#1',
                    isActive: true,
                    resource: 'git://foo?1#/bar2.ts',
                    selections: [],
                    type: 'CodeEditor',
                },
            ])

            // // Simulate codeView1 removal
            mutations.next([{ addedNodes: [], removedNodes: [codeView1] }])
            // One editor should have been removed, model should still exist
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()

            expect(getEditors(extensionAPI)).toEqual([
                {
                    viewerId: 'viewer#1',
                    isActive: true,
                    resource: 'git://foo?1#/bar2.ts',
                    selections: [],
                    type: 'CodeEditor',
                },
            ])
            // // Simulate codeView2 removal
            mutations.next([{ addedNodes: [], removedNodes: [codeView2] }])
            // // Second editor and model should have been removed
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()
            expect(getEditors(extensionAPI)).toEqual([])
        })

        test('Hoverifies a view if the code host has no nativeTooltipResolvers', async () => {
            const { extensionHostAPI, extensionAPI } = await integrationTestContext(undefined, {
                roots: [],
                viewers: [],
            })
            const codeView = createTestElement()
            codeView.id = 'code'
            const codeElement = document.createElement('span')
            codeElement.textContent = 'alert(1)'
            codeView.append(codeElement)
            const dom = {
                getCodeElementFromTarget: sinon.spy(() => codeElement),
                getCodeElementFromLineNumber: sinon.spy(() => codeElement),
                getLineElementFromLineNumber: sinon.spy(() => codeElement),
                getLineNumberFromCodeElement: sinon.spy(() => 1),
            }
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        notificationClassNames,
                        codeViewResolvers: [
                            toCodeViewResolver('#code', {
                                dom,
                                resolveFileInfo: codeView =>
                                    of({
                                        blob: {
                                            rawRepoName: 'foo',
                                            filePath: '/bar.ts',
                                            commitID: '1',
                                        },
                                    }),
                            }),
                        ],
                    },
                    extensionsController: createMockController(extensionHostAPI),
                    showGlobalDebug: true,
                })
            )
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()
            expect(getEditors(extensionAPI).length).toEqual(1)
            await tick()
            codeView.dispatchEvent(new MouseEvent('mouseover'))
            sinon.assert.called(dom.getCodeElementFromTarget)
        })

        test('Does not hoverify a view if the code host has nativeTooltipResolvers and they are enabled from settings', async () => {
            const { extensionHostAPI, extensionAPI } = await integrationTestContext(undefined, {
                roots: [],
                viewers: [],
            })
            const codeView = createTestElement()
            codeView.id = 'code'
            const codeElement = document.createElement('span')
            codeElement.textContent = 'alert(1)'
            codeView.append(codeElement)
            const dom = {
                getCodeElementFromTarget: sinon.spy(() => codeElement),
                getCodeElementFromLineNumber: sinon.spy(() => codeElement),
                getLineElementFromLineNumber: sinon.spy(() => codeElement),
                getLineNumberFromCodeElement: sinon.spy(() => 1),
            }
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        notificationClassNames,
                        nativeTooltipResolvers: [{ selector: '.native', resolveView: element => ({ element }) }],
                        codeViewResolvers: [
                            toCodeViewResolver('#code', {
                                dom,
                                resolveFileInfo: codeView =>
                                    of({
                                        blob: {
                                            rawRepoName: 'foo',
                                            filePath: '/bar.ts',
                                            commitID: '1',
                                        },
                                    }),
                            }),
                        ],
                    },
                    extensionsController: createMockController(extensionHostAPI),
                    showGlobalDebug: true,
                    platformContext: {
                        ...createMockPlatformContext(),
                        settings: of({
                            subjects: [],
                            final: {
                                extensions: {},
                                'codeHost.useNativeTooltips': true,
                            },
                        }),
                    },
                })
            )
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()

            expect(getEditors(extensionAPI).length).toEqual(1)
            await tick()

            codeView.dispatchEvent(new MouseEvent('mouseover'))
            sinon.assert.notCalled(dom.getCodeElementFromTarget)
        })

        test('Hides native tooltips if they are disabled from settings', async () => {
            const { extensionHostAPI, extensionAPI } = await integrationTestContext(undefined, {
                roots: [],
                viewers: [],
            })
            const codeView = createTestElement()
            codeView.id = 'code'
            const codeElement = document.createElement('span')
            codeElement.textContent = 'alert(1)'
            codeView.append(codeElement)
            const nativeTooltip = createTestElement()
            nativeTooltip.classList.add('native')
            const dom = {
                getCodeElementFromTarget: sinon.spy(() => codeElement),
                getCodeElementFromLineNumber: sinon.spy(() => codeElement),
                getLineElementFromLineNumber: sinon.spy(() => codeElement),
                getLineNumberFromCodeElement: sinon.spy(() => 1),
            }
            subscriptions.add(
                await handleCodeHost({
                    ...commonArguments(),
                    codeHost: {
                        type: 'github',
                        name: 'GitHub',
                        check: () => true,
                        notificationClassNames,
                        nativeTooltipResolvers: [{ selector: '.native', resolveView: element => ({ element }) }],
                        codeViewResolvers: [
                            toCodeViewResolver('#code', {
                                dom,
                                resolveFileInfo: codeView =>
                                    of({
                                        blob: {
                                            rawRepoName: 'foo',
                                            filePath: '/bar.ts',
                                            commitID: '1',
                                        },
                                    }),
                            }),
                        ],
                    },
                    extensionsController: createMockController(extensionHostAPI),
                    showGlobalDebug: true,
                    platformContext: {
                        ...createMockPlatformContext(),
                        settings: of({
                            subjects: [],
                            final: {
                                extensions: {},
                                'codeHost.useNativeTooltips': false,
                            },
                        }),
                    },
                })
            )
            await wrapRemoteObservable(extensionHostAPI.viewerUpdates()).pipe(first()).toPromise()
            expect(getEditors(extensionAPI).length).toEqual(1)
            await tick()
            codeView.dispatchEvent(new MouseEvent('mouseover'))
            sinon.assert.called(dom.getCodeElementFromTarget)
            expect(nativeTooltip).toHaveAttribute('data-native-tooltip-hidden', 'true')
        })
    })

    describe('observeHoverOverlayMountLocation()', () => {
        test('emits document.body if the getMountLocationSelector() returns null', () => {
            scheduler().run(({ cold, expectObservable }) => {
                expectObservable(
                    observeHoverOverlayMountLocation(
                        () => null,
                        cold<MutationRecordLike[]>('a', {
                            a: [
                                {
                                    addedNodes: [document.body],
                                    removedNodes: [],
                                },
                            ],
                        })
                    )
                ).toBe('a', {
                    a: document.body,
                })
            })
        })

        test('emits a custom mount location if a node matching the selector is in addedNodes()', () => {
            const element = createTestElement()
            scheduler().run(({ cold, expectObservable }) => {
                expectObservable(
                    observeHoverOverlayMountLocation(
                        () => '.test',
                        cold<MutationRecordLike[]>('-b', {
                            b: [
                                {
                                    addedNodes: [element],
                                    removedNodes: [],
                                },
                            ],
                        })
                    )
                ).toBe('ab', {
                    a: document.body,
                    b: element,
                })
            })
        })

        test('emits a custom mount location if a node matching the selector is nested in an addedNode', () => {
            const element = createTestElement()
            const nested = document.createElement('div')
            nested.classList.add('nested')
            element.append(nested)
            scheduler().run(({ cold, expectObservable }) => {
                expectObservable(
                    observeHoverOverlayMountLocation(
                        () => '.nested',
                        cold<MutationRecordLike[]>('-b', {
                            b: [
                                {
                                    addedNodes: [element],
                                    removedNodes: [],
                                },
                            ],
                        })
                    )
                ).toBe('ab', {
                    a: document.body,
                    b: nested,
                })
            })
        })

        test('emits document.body if a node matching the selector is removed', () => {
            const element = createTestElement()
            scheduler().run(({ cold, expectObservable }) => {
                expectObservable(
                    observeHoverOverlayMountLocation(
                        () => '.test',
                        cold<MutationRecordLike[]>('-bc', {
                            b: [
                                {
                                    addedNodes: [element],
                                    removedNodes: [],
                                },
                            ],
                            c: [
                                {
                                    addedNodes: [],
                                    removedNodes: [element],
                                },
                            ],
                        })
                    )
                ).toBe('abc', {
                    a: document.body,
                    b: element,
                    c: document.body,
                })
            })
        })
    })
})
