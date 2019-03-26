import { EMPTY, Observable, of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ContributableViewContainer } from '../../protocol'
import { Entry } from './registry'
import { getView, getViews, PanelViewWithComponent, ViewProviderRegistrationOptions } from './view'

const FIXTURE_CONTAINER = ContributableViewContainer.Panel

const FIXTURE_ENTRY_1: Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>> = {
    registrationOptions: { container: FIXTURE_CONTAINER, id: '1' },
    provider: of<PanelViewWithComponent>({
        title: 't1',
        content: 'c1',
        priority: 0,
        locationsOrCustom: { locations: EMPTY },
    }),
}
const FIXTURE_RESULT_1 = { container: FIXTURE_CONTAINER, id: '1', title: 't1', content: 'c1', priority: 0 }

const FIXTURE_ENTRY_2: Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>> = {
    registrationOptions: { container: FIXTURE_CONTAINER, id: '2' },
    provider: of<PanelViewWithComponent>({
        title: 't2',
        content: 'c2',
        priority: 0,
        locationsOrCustom: { locations: EMPTY },
    }),
}
const FIXTURE_RESULT_2 = { container: FIXTURE_CONTAINER, id: '2', title: 't2', content: 'c2', priority: 0 }

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('getView', () => {
    describe('0 providers', () => {
        test('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getView(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-|', {
                            a: [],
                        }),
                        '1'
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    test('returns result from provider', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getView(
                    cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-|', {
                        a: [FIXTURE_ENTRY_1],
                    }),
                    '1'
                )
            ).toBe('-a-|', {
                a: FIXTURE_RESULT_1,
            })
        ))

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getView(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-b-|', {
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
        test('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getViews(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-|', {
                            a: [],
                        }),
                        FIXTURE_CONTAINER
                    )
                ).toBe('-a-|', {
                    a: [],
                })
            ))
    })

    test('returns result from provider', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getViews(
                    cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-|', {
                        a: [FIXTURE_ENTRY_1],
                    }),
                    FIXTURE_CONTAINER
                )
            ).toBe('-a-|', {
                a: [FIXTURE_RESULT_1],
            })
        ))

    test('continues if provider has error', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getViews(
                    cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-|', {
                        a: [
                            {
                                registrationOptions: { container: FIXTURE_CONTAINER, id: 'err' },
                                provider: throwError('err'),
                            },
                            FIXTURE_ENTRY_1,
                        ],
                    }),
                    FIXTURE_CONTAINER,
                    false
                )
            ).toBe('-a-|', {
                a: [FIXTURE_RESULT_1],
            })
        ))

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getViews(
                        cold<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent>>[]>('-a-b-|', {
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
