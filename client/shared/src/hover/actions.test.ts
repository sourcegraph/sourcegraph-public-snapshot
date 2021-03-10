import { HoveredToken, LOADER_DELAY, MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { Location } from '@sourcegraph/extension-api-types'
import { createMemoryHistory, MemoryHistory, createPath } from 'history'
import { from, Observable, of, throwError, Subscription } from 'rxjs'
import { first, map, switchMap } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import * as sinon from 'sinon'
import { ActionItemAction } from '../actions/ActionItem'
import { ContributableMenu, TextDocumentPositionParameters } from '../api/protocol'
import { PrivateRepoPublicSourcegraphComError } from '../backend/errors'
import { getContributedActionItems } from '../contributions/contributions'
import { SuccessGraphQLResult } from '../graphql/graphql'
import { PlatformContext, URLToFileContext } from '../platform/context'
import { resetAllMemoizationCaches } from '../util/memoizeObservable'
import {
    FileSpec,
    UIPositionSpec,
    RawRepoSpec,
    RepoSpec,
    RevisionSpec,
    ViewStateSpec,
    toAbsoluteBlobURL,
    toPrettyBlobURL,
} from '../util/url'
import { getDefinitionURL, getHoverActionsContext, HoverActionsContext, registerHoverContributions } from './actions'
import { HoverContext } from './HoverOverlay'
import { pretendRemote } from '../api/util'
import { FlatExtensionHostAPI } from '../api/contract'
import { proxySubscribable } from '../api/extension/api/common'
import { Remote } from 'comlink'
import { integrationTestContext } from '../api/integration-test/testHelpers'
import { wrapRemoteObservable } from '../api/client/api/common'
import { ExposedToClient } from '../api/client/mainthread-api'
import { WorkspaceRootWithMetadata } from '../api/extension/flatExtensionApi'

const FIXTURE_PARAMS: TextDocumentPositionParameters & URLToFileContext = {
    textDocument: { uri: 'git://r?c#f' },
    position: { line: 1, character: 1 },
    part: undefined,
}

const FIXTURE_LOCATION: Location = {
    uri: 'git://r2?c2#f2',
    range: {
        start: { line: 2, character: 2 },
        end: { line: 3, character: 3 },
    },
}

const FIXTURE_HOVER_CONTEXT: HoveredToken & HoverContext = {
    repoName: 'r',
    commitID: 'c',
    revision: 'v',
    filePath: 'f',
    line: 2,
    character: 2,
}

const FIXTURE_WORKSPACE: WorkspaceRootWithMetadata = { uri: 'git://r3?c3', inputRevision: 'v3' }

// Use toPrettyBlobURL as the urlToFile passed to these functions because it results in the most readable/familiar
// expected test output.
// Some tests may override this with .callsFake()
let urlToFile!: sinon.SinonStub<Parameters<PlatformContext['urlToFile']>, string>
beforeEach(() => {
    urlToFile = sinon.stub<Parameters<PlatformContext['urlToFile']>, string>().callsFake(toPrettyBlobURL)
})

const requestGraphQL: PlatformContext['requestGraphQL'] = <R>({ variables }: { variables: { [key: string]: any } }) =>
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    of({
        data: {
            repository: {
                uri: variables.repoName,
                mirrorInfo: {
                    cloned: true,
                },
            },
        },
    } as SuccessGraphQLResult<any>)

const scheduler = (): TestScheduler => new TestScheduler((actual, expected) => expect(actual).toStrictEqual(expected))

// TODO(tj): use fake timers + integration test setup

describe('getHoverActionsContext', () => {
    beforeEach(() => resetAllMemoizationCaches())
    it('shows a loader for the definition if slow', () => {
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    getHoverActionsContext(
                        {
                            getDefinition: () =>
                                cold<MaybeLoadingResult<Location[]>>(`l ${LOADER_DELAY + 100}ms r`, {
                                    l: { isLoading: true, result: [] },
                                    r: { isLoading: false, result: [FIXTURE_LOCATION] },
                                }),
                            extensionsController: {
                                extHostAPI: Promise.resolve(
                                    pretendRemote<FlatExtensionHostAPI>({
                                        getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        hasReferenceProvidersForDocument: () =>
                                            proxySubscribable(
                                                cold<boolean>('a', { a: true })
                                            ),
                                    })
                                ),
                            },
                            platformContext: { urlToFile, requestGraphQL },
                        },
                        FIXTURE_HOVER_CONTEXT
                    )
                )
            ).toBe(
                `n ${LOADER_DELAY - 1}ms (lf) 97ms g`,
                ((): {
                    [key: string]: HoverActionsContext
                } => ({
                    // Show nothing
                    n: {
                        'goToDefinition.showLoading': false,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': null,
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Show loader
                    l: {
                        'goToDefinition.showLoading': true,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': null,
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Show find references button (same tick)
                    f: {
                        'goToDefinition.showLoading': true,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Show go to definition button, hide loader, show find references button
                    g: {
                        'goToDefinition.showLoading': false,
                        'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                        hoverPosition: FIXTURE_PARAMS,
                    },
                }))()
            )
        )
    })

    it('shows a loader when definition providers are registered after invocation and a find-references button after the result returned', () => {
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    getHoverActionsContext(
                        {
                            getDefinition: () =>
                                cold<MaybeLoadingResult<Location[]>>(`l e 50ms l ${LOADER_DELAY}ms r`, {
                                    l: { isLoading: true, result: [] },
                                    e: { isLoading: false, result: [] },
                                    r: { isLoading: false, result: [FIXTURE_LOCATION] },
                                }),
                            extensionsController: {
                                extHostAPI: Promise.resolve(
                                    pretendRemote<FlatExtensionHostAPI>({
                                        getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        hasReferenceProvidersForDocument: () =>
                                            proxySubscribable(
                                                cold<boolean>('a', { a: true })
                                            ),
                                    })
                                ),
                            },
                            platformContext: { urlToFile, requestGraphQL },
                        },
                        FIXTURE_HOVER_CONTEXT
                    )
                )
            ).toBe(
                `e n ${LOADER_DELAY - 2}ms (lf) 49ms g`,
                ((): { [key: string]: HoverActionsContext } => ({
                    // Show nothing
                    e: {
                        'goToDefinition.showLoading': false,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': null,
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Not found
                    n: {
                        'goToDefinition.showLoading': false,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': true,
                        'goToDefinition.error': false,
                        'findReferences.url': null,
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Show loader
                    l: {
                        'goToDefinition.showLoading': true,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': null,
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Show find references button (same tick)
                    f: {
                        'goToDefinition.showLoading': true,
                        'goToDefinition.url': null,
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                        hoverPosition: FIXTURE_PARAMS,
                    },
                    // Show go to definition button, hide loader, show find references button
                    g: {
                        'goToDefinition.showLoading': false,
                        'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                        'goToDefinition.notFound': false,
                        'goToDefinition.error': false,
                        'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                        hoverPosition: FIXTURE_PARAMS,
                    },
                }))()
            )
        )
    })

    it('shows the find references button when reference providers are registered later', () => {
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    getHoverActionsContext(
                        {
                            getDefinition: () =>
                                cold<MaybeLoadingResult<Location[]>>('-b', {
                                    b: { isLoading: false, result: [FIXTURE_LOCATION] },
                                }),
                            extensionsController: {
                                extHostAPI: Promise.resolve(
                                    pretendRemote<FlatExtensionHostAPI>({
                                        getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        hasReferenceProvidersForDocument: () =>
                                            proxySubscribable(
                                                cold<boolean>('a 10ms b', { a: false, b: true })
                                            ),
                                    })
                                ),
                            },
                            platformContext: { urlToFile, requestGraphQL },
                        },
                        FIXTURE_HOVER_CONTEXT
                    )
                )
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            ).toBe('ng 9ms f', {
                // Nothing
                n: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                // Go to definition
                g: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                // Find references
                f: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                },
            } as {
                [key: string]: HoverActionsContext
            })
        )
    })

    it('shows no loader for the definition if fast', () => {
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    getHoverActionsContext(
                        {
                            getDefinition: () =>
                                cold<MaybeLoadingResult<Location[]>>('-b', {
                                    b: { isLoading: false, result: [FIXTURE_LOCATION] },
                                }),
                            extensionsController: {
                                extHostAPI: Promise.resolve(
                                    pretendRemote<FlatExtensionHostAPI>({
                                        getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        hasReferenceProvidersForDocument: () =>
                                            proxySubscribable(
                                                cold<boolean>('a', { a: true })
                                            ),
                                    })
                                ),
                            },
                            platformContext: { urlToFile, requestGraphQL },
                        },
                        FIXTURE_HOVER_CONTEXT
                    )
                )
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            ).toBe('n(gf)', {
                // Nothing
                n: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                // Go to definition
                g: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                // Find references button
                f: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                },
            } as {
                [key: string]: HoverActionsContext
            })
        )
    })
})

