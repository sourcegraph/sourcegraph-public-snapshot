import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { transformQuery, TransformQuerySignature } from './queryTransformer'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_INPUT = 'foo'
const FIXTURE_RESULT = 'bar'
const FIXTURE_RESULT_TWO = 'qux'
const FIXTURE_RESULT_MERGED = 'foo bar qux'

describe('transformQuery', () => {
    describe('0 providers', () => {
        test('returns original query', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    transformQuery(
                        cold<TransformQuerySignature[]>('-a-|', { a: [] }),
                        FIXTURE_INPUT
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_INPUT,
                })
            ))
    })

    describe('1 provider', () => {
        test('returns result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    transformQuery(
                        cold<TransformQuerySignature[]>('-a-|', {
                            a: [q => of(FIXTURE_RESULT)],
                        }),
                        FIXTURE_INPUT
                    )
                ).toBe('-a-|', { a: FIXTURE_RESULT })
            ))
    })

    describe('2 providers', () => {
        test('returns a single query transformed by both providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    transformQuery(
                        cold<TransformQuerySignature[]>('-a-|', {
                            a: [q => of(`${q} ${FIXTURE_RESULT}`), q => of(`${q} ${FIXTURE_RESULT_TWO}`)],
                        }),
                        FIXTURE_INPUT
                    )
                ).toBe('-a-|', { a: FIXTURE_RESULT_MERGED })
            ))
    })

    describe('Multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    transformQuery(
                        cold<TransformQuerySignature[]>('-a-b-|', {
                            a: [q => of(`${q} ${FIXTURE_RESULT}`)],
                            b: [q => of(`${q} ${FIXTURE_RESULT_TWO}`)],
                        }),
                        FIXTURE_INPUT
                    )
                ).toBe('-a-b-|', {
                    a: `${FIXTURE_INPUT} ${FIXTURE_RESULT}`,
                    b: `${FIXTURE_INPUT} ${FIXTURE_RESULT_TWO}`,
                })
            ))
    })
})
