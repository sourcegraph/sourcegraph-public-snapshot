import * as assert from 'assert'
import { Observable, of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ContributableViewContainer } from '../../protocol'
import * as plain from '../../protocol/plainTypes'
import { Entry } from './registry'
import { getView, getViews, ViewProviderRegistrationOptions } from './view'

const FIXTURE_CONTAINER = ContributableViewContainer.Panel

const FIXTURE_ENTRY_1: Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>> = {
    registrationOptions: { container: FIXTURE_CONTAINER, id: '1' },
    provider: of<plain.PanelView>({ title: 't1', content: 'c1' }),
}
const FIXTURE_RESULT_1 = { container: FIXTURE_CONTAINER, id: '1', title: 't1', content: 'c1' }

const FIXTURE_ENTRY_2: Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>> = {
    registrationOptions: { container: FIXTURE_CONTAINER, id: '2' },
    provider: of<plain.PanelView>({ title: 't2', content: 'c2' }),
}
const FIXTURE_RESULT_2 = { container: FIXTURE_CONTAINER, id: '2', title: 't2', content: 'c2' }

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

describe('getView', () => {
    describe('0 providers', () => {
        it('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getView(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-|', { a: [] }),
                        '1'
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    it('returns result from provider', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getView(
                    cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-|', {
                        a: [FIXTURE_ENTRY_1],
                    }),
                    '1'
                )
            ).toBe('-a-|', {
                a: FIXTURE_RESULT_1,
            })
        ))

    describe('multiple emissions', () => {
        it('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getView(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-b-|', {
                            a: [FIXTURE_ENTRY_1],
                            b: [FIXTURE_ENTRY_1, FIXTURE_ENTRY_2],
                        }),
                        '2'
                    )
                ).toBe('-a-b-|', {
                    a: null,
                    b: FIXTURE_RESULT_2,
                })
            ))
    })
})

describe('getViews', () => {
    describe('0 providers', () => {
        it('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getViews(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-|', { a: [] }),
                        FIXTURE_CONTAINER
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    it('returns result from provider', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getViews(
                    cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-|', {
                        a: [FIXTURE_ENTRY_1],
                    }),
                    FIXTURE_CONTAINER
                )
            ).toBe('-a-|', {
                a: [FIXTURE_RESULT_1],
            })
        ))

    it('continues if provider has error', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getViews(
                    cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-|', {
                        a: [
                            {
                                registrationOptions: { container: FIXTURE_CONTAINER, id: 'err' },
                                provider: throwError('err'),
                            },
                            FIXTURE_ENTRY_1,
                        ],
                    }),
                    FIXTURE_CONTAINER
                )
            ).toBe('-a-|', {
                a: [FIXTURE_RESULT_1],
            })
        ))

    describe('multiple emissions', () => {
        it('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getViews(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>('-a-b-|', {
                            a: [FIXTURE_ENTRY_1],
                            b: [FIXTURE_ENTRY_1, FIXTURE_ENTRY_2],
                        }),
                        FIXTURE_CONTAINER
                    )
                ).toBe('-a-b-|', {
                    a: [FIXTURE_RESULT_1],
                    b: [FIXTURE_RESULT_1, FIXTURE_RESULT_2],
                })
            ))
    })
})