describe('getDefinitionURL', () => {
    beforeEach(() => resetAllMemoizationCaches())

    it('emits null if the locations result is empty', () =>
        expect(
            of({ isLoading: false, result: [] })
                .pipe(
                    getDefinitionURL(
                        { urlToFile, requestGraphQL },
                        {
                            extHostAPI: Promise.resolve(
                                pretendRemote<FlatExtensionHostAPI>({
                                    getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                })
                            ),
                        },
                        FIXTURE_PARAMS
                    ),
                    first(({ isLoading }) => !isLoading)
                )
                .toPromise()
        ).resolves.toStrictEqual({ isLoading: false, result: null }))

    describe('if there is exactly 1 location result', () => {
        it('resolves the raw repo name and passes it to urlToFile()', async () => {
            const requestGraphQL = <R>({ variables }: { variables: any }): Observable<SuccessGraphQLResult<R>> =>
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                of({
                    data: {
                        repository: {
                            uri: `github.com/${variables.repoName as string}`,
                            mirrorInfo: {
                                cloned: true,
                            },
                        },
                    },
                } as SuccessGraphQLResult<any>)
            const urlToFile = sinon.spy(
                (
                    _location: RepoSpec &
                        Partial<RawRepoSpec> &
                        RevisionSpec &
                        FileSpec &
                        Partial<UIPositionSpec> &
                        Partial<ViewStateSpec>
                ) => ''
            )
            await of<MaybeLoadingResult<Location[]>>({
                isLoading: false,
                result: [{ uri: 'git://r3?c3#f' }],
            })
                .pipe(
                    getDefinitionURL(
                        { urlToFile, requestGraphQL },
                        {
                            extHostAPI: Promise.resolve(
                                pretendRemote<FlatExtensionHostAPI>({
                                    getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                })
                            ),
                        },
                        FIXTURE_PARAMS
                    ),
                    first(({ isLoading }) => !isLoading)
                )
                .toPromise()
            sinon.assert.calledOnce(urlToFile)
            expect(urlToFile.getCalls()[0].args[0]).toMatchObject({
                filePath: 'f',
                position: undefined,
                rawRepoName: 'github.com/r3',
                repoName: 'r3',
                revision: 'v3',
            })
        })

        it('fails gracefully when resolveRawRepoName() fails with a PrivateRepoPublicSourcegraph error', async () => {
            const requestGraphQL = (): Observable<never> =>
                throwError(new PrivateRepoPublicSourcegraphComError('ResolveRawRepoName'))
            const urlToFile = sinon.spy()
            await of<MaybeLoadingResult<Location[]>>({
                isLoading: false,
                result: [{ uri: 'git://r3?c3#f' }],
            })
                .pipe(
                    getDefinitionURL(
                        { urlToFile, requestGraphQL },
                        {
                            extHostAPI: Promise.resolve(
                                pretendRemote<FlatExtensionHostAPI>({
                                    getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                })
                            ),
                        },
                        FIXTURE_PARAMS
                    ),
                    first(({ isLoading }) => !isLoading)
                )
                .toPromise()
            sinon.assert.calledOnce(urlToFile)
            sinon.assert.calledWith(urlToFile, {
                commitID: undefined,
                filePath: 'f',
                position: undefined,
                range: undefined,
                rawRepoName: 'r3',
                repoName: 'r3',
                revision: 'v3',
            })
        })

        describe('when the result is inside the current root', () => {
            it('emits the definition URL the user input revision (not commit SHA) of the root', () =>
                expect(
                    of<MaybeLoadingResult<Location[]>>({
                        isLoading: false,
                        result: [{ uri: 'git://r3?c3#f' }],
                    })
                        .pipe(
                            getDefinitionURL(
                                { urlToFile, requestGraphQL },
                                {
                                    extHostAPI: Promise.resolve(
                                        pretendRemote<FlatExtensionHostAPI>({
                                            getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        })
                                    ),
                                },
                                FIXTURE_PARAMS
                            ),
                            first(({ isLoading }) => !isLoading)
                        )
                        .toPromise()
                ).resolves.toEqual({ isLoading: false, result: { url: '/r3@v3/-/blob/f', multiple: false } }))
        })

        describe('when the result is not inside the current root (different repo and/or commit)', () => {
            it('emits the definition URL with range', () =>
                expect(
                    of<MaybeLoadingResult<Location[]>>({
                        isLoading: false,
                        result: [FIXTURE_LOCATION],
                    })
                        .pipe(
                            getDefinitionURL(
                                { urlToFile, requestGraphQL },
                                {
                                    extHostAPI: Promise.resolve(
                                        pretendRemote<FlatExtensionHostAPI>({
                                            getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        })
                                    ),
                                },
                                FIXTURE_PARAMS
                            ),
                            first(({ isLoading }) => !isLoading)
                        )
                        .toPromise()
                ).resolves.toEqual({ isLoading: false, result: { url: '/r2@c2/-/blob/f2#L3:3', multiple: false } }))

            it('emits the definition URL without range', () =>
                expect(
                    of<MaybeLoadingResult<Location[]>>({
                        isLoading: false,
                        result: [{ ...FIXTURE_LOCATION, range: undefined }],
                    })
                        .pipe(
                            getDefinitionURL(
                                { urlToFile, requestGraphQL },
                                {
                                    extHostAPI: Promise.resolve(
                                        pretendRemote<FlatExtensionHostAPI>({
                                            getWorkspaceRoots: () => [FIXTURE_WORKSPACE],
                                        })
                                    ),
                                },
                                FIXTURE_PARAMS
                            ),
                            first(({ isLoading }) => !isLoading)
                        )
                        .toPromise()
                ).resolves.toEqual({ isLoading: false, result: { url: '/r2@c2/-/blob/f2', multiple: false } }))
        })
    })

    it('emits the definition panel URL if there is more than 1 location result', () =>
        expect(
            of<MaybeLoadingResult<Location[]>>({
                isLoading: false,
                result: [FIXTURE_LOCATION, { ...FIXTURE_LOCATION, uri: 'other' }],
            })
                .pipe(
                    getDefinitionURL(
                        { urlToFile, requestGraphQL },
                        {
                            extHostAPI: Promise.resolve(
                                pretendRemote<FlatExtensionHostAPI>({
                                    getWorkspaceRoots: () => [{ uri: 'git://r?c', inputRevision: 'v' }],
                                })
                            ),
                        },
                        FIXTURE_PARAMS
                    ),
                    first()
                )
                .toPromise()
        ).resolves.toEqual({ isLoading: false, result: { url: '/r@v/-/blob/f#L2:2&tab=def', multiple: true } }))
})

