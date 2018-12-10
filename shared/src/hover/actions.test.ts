import { HoveredToken, LOADER_DELAY } from '@sourcegraph/codeintellify'
import { Location } from '@sourcegraph/extension-api-types'
import { createMemoryHistory } from 'history'
import { from, of } from 'rxjs'
import { first, map } from 'rxjs/operators'
// tslint:disable-next-line:no-submodule-imports
import { TestScheduler } from 'rxjs/testing'
import { ActionItemProps } from '../actions/ActionItem'
import { EMPTY_MODEL, Model } from '../api/client/model'
import { Services } from '../api/client/services'
import { CommandRegistry } from '../api/client/services/command'
import { ContributionRegistry } from '../api/client/services/contribution'
import { ProvideTextDocumentLocationSignature } from '../api/client/services/location'
import { ContributableMenu, ReferenceParams, TextDocumentPositionParams } from '../api/protocol'
import { getContributedActionItems } from '../contributions/contributions'
import { EMPTY_SETTINGS_CASCADE } from '../settings/settings'
import { toPrettyBlobURL } from '../util/url'
import { getDefinitionURL, getHoverActionsContext, HoverActionsContext, registerHoverContributions } from './actions'
import { HoverContext } from './HoverOverlay'

const FIXTURE_PARAMS: TextDocumentPositionParams = {
    textDocument: { uri: 'git://r?c#f' },
    position: { line: 1, character: 1 },
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
    rev: 'v',
    filePath: 'f',
    line: 2,
    character: 2,
}

function testModelService(
    roots: Model['roots'] = [{ uri: 'git://r3?c3', inputRevision: 'v3' }]
): { model: { value: Pick<Model, 'roots'> } } {
    return { model: { value: { roots } } }
}

// Use toPrettyBlobURL as the urlToFile passed to these functions because it results in the most readable/familiar
// expected test output.
const urlToFile = toPrettyBlobURL

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('getHoverActionsContext', () => {
    test('shows a loader for the definition if slow', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    getHoverActionsContext(
                        {
                            extensionsController: {
                                services: {
                                    model: testModelService(),
                                    textDocumentDefinition: {
                                        getLocations: () =>
                                            cold<Location[]>(`- ${LOADER_DELAY}ms --- d`, { d: [FIXTURE_LOCATION] }),
                                    },
                                    textDocumentReferences: {
                                        providersForDocument: () =>
                                            cold<ProvideTextDocumentLocationSignature<ReferenceParams, Location>[]>(
                                                'a',
                                                { a: [() => of(null)] }
                                            ),
                                    },
                                },
                            },
                            platformContext: { urlToFile },
                        },
                        FIXTURE_HOVER_CONTEXT
                    )
                )
                // tslint:disable-next-line:no-object-literal-type-assertion
            ).toBe(`a ${LOADER_DELAY - 1}ms (bc)d`, {
                a: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                b: {
                    'goToDefinition.showLoading': true,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                c: {
                    'goToDefinition.showLoading': true,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                },
                d: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                },
            } as { [key: string]: HoverActionsContext })
        ))

    test('shows no loader for the definition if fast', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    getHoverActionsContext(
                        {
                            extensionsController: {
                                services: {
                                    model: testModelService(),
                                    textDocumentDefinition: {
                                        getLocations: () => cold<Location[]>(`-b`, { b: [FIXTURE_LOCATION] }),
                                    },
                                    textDocumentReferences: {
                                        providersForDocument: () =>
                                            cold<ProvideTextDocumentLocationSignature<ReferenceParams, Location>[]>(
                                                'a',
                                                { a: [() => of(null)] }
                                            ),
                                    },
                                },
                            },
                            platformContext: { urlToFile },
                        },
                        FIXTURE_HOVER_CONTEXT
                    )
                )
                // tslint:disable-next-line:no-object-literal-type-assertion
            ).toBe(`a(bc)`, {
                a: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                b: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': null,
                    hoverPosition: FIXTURE_PARAMS,
                },
                c: {
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': '/r2@c2/-/blob/f2#L3:3',
                    'goToDefinition.notFound': false,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                },
            } as { [key: string]: HoverActionsContext })
        ))
})

