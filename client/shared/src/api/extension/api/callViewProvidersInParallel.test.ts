import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'

import { ContributableViewContainer } from '@sourcegraph/client-api'

import { callViewProvidersInParallel } from './callViewProvidersInParallel'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('callViewProviderInParallel', () => {
    describe('with 1 provider', () => {
        it('returns sync view provider result', () => {
            scheduler().run(({ expectObservable }) => {
                const providers = of([
                    {
                        id: 'first view provider',
                        viewProvider: {
                            provideView: () =>
                                of({
                                    title: 'Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                }),
                        },
                    },
                ])

                expectObservable(callViewProvidersInParallel(ContributableViewContainer.InsightsPage, providers)).toBe(
                    '(ab|',
                    {
                        a: [{ id: 'first view provider', view: undefined }],
                        b: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                        ],
                    }
                )
            })
        })
        it('returns async view provider result', () => {
            scheduler().run(({ expectObservable, cold }) => {
                const providers = of([
                    {
                        id: 'first view provider',
                        viewProvider: {
                            provideView: () =>
                                cold('--b', {
                                    b: {
                                        title: 'Title view',
                                        content: [{ component: 'customComponent', props: {} }],
                                    },
                                }),
                        },
                    },
                ])

                expectObservable(callViewProvidersInParallel(ContributableViewContainer.InsightsPage, providers)).toBe(
                    'a)b',
                    {
                        a: [{ id: 'first view provider', view: undefined }],
                        b: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                        ],
                    }
                )
            })
        })
    })

    describe('with 2 providers', () => {
        it('returns two async view providers result', () => {
            scheduler().run(({ expectObservable, cold }) => {
                const providers = of([
                    {
                        id: 'first view provider',
                        viewProvider: {
                            provideView: () =>
                                cold('---d', {
                                    d: {
                                        title: 'First Title view',
                                        content: [{ component: 'customComponent', props: {} }],
                                    },
                                }),
                        },
                    },
                    {
                        id: 'second view provider',
                        viewProvider: {
                            provideView: () =>
                                cold('---i', {
                                    i: {
                                        title: 'Second Title view',
                                        content: [{ component: 'customComponent', props: {} }],
                                    },
                                }),
                        },
                    },
                ])

                expectObservable(callViewProvidersInParallel(ContributableViewContainer.InsightsPage, providers)).toBe(
                    'a)-(bc)',
                    {
                        a: [
                            { id: 'first view provider', view: undefined },
                            { id: 'second view provider', view: undefined },
                        ],
                        b: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'First Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            { id: 'second view provider', view: undefined },
                        ],
                        c: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'First Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            {
                                id: 'second view provider',
                                view: {
                                    title: 'Second Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                        ],
                    }
                )
            })
        })
    })

    describe('with 3 providers', () => {
        it('returns 2 sync view providers and last in the next frame ', () => {
            scheduler().run(({ expectObservable, cold }) => {
                const providers = of([
                    {
                        id: 'first view provider',
                        viewProvider: {
                            provideView: () =>
                                cold('---d|', {
                                    d: {
                                        title: 'First Title view',
                                        content: [{ component: 'customComponent', props: {} }],
                                    },
                                }),
                        },
                    },
                    {
                        id: 'second view provider',
                        viewProvider: {
                            provideView: () =>
                                cold('---i|', {
                                    i: {
                                        title: 'Second Title view',
                                        content: [{ component: 'customComponent', props: {} }],
                                    },
                                }),
                        },
                    },
                    {
                        id: 'third view provider',
                        viewProvider: {
                            provideView: () =>
                                cold('---i|', {
                                    i: {
                                        title: 'Third Title view',
                                        content: [{ component: 'customComponent', props: {} }],
                                    },
                                }),
                        },
                    },
                ])

                expectObservable(callViewProvidersInParallel(ContributableViewContainer.InsightsPage, providers)).toBe(
                    'a)-(bc)d|',
                    {
                        a: [
                            { id: 'first view provider', view: undefined },
                            { id: 'second view provider', view: undefined },
                            { id: 'third view provider', view: undefined },
                        ],
                        b: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'First Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            { id: 'second view provider', view: undefined },
                            { id: 'third view provider', view: undefined },
                        ],
                        c: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'First Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            {
                                id: 'second view provider',
                                view: {
                                    title: 'Second Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            { id: 'third view provider', view: undefined },
                        ],
                        d: [
                            {
                                id: 'first view provider',
                                view: {
                                    title: 'First Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            {
                                id: 'second view provider',
                                view: {
                                    title: 'Second Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                            {
                                id: 'third view provider',
                                view: {
                                    title: 'Third Title view',
                                    content: [{ component: 'customComponent', props: {} }],
                                },
                            },
                        ],
                    }
                )
            })
        })
    })
})