describe('registerHoverContributions()', () => {
    const subscription = new Subscription()
    let history!: MemoryHistory

    let extensionHostAPI!: Promise<Remote<FlatExtensionHostAPI>>
    let exposedToClient!: ExposedToClient
    let locationAssign!: sinon.SinonSpy<[string], void>
    beforeEach(async () => {
        resetAllMemoizationCaches()

        const { extensionHostAPI: API, exposedToClient: clientAPI, unsubscribe } = await integrationTestContext({})
        subscription.add(() => unsubscribe())
        extensionHostAPI = Promise.resolve(API)
        exposedToClient = clientAPI

        history = createMemoryHistory()
        locationAssign = sinon.spy((_url: string) => undefined)
        subscription.add(
            registerHoverContributions({
                extensionsController: {
                    extHostAPI: Promise.resolve(extensionHostAPI),
                    registerCommand: exposedToClient.registerCommand,
                },
                platformContext: { urlToFile, requestGraphQL },
                history,
                locationAssign,
            })
        )
    })
    afterAll(() => subscription.unsubscribe())

    const getHoverActions = (context: HoverActionsContext): Promise<ActionItemAction[]> =>
        from(extensionHostAPI)
            .pipe(
                switchMap(extensionHostAPI =>
                    wrapRemoteObservable(extensionHostAPI.getContributions(undefined, context))
                ),
                first(),
                map(contributions => getContributedActionItems(contributions, ContributableMenu.Hover))
            )
            .toPromise()

    describe('getHoverActions()', () => {
        const GO_TO_DEFINITION_ACTION: ActionItemAction = {
            action: {
                command: 'goToDefinition',
                commandArguments: ['{"textDocument":{"uri":"git://r?c#f"},"position":{"line":1,"character":1}}'],
                id: 'goToDefinition',
                title: 'Go to definition',
                actionItem: undefined,
                category: undefined,
                description: undefined,
                iconURL: undefined,
            },
            altAction: undefined,
        }
        const GO_TO_DEFINITION_PRELOADED_ACTION: ActionItemAction = {
            action: {
                command: 'open',
                commandArguments: ['/r2@c2/-/blob/f2#L3:3'],
                id: 'goToDefinition.preloaded',
                title: 'Go to definition',
            },
            altAction: undefined,
        }
        const FIND_REFERENCES_ACTION: ActionItemAction = {
            action: {
                command: 'open',
                commandArguments: ['/r@v/-/blob/f#L2:2&tab=references'],
                id: 'findReferences',
                title: 'Find references',
            },
            altAction: undefined,
        }

        it('shows goToDefinition (non-preloaded) when the definition is loading', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': true,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([GO_TO_DEFINITION_ACTION]))

        it('shows goToDefinition (non-preloaded) when the definition had an error', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': true,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([GO_TO_DEFINITION_ACTION]))

        it('hides goToDefinition when the definition was not found', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': true,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([]))

        it('shows goToDefinition.preloaded when goToDefinition.url is available', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([GO_TO_DEFINITION_PRELOADED_ACTION]))

        it('shows findReferences when the definition exists', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([GO_TO_DEFINITION_PRELOADED_ACTION, FIND_REFERENCES_ACTION]))

        it('hides findReferences when the definition might exist (and is still loading)', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': true,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([GO_TO_DEFINITION_ACTION, FIND_REFERENCES_ACTION]))

        it('shows findReferences when the definition had an error', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': true,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([GO_TO_DEFINITION_ACTION, FIND_REFERENCES_ACTION]))

        it('does not show findReferences when the definition was not found', () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': true,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([]))
    })

    describe('goToDefinition command', () => {
        test('reports no definition found', async () => {
            extensionHostAPI = Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getDefinition: () => proxySubscribable(of({ isLoading: false, result: [] })), // mock
                })
            )

            await expect(
                exposedToClient.executeCommand({ command: 'goToDefinition', args: [JSON.stringify(FIXTURE_PARAMS)] })
            ).rejects.toMatchObject({ message: 'No definition found.' })
        })

        test('navigates to an in-app URL using the passed history object', async () => {
            jsdom.reconfigure({ url: 'https://sourcegraph.test/r2@c2/-/blob/f1' })
            history.replace('/r2@c2/-/blob/f1')
            expect(history).toHaveLength(1)

            extensionHostAPI = Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getDefinition: () => proxySubscribable(of({ isLoading: false, result: [] })), // mock
                })
            )

            await exposedToClient.executeCommand({ command: 'goToDefinition', args: [JSON.stringify(FIXTURE_PARAMS)] })
            sinon.assert.notCalled(locationAssign)
            expect(history).toHaveLength(2)
            expect(createPath(history.location)).toBe('/r2@c2/-/blob/f2#L3:3')
        })

        test('navigates to an external URL using the global location object', async () => {
            jsdom.reconfigure({ url: 'https://github.test/r2@c2/-/blob/f1' })
            history.replace('/r2@c2/-/blob/f1')
            expect(history).toHaveLength(1)
            urlToFile.callsFake(toAbsoluteBlobURL.bind(null, 'https://sourcegraph.test'))

            extensionHostAPI = Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getDefinition: () =>
                        proxySubscribable(
                            of({
                                isLoading: false,
                                result: [FIXTURE_LOCATION, { ...FIXTURE_LOCATION, uri: 'git://r3?v3#f3' }],
                            })
                        ), // mock
                })
            )

            await exposedToClient.executeCommand({ command: 'goToDefinition', args: [JSON.stringify(FIXTURE_PARAMS)] })
            sinon.assert.calledOnce(locationAssign)
            sinon.assert.calledWith(locationAssign, 'https://sourcegraph.test/r@c/-/blob/f#L2:2&tab=def')
            expect(history).toHaveLength(1)
        })

        test('reports panel already visible', async () => {
            extensionHostAPI = Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getDefinition: () =>
                        proxySubscribable(
                            of({
                                isLoading: false,
                                result: [FIXTURE_LOCATION, { ...FIXTURE_LOCATION, uri: 'git://r3?v3#f3' }],
                            })
                        ), // mock
                })
            )

            history.push('/r@c/-/blob/f#L2:2&tab=def')
            await expect(
                exposedToClient.executeCommand({ command: 'goToDefinition', args: [JSON.stringify(FIXTURE_PARAMS)] })
            ).rejects.toMatchObject({ message: 'Multiple definitions shown in panel below.' })
        })

        test('reports already at the definition', async () => {
            extensionHostAPI = Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getDefinition: () =>
                        proxySubscribable(
                            of({
                                isLoading: false,
                                result: [FIXTURE_LOCATION],
                            })
                        ), // mock
                })
            )

            history.push('/r2@c2/-/blob/f2#L3:3')
            await expect(
                exposedToClient.executeCommand({ command: 'goToDefinition', args: [JSON.stringify(FIXTURE_PARAMS)] })
            ).rejects.toBe('Already at the definition.')
        })
    })
})
