const RENDER = jest.fn()
jest.mock('react-dom', () => ({
    createPortal: jest.fn(el => el),
    render: RENDER,
    unmountComponentAtNode: jest.fn(),
}))

import { uniqueId } from 'lodash'
import MutationObserver from 'mutation-observer'
import renderer from 'react-test-renderer'
import { from, NEVER, of, Subject, Subscription } from 'rxjs'
import { filter, map, skip, switchMap, take } from 'rxjs/operators'
import { Services } from '../../../../../shared/src/api/client/services'
import { Range } from '../../../../../shared/src/api/extension/types/range'
import { integrationTestContext } from '../../../../../shared/src/api/integration-test/testHelpers'
import { Controller } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isDefined } from '../../../../../shared/src/util/types'
import { FileInfo, handleCodeHost } from './code_intelligence'

const elementRenderedAtMount = (mount: Element): renderer.ReactTestRendererJSON | undefined => {
    const call = RENDER.mock.calls.find(call => call[1] === mount)
    return call && call[0]
}

jest.mock('uuid', () => ({
    v4: () => 'uuid',
}))

const createMockController = (services: Services): Controller => ({
    services,
    notifications: NEVER,
    executeCommand: jest.fn(),
    unsubscribe: jest.fn(),
})

const createMockPlatformContext = (
    partialMocks?: Partial<PlatformContextProps<'forceUpdateTooltip' | 'sideloadedExtensionURL' | 'urlToFile'>>
): PlatformContextProps<'forceUpdateTooltip' | 'sideloadedExtensionURL' | 'urlToFile'> => ({
    platformContext: {
        forceUpdateTooltip: jest.fn(),
        urlToFile: jest.fn(),
        sideloadedExtensionURL: new Subject<string | null>(),
        ...partialMocks,
    },
})

