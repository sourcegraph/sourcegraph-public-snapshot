import { Selection } from '@sourcegraph/extension-api-classes'
import { Location } from '@sourcegraph/extension-api-types'
import { Observable, of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { TextDocumentRegistrationOptions } from '../../protocol'
import {
    getLocationsFromProviders,
    ProvideTextDocumentLocationSignature,
    TextDocumentLocationProviderRegistry,
} from './location'
import { Entry } from './registry'
import { FIXTURE } from './registry.test'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_LOCATION: Location = {
    uri: 'file:///f',
    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
}
const FIXTURE_LOCATIONS: Location | Location[] | null = [FIXTURE_LOCATION, FIXTURE_LOCATION]

/**
 * Allow overriding {@link TextDocumentLocationProviderRegistry#entries} for tests.
 */
class TestTextDocumentLocationProviderRegistry extends TextDocumentLocationProviderRegistry {
    constructor(entries?: Observable<Entry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>[]>) {
        super()
        if (entries) {
            entries.subscribe(entries => this.entries.next(entries))
        }
    }
}

describe('TextDocumentLocationProviderRegistry', () => {
    describe('hasProvidersForActiveTextDocument', () => {
        test('false if no position params', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new TestTextDocumentLocationProviderRegistry(
                    cold<Entry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>[]>('a', {
                        a: [{ provider: () => of(null), registrationOptions: { documentSelector: ['*'] } }],
                    })
                )
                expectObservable(registry.hasProvidersForActiveTextDocument([])).toBe('a', {
                    a: false,
                })
            })
        })

        test('true if matching document', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new TestTextDocumentLocationProviderRegistry(
                    cold<Entry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>[]>('a', {
                        a: [{ provider: () => of(null), registrationOptions: { documentSelector: ['l'] } }],
                    })
                )
                expectObservable(
                    registry.hasProvidersForActiveTextDocument([
                        {
                            isActive: true,
                            editorId: 'editor#0',
                            type: 'CodeEditor' as const,
                            selections: [new Selection(1, 2, 3, 4).toPlain()],
                            resource: 'u',
                            model: { languageId: 'l' },
                        },
                    ])
                ).toBe('a', {
                    a: true,
                })
            })
        })

        test('false if no matching document', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new TestTextDocumentLocationProviderRegistry(
                    cold<Entry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>[]>('a', {
                        a: [{ provider: () => of(null), registrationOptions: { documentSelector: ['otherlang'] } }],
                    })
                )
                expectObservable(
                    registry.hasProvidersForActiveTextDocument([
                        {
                            isActive: true,
                            editorId: 'editor#0',
                            type: 'CodeEditor' as const,
                            selections: [new Selection(1, 2, 3, 4).toPlain()],
                            resource: 'u',
                            model: { languageId: 'l' },
                        },
                    ])
                ).toBe('a', {
                    a: false,
                })
            })
        })
    })
})

describe('getLocationsFromProviders', () => {
    describe('0 providers', () => {
        test('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', { a: [] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: cold<Location[] | null>('(a|)', { a: null }),
                })
            ))
    })

    describe('1 provider', () => {
        test('returns null result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', { a: [() => of(null)] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: cold<Location[] | null>('(a|)', { a: null }),
                })
            ))

        test('returns result array from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_LOCATIONS)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: cold<Location[] | null>('(a|)', { a: FIXTURE_LOCATIONS }),
                })
            ))
    })

    test('errors do not propagate', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getLocationsFromProviders(
                    cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                        a: [() => of([FIXTURE_LOCATION]), () => throwError(new Error('x'))],
                    }),
                    FIXTURE.TextDocumentPositionParams,
                    false
                )
            ).toBe('-a-|', {
                a: cold<Location[] | null>('(a|)', { a: [FIXTURE_LOCATION] }),
            })
        ))

    describe('2 providers', () => {
        test('returns null result if both providers return null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(null), () => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: cold<Location[] | null>('(a|)', { a: null }),
                })
            ))

        test('omits null result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_LOCATIONS), () => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: cold<Location[] | null>('(a|)', { a: FIXTURE_LOCATIONS }),
                })
            ))

        test('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [
                                () =>
                                    of([
                                        {
                                            uri: 'file:///f1',
                                            range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                                        },
                                    ]),
                                () =>
                                    of([
                                        {
                                            uri: 'file:///f2',
                                            range: { start: { line: 5, character: 6 }, end: { line: 7, character: 8 } },
                                        },
                                    ]),
                            ],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: cold<Location[] | null>('(a|)', {
                        a: [
                            {
                                uri: 'file:///f1',
                                range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                            },
                            {
                                uri: 'file:///f2',
                                range: { start: { line: 5, character: 6 }, end: { line: 7, character: 8 } },
                            },
                        ],
                    }),
                })
            ))
    })

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocationsFromProviders(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-b-|', {
                            a: [() => of(FIXTURE_LOCATIONS)],
                            b: [() => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-b-|', {
                    a: cold<Location[] | null>('(a|)', { a: FIXTURE_LOCATIONS }),
                    b: cold<Location[] | null>('(a|)', { a: null }),
                })
            ))
    })
})
