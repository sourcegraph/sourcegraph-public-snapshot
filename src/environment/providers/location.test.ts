import * as assert from 'assert'
import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { Location, Position, Range } from 'vscode-languageserver-types'
import {
    getLocation,
    getLocations,
    getLocationsWithExtensionID,
    ProvideTextDocumentLocationSignature,
} from './location'
import { FIXTURE } from './registry.test'

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

const FIXTURE_LOCATION: Location = {
    uri: 'file:///f',
    range: Range.create(Position.create(1, 2), Position.create(3, 4)),
}
const FIXTURE_LOCATIONS: Location | Location[] | null = [FIXTURE_LOCATION, FIXTURE_LOCATION]

describe('getLocation', () => {
    describe('0 providers', () => {
        it('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', { a: [] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    describe('1 provider', () => {
        it('returns null result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', { a: [() => of(null)] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        it('returns result array from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_LOCATIONS)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_LOCATIONS,
                })
            ))

        it('returns single result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_LOCATION)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_LOCATION,
                })
            ))
    })

    describe('2 providers', () => {
        it('returns null result if both providers return null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(null), () => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        it('omits null result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_LOCATIONS), () => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_LOCATIONS,
                })
            ))

        it('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                            a: [
                                () =>
                                    of({
                                        uri: 'file:///f1',
                                        range: { start: Position.create(1, 2), end: Position.create(3, 4) },
                                    }),
                                () =>
                                    of({
                                        uri: 'file:///f2',
                                        range: { start: Position.create(5, 6), end: Position.create(7, 8) },
                                    }),
                            ],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: [
                        {
                            uri: 'file:///f1',
                            range: { start: Position.create(1, 2), end: Position.create(3, 4) },
                        },
                        {
                            uri: 'file:///f2',
                            range: { start: Position.create(5, 6), end: Position.create(7, 8) },
                        },
                    ],
                })
            ))
    })

    describe('multiple emissions', () => {
        it('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getLocation(
                        cold<ProvideTextDocumentLocationSignature[]>('-a-b-|', {
                            a: [() => of(FIXTURE_LOCATIONS)],
                            b: [() => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-b-|', {
                    a: FIXTURE_LOCATIONS,
                    b: null,
                })
            ))
    })
})

describe('getLocations', () => {
    it('wraps single result in array', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getLocations(
                    cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                        a: [() => of(FIXTURE_LOCATION)],
                    }),
                    FIXTURE.TextDocumentPositionParams
                )
            ).toBe('-a-|', {
                a: [FIXTURE_LOCATION],
            })
        ))

    it('preserves array results', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getLocations(
                    cold<ProvideTextDocumentLocationSignature[]>('-a-|', {
                        a: [() => of(FIXTURE_LOCATIONS)],
                    }),
                    FIXTURE.TextDocumentPositionParams
                )
            ).toBe('-a-|', {
                a: FIXTURE_LOCATIONS,
            })
        ))
})

describe('getLocationsWithExtensionID', () => {
    it('wraps single result in array', () =>
        scheduler().run(({ cold, expectObservable }) => {
            const res = getLocationsWithExtensionID(
                cold<{ extensionID: string; provider: ProvideTextDocumentLocationSignature }[]>('-a-|', {
                    a: [{ extensionID: 'test', provider: () => of(FIXTURE_LOCATION) }],
                }),
                FIXTURE.TextDocumentPositionParams
            )
            expectObservable(res).toBe('-a-|', {
                a: [{ extensionID: 'test', location: FIXTURE_LOCATION }],
            })
        }))
    it('preserves array results', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getLocationsWithExtensionID(
                    cold<{ extensionID: string; provider: ProvideTextDocumentLocationSignature }[]>('-a-|', {
                        a: [{ extensionID: 'test', provider: () => of(FIXTURE_LOCATIONS) }],
                    }),
                    FIXTURE.TextDocumentPositionParams
                )
            ).toBe('-a-|', {
                a: FIXTURE_LOCATIONS.map(l => ({ extensionID: 'test', location: l })),
            })
        ))
})