describe('getDefinitionURL', () => {
    test('emits null if the locations result is null', async () =>
        expect(
            getDefinitionURL(
                { urlToFile },
                {
                    model: testModelService(),
                    textDocumentDefinition: { getLocations: () => of(null) },
                },
                FIXTURE_PARAMS
            )
                .pipe(first())
                .toPromise()
        ).resolves.toBe(null))

    test('emits null if the locations result is empty', async () =>
        expect(
            getDefinitionURL(
                { urlToFile },
                {
                    model: testModelService(),
                    textDocumentDefinition: { getLocations: () => of([]) },
                },
                FIXTURE_PARAMS
            )
                .pipe(first())
                .toPromise()
        ).resolves.toBe(null))

    describe('if there is exactly 1 location result', () => {
        describe('when the result is inside the current root', () => {
            test('emits the definition URL the user input revision (not commit SHA) of the root', async () =>
                expect(
                    getDefinitionURL(
                        { urlToFile },
                        {
                            model: testModelService(),
                            textDocumentDefinition: { getLocations: () => of<Location[]>([{ uri: 'git://r3?c3#f' }]) },
                        },
                        FIXTURE_PARAMS
                    )
                        .pipe(first())
                        .toPromise()
                ).resolves.toEqual({ url: '/r3@v3/-/blob/f', multiple: false }))
        })

        describe('when the result is not inside the current root (different repo and/or commit)', () => {
            test('emits the definition URL with range', async () =>
                expect(
                    getDefinitionURL(
                        { urlToFile },
                        {
                            model: testModelService(),
                            textDocumentDefinition: { getLocations: () => of<Location[]>([FIXTURE_LOCATION]) },
                        },
                        FIXTURE_PARAMS
                    )
                        .pipe(first())
                        .toPromise()
                ).resolves.toEqual({ url: '/r2@c2/-/blob/f2#L3:3', multiple: false }))

            test('emits the definition URL without range', async () =>
                expect(
                    getDefinitionURL(
                        { urlToFile },
                        {
                            model: testModelService(),
                            textDocumentDefinition: {
                                getLocations: () => of<Location[]>([{ ...FIXTURE_LOCATION, range: undefined }]),
                            },
                        },
                        FIXTURE_PARAMS
                    )
                        .pipe(first())
                        .toPromise()
                ).resolves.toEqual({ url: '/r2@c2/-/blob/f2', multiple: false }))
        })
    })

    test('emits the definition panel URL if there is more than 1 location result', async () =>
        expect(
            getDefinitionURL(
                { urlToFile },
                {
                    model: testModelService([{ uri: 'git://r?c', inputRevision: 'v' }]),
                    textDocumentDefinition: {
                        getLocations: () => of<Location[]>([FIXTURE_LOCATION, { ...FIXTURE_LOCATION, uri: 'other' }]),
                    },
                },
                FIXTURE_PARAMS
            )
                .pipe(first())
                .toPromise()
        ).resolves.toEqual({ url: '/r@v/-/blob/f#L2:2&tab=def', multiple: true }))
})

describe('registerHoverContributions', () => {
    const contribution = new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
    const commands = new CommandRegistry()
    const textDocumentDefinition: Pick<Services['textDocumentDefinition'], 'getLocations'> = {
        getLocations: () => of(null),
    }
    const history = createMemoryHistory()
    const subscription = registerHoverContributions({
        extensionsController: {
            services: {
                contribution,
                commands,
                model: testModelService(),
                textDocumentDefinition,
            },
        },
        platformContext: { urlToFile },
        history,
    })
    afterAll(() => subscription.unsubscribe())

    const getHoverActions = (context: HoverActionsContext) =>
        contribution
            .getContributions(undefined, context)
            .pipe(
                first(),
                map(contributions => getContributedActionItems(contributions, ContributableMenu.Hover))
            )
            .toPromise()

    describe('getHoverActions', () => {
        const GO_TO_DEFINITION_ACTION: ActionItemProps = {
            action: {
                command: 'goToDefinition',
                commandArguments: ['{"textDocument":{"uri":"git://r?c#f"},"position":{"line":1,"character":1}}'],
                id: 'goToDefinition',
                title: 'Go to definition',
            },
            altAction: undefined,
        }
        const GO_TO_DEFINITION_PRELOADED_ACTION: ActionItemProps = {
            action: {
                command: 'open',
                commandArguments: ['/r2@c2/-/blob/f2#L3:3'],
                id: 'goToDefinition.preloaded',
                title: 'Go to definition',
            },
            altAction: undefined,
        }
        const FIND_REFERENCES_ACTION: ActionItemProps = {
            action: {
                command: 'open',
                commandArguments: ['/r@v/-/blob/f#L2:2&tab=references'],
                id: 'findReferences',
                title: 'Find references',
            },
            altAction: undefined,
        }

        test('shows goToDefinition (non-preloaded) when the definition is loading', async () =>
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

        test('shows goToDefinition (non-preloaded) when the definition had an error', async () =>
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

        test('hides goToDefinition when the definition was not found', async () =>
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

        test('shows goToDefinition.preloaded when goToDefinition.url is available', async () =>
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

        test('shows findReferences when the definition exists', async () =>
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

        test('hides findReferences when the definition might exist (and is still loading)', async () =>
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

        test('shows findReferences when the definition had an error', async () =>
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

        test('shows findReferences when the definition was not found', async () =>
            expect(
                getHoverActions({
                    'goToDefinition.showLoading': false,
                    'goToDefinition.url': null,
                    'goToDefinition.notFound': true,
                    'goToDefinition.error': false,
                    'findReferences.url': '/r@v/-/blob/f#L2:2&tab=references',
                    hoverPosition: FIXTURE_PARAMS,
                })
            ).resolves.toEqual([FIND_REFERENCES_ACTION]))
    })

    describe('goToDefinition command', () => {
        test('reports no definition found', async () => {
            textDocumentDefinition.getLocations = () => of(null) // mock
            return expect(
                commands.executeCommand({ command: 'goToDefinition', arguments: [JSON.stringify(FIXTURE_PARAMS)] })
            ).rejects.toMatchObject({ message: 'No definition found.' })
        })

        test('reports panel already visible', async () => {
            textDocumentDefinition.getLocations = () =>
                of([FIXTURE_LOCATION, { ...FIXTURE_LOCATION, uri: 'git://r3?v3#f3' }]) // mock
            history.push('/r@c/-/blob/f#L2:2&tab=def')
            return expect(
                commands.executeCommand({ command: 'goToDefinition', arguments: [JSON.stringify(FIXTURE_PARAMS)] })
            ).rejects.toMatchObject({ message: 'Multiple definitions shown in panel below.' })
        })

        test('reports already at the definition', async () => {
            textDocumentDefinition.getLocations = () => of([FIXTURE_LOCATION]) // mock
            history.push('/r2@c2/-/blob/f2#L3:3')
            return expect(
                commands.executeCommand({ command: 'goToDefinition', arguments: [JSON.stringify(FIXTURE_PARAMS)] })
            ).rejects.toMatchObject({ message: 'Already at the definition.' })
        })
    })
})