describe('handleCodeHost()', () => {
    beforeAll(() => {
        // jsdom doesn't support MutationObserver or IntersectionObserver, so we need to mock them
        ;(window as any).MutationObserver = MutationObserver
        ;(window as any).IntersectionObserver = class {
            constructor(
                private callback: (entries: Pick<IntersectionObserverEntry, 'target' | 'isIntersecting'>[]) => void
            ) {}
            public observe = (el: Element) => setTimeout(() => this.callback([{ isIntersecting: true, target: el }]), 0)
        }
    })

    afterAll(() => {
        delete (window as any).MutationObserver
        delete (window as any).IntersectionObserver
    })

    let subscriptions = new Subscription()

    afterEach(() => {
        for (const el of document.querySelectorAll('.test')) {
            el.remove()
        }
        RENDER.mockClear()
        subscriptions.unsubscribe()
        subscriptions = new Subscription()
    })

    const createTestElement = () => {
        const el = document.createElement('div')
        el.className = `test test-${uniqueId()}`
        document.body.appendChild(el)
        return el
    }

    test('renders the hoverlay container', async () => {
        const { services } = await integrationTestContext()
        subscriptions.add(
            handleCodeHost({
                codeHost: {
                    name: 'test',
                    check: () => true,
                },
                extensionsController: createMockController(services),
                showGlobalDebug: false,
                ...createMockPlatformContext(),
            })
        )
        const overlayMount = document.body.firstChild! as HTMLElement
        expect(overlayMount.className).toBe('overlay-mount-container')
        const renderedOverlay = elementRenderedAtMount(overlayMount)
        expect(renderedOverlay).not.toBeUndefined()
    })

    test('renders the command palette if codeHost.getCommandPaletteMount is defined', async () => {
        const { services } = await integrationTestContext()
        const commandPaletteMount = createTestElement()
        subscriptions.add(
            handleCodeHost({
                codeHost: {
                    name: 'test',
                    check: () => true,
                    getCommandPaletteMount: () => commandPaletteMount,
                },
                extensionsController: createMockController(services),
                showGlobalDebug: false,
                ...createMockPlatformContext(),
            })
        )
        const renderedCommandPalette = elementRenderedAtMount(commandPaletteMount)
        expect(renderedCommandPalette).not.toBeUndefined()
    })

    test('creates a .global-debug element and renders the debug palette if showGlobalDebug is true', async () => {
        const { services } = await integrationTestContext()
        subscriptions.add(
            handleCodeHost({
                codeHost: {
                    name: 'test',
                    check: () => true,
                },
                extensionsController: createMockController(services),
                showGlobalDebug: true,
                ...createMockPlatformContext(),
            })
        )
        const globalDebugMount = document.querySelector('.global-debug')
        expect(globalDebugMount).not.toBeUndefined()
        const renderedDebugElement = elementRenderedAtMount(globalDebugMount!)
        expect(renderedDebugElement).not.toBeUndefined()
    })

    test('renders the debug palette to the provided mount if codeHost.globalDebugMount is defined', async () => {
        const { services } = await integrationTestContext()
        const globalDebugMount = createTestElement()
        subscriptions.add(
            handleCodeHost({
                codeHost: {
                    name: 'test',
                    check: () => true,
                    getGlobalDebugMount: () => globalDebugMount,
                },
                extensionsController: createMockController(services),
                showGlobalDebug: true,
                ...createMockPlatformContext(),
            })
        )
        const renderedDebugElement = elementRenderedAtMount(globalDebugMount)
        expect(renderedDebugElement).not.toBeUndefined()
    })

    test('detects code views based on selectors', async () => {
        const { services } = await integrationTestContext()
        const codeView = createTestElement()
        codeView.id = 'code'
        const toolbarMount = document.createElement('div')
        codeView.appendChild(toolbarMount)
        const fileInfo: FileInfo = {
            repoName: 'foo',
            filePath: '/bar.ts',
            commitID: '1',
        }
        subscriptions.add(
            handleCodeHost({
                codeHost: {
                    name: 'test',
                    check: () => true,
                    codeViewSpecs: [
                        {
                            selector: `#code`,
                            dom: {
                                getCodeElementFromTarget: jest.fn(),
                                getCodeElementFromLineNumber: jest.fn(),
                                getLineNumberFromCodeElement: jest.fn(),
                            },
                            resolveFileInfo: codeView => of(fileInfo),
                            getToolbarMount: () => toolbarMount,
                        },
                    ],
                    selectionsChanges: () => of([]),
                },
                extensionsController: createMockController(services),
                showGlobalDebug: true,
                ...createMockPlatformContext(),
            })
        )
        const viewComponents = await from(services.model.model)
            .pipe(
                skip(1),
                take(1),
                map(({ visibleViewComponents }) => visibleViewComponents)
            )
            .toPromise()
        expect(viewComponents).toEqual([
            {
                isActive: true,
                item: {
                    languageId: 'typescript',
                    text: undefined,
                    uri: 'git://foo?1#/bar.ts',
                },
                selections: [],
                type: 'textEditor',
            },
        ])
        expect(codeView.classList.contains('sg-mounted')).toBe(true)
        const toolbar = elementRenderedAtMount(toolbarMount)
        expect(toolbar).not.toBeUndefined()
    })

    test('decorates a code view', async () => {
        const { extensionAPI, services } = await integrationTestContext(undefined, {
            roots: [],
            visibleViewComponents: [],
        })
        const codeView = createTestElement()
        codeView.id = 'code'
        const fileInfo: FileInfo = {
            repoName: 'foo',
            filePath: '/bar.ts',
            commitID: '1',
        }
        const line = document.createElement('div')
        codeView.appendChild(line)
        subscriptions.add(
            handleCodeHost({
                codeHost: {
                    name: 'test',
                    check: () => true,
                    codeViewSpecs: [
                        {
                            selector: `#code`,
                            dom: {
                                getCodeElementFromTarget: jest.fn(),
                                getCodeElementFromLineNumber: () => line,
                                getLineNumberFromCodeElement: jest.fn(),
                            },
                            resolveFileInfo: codeView => of(fileInfo),
                        },
                    ],
                    selectionsChanges: () => of([]),
                },
                extensionsController: createMockController(services),
                showGlobalDebug: true,
                ...createMockPlatformContext(),
            })
        )
        const activeEditor = await from(extensionAPI.app.activeWindowChanges)
            .pipe(
                filter(isDefined),
                switchMap(window => window.activeViewComponentChanges),
                filter(isDefined),
                take(1)
            )
            .toPromise()
        const decorationType = extensionAPI.app.createDecorationType()
        const decorated = () =>
            services.textDocumentDecoration
                .getDecorations({ uri: 'git://foo?1#/bar.ts' })
                .pipe(
                    filter(decorations => decorations !== []),
                    take(1)
                )
                .toPromise()

        // Set decorations and verify that a decoration attachment has been added
        activeEditor.setDecorations(decorationType, [
            {
                range: new Range(0, 0, 0, 0),
                after: {
                    contentText: 'test decoration',
                },
            },
        ])
        await decorated()
        expect(line.querySelectorAll('.line-decoration-attachment').length).toBe(1)
        expect(line.querySelector('.line-decoration-attachment')!.textContent).toEqual('test decoration')

        // Decorate the code view again, and verify that previous decorations
        // are cleaned up and replaced by the new decorations.
        activeEditor.setDecorations(decorationType, [
            {
                range: new Range(0, 0, 0, 0),
                after: {
                    contentText: 'test decoration 2',
                },
            },
        ])
        await decorated()
        expect(line.querySelectorAll('.line-decoration-attachment').length).toBe(1)
        expect(line.querySelector('.line-decoration-attachment')!.textContent).toEqual('test decoration 2')
    })
})
