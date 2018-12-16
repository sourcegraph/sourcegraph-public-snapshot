import * as assert from 'assert'
import { Observable, of } from 'rxjs'
import { Subscription } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { ContributableMenu, Contributions } from '../../protocol'
import { Context, ContributionScope } from '../context/context'
import { EMPTY_COMPUTED_CONTEXT } from '../context/expr/evaluator'
import { EMPTY_MODEL, Model } from '../model'
import {
    contextFilter,
    ContributionRegistry,
    ContributionsEntry,
    evaluateContributions,
    filterContributions,
    mergeContributions,
} from './contribution'

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

const FIXTURE_CONTRIBUTIONS_1: Contributions = {
    actions: [{ id: '1.a', command: 'c', title: '1.A' }, { id: '1.b', command: 'c', title: '1.B' }],
    menus: {
        [ContributableMenu.CommandPalette]: [{ action: '1.a' }],
        [ContributableMenu.GlobalNav]: [{ action: '1.a' }, { action: '1.b' }],
    },
}

const FIXTURE_CONTRIBUTIONS_2: Contributions = {
    actions: [{ id: '2.a', command: 'c', title: '2.A' }, { id: '2.b', command: 'c', title: '2.B' }],
    menus: {
        [ContributableMenu.CommandPalette]: [{ action: '2.a' }],
        [ContributableMenu.EditorTitle]: [{ action: '2.a' }, { action: '2.b' }],
    },
}

const FIXTURE_CONTRIBUTIONS_MERGED: Contributions = {
    actions: [
        { id: '1.a', command: 'c', title: '1.A' },
        { id: '1.b', command: 'c', title: '1.B' },
        { id: '2.a', command: 'c', title: '2.A' },
        { id: '2.b', command: 'c', title: '2.B' },
    ],
    menus: {
        [ContributableMenu.CommandPalette]: [{ action: '1.a' }, { action: '2.a' }],
        [ContributableMenu.GlobalNav]: [{ action: '1.a' }, { action: '1.b' }],
        [ContributableMenu.EditorTitle]: [{ action: '2.a' }, { action: '2.b' }],
    },
}

describe('ContributionRegistry', () => {
    it('is initially empty', () => {
        assert.deepStrictEqual(
            new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({})).entries.value,
            []
        )
    })

    it('registers and unregisters contributions', () => {
        const subscriptions = new Subscription()
        const registry = new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
        const entry1: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_1 }
        const entry2: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_2 }

        const unregister1 = subscriptions.add(registry.registerContributions(entry1))
        assert.deepStrictEqual(registry.entries.value, [entry1])

        const unregister2 = subscriptions.add(registry.registerContributions(entry2))
        assert.deepStrictEqual(registry.entries.value, [entry1, entry2])

        unregister1.unsubscribe()
        assert.deepStrictEqual(registry.entries.value, [entry2])

        unregister2.unsubscribe()
        assert.deepStrictEqual(registry.entries.value, [])
    })

    it('replaces contributions', () => {
        const registry = new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
        const entry1: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_1 }
        const entry2: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_2 }

        const unregister1 = registry.registerContributions(entry1)
        assert.deepStrictEqual(registry.entries.value, [entry1])

        const unregister2 = registry.replaceContributions(unregister1, entry2)
        assert.deepStrictEqual(registry.entries.value, [entry2])

        unregister1.unsubscribe()
        assert.deepStrictEqual(registry.entries.value, [entry2])

        unregister2.unsubscribe()
        assert.deepStrictEqual(registry.entries.value, [])
    })

    describe('contributions observable', () => {
        it('emits stream of results of registrations', () => {
            const registry = new class extends ContributionRegistry {
                public getContributionsFromEntries(
                    entries: Observable<ContributionsEntry[]>
                ): Observable<Contributions> {
                    return super.getContributionsFromEntries(entries, undefined)
                }
            }(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    registry.getContributionsFromEntries(
                        cold<ContributionsEntry[]>('-a-b-c-|', {
                            a: [{ contributions: FIXTURE_CONTRIBUTIONS_1 }],
                            b: [{ contributions: FIXTURE_CONTRIBUTIONS_1 }, { contributions: {} }],
                            c: [{ contributions: FIXTURE_CONTRIBUTIONS_1 }, { contributions: FIXTURE_CONTRIBUTIONS_2 }],
                        })
                    )
                ).toBe('-a---c-|', {
                    a: FIXTURE_CONTRIBUTIONS_1,
                    c: FIXTURE_CONTRIBUTIONS_MERGED,
                })
            )
        })

        it('supports registration of an observable', () => {
            const registry = new class extends ContributionRegistry {
                public getContributionsFromEntries(
                    entries: Observable<ContributionsEntry[]>
                ): Observable<Contributions> {
                    return super.getContributionsFromEntries(entries, undefined)
                }
            }(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    registry.getContributionsFromEntries(
                        cold<ContributionsEntry[]>('-a-----|', {
                            a: [
                                {
                                    contributions: cold<Contributions | Contributions[]>('--b-c-|', {
                                        b: FIXTURE_CONTRIBUTIONS_1,
                                        c: [FIXTURE_CONTRIBUTIONS_2],
                                    }),
                                },
                            ],
                        })
                    )
                ).toBe('---b-c-|', {
                    b: FIXTURE_CONTRIBUTIONS_1,
                    c: FIXTURE_CONTRIBUTIONS_2,
                })
            )
        })

        it('emits when context changes and filters on context', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new class extends ContributionRegistry {
                    public constructor() {
                        super(
                            cold<Model>('-a-b-|', {
                                a: EMPTY_MODEL,
                                b: EMPTY_MODEL,
                            }),
                            {
                                data: cold<SettingsCascadeOrError>('-a-b-|', {
                                    a: EMPTY_SETTINGS_CASCADE,
                                    b: EMPTY_SETTINGS_CASCADE,
                                }),
                            },
                            cold<Context>('-a-b-|', { a: { x: 1, y: 2 }, b: { x: 1, y: 1 } })
                        )
                    }

                    public getContributionsFromEntries(
                        entries: Observable<ContributionsEntry[]>
                    ): Observable<Contributions> {
                        return super.getContributionsFromEntries(entries, undefined)
                    }
                }()
                expectObservable(
                    registry.getContributionsFromEntries(
                        of([
                            {
                                contributions: { menus: { commandPalette: [{ action: 'a', when: 'x == y' }] } },
                            },
                        ] as ContributionsEntry[])
                    )
                ).toBe('-a-b-|', {
                    a: { menus: { commandPalette: [] } } as Contributions,
                    b: { menus: { commandPalette: [{ action: 'a', when: 'x == y' }] } } as Contributions,
                })
            })
        })

        it('continues after error thrown during evaluation', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new class extends ContributionRegistry {
                    public constructor() {
                        super(
                            cold<Model>('a', { a: EMPTY_MODEL }),
                            { data: cold<SettingsCascadeOrError>('a', { a: EMPTY_SETTINGS_CASCADE }) },
                            cold<Context>('a', {})
                        )
                    }

                    public getContributionsFromEntries(
                        entries: Observable<ContributionsEntry[]>,
                        scope?: ContributionScope
                    ): Observable<Contributions> {
                        return super.getContributionsFromEntries(entries, scope, () => void 0 /* noop log */)
                    }
                }()
                expectObservable(
                    registry.getContributionsFromEntries(
                        of([
                            {
                                // Expression "!" will cause an error to be thrown.
                                contributions: { menus: { commandPalette: [{ action: 'a1', when: '!' }] } },
                            },
                            {
                                contributions: { menus: { commandPalette: [{ action: 'a2' }] } },
                            },
                        ] as ContributionsEntry[])
                    )
                ).toBe('a', {
                    a: { menus: { commandPalette: [{ action: 'a2' }] } } as Contributions,
                })
            })
        })
    })
})

describe('mergeContributions', () => {
    it('handles an empty array', () => assert.deepStrictEqual(mergeContributions([]), {}))
    it('handles an single item', () =>
        assert.deepStrictEqual(mergeContributions([FIXTURE_CONTRIBUTIONS_1]), FIXTURE_CONTRIBUTIONS_1))
    it('handles multiple items', () =>
        assert.deepStrictEqual(
            mergeContributions([FIXTURE_CONTRIBUTIONS_1, FIXTURE_CONTRIBUTIONS_2, {}, { actions: [] }, { menus: {} }]),
            FIXTURE_CONTRIBUTIONS_MERGED
        ))
})

const FIXTURE_CONTEXT = new Map<string, any>(
    Object.entries({
        a: true,
        b: false,
    })
)

describe('contextFilter', () => {
    it('filters', () =>
        assert.deepStrictEqual(
            contextFilter(
                FIXTURE_CONTEXT,
                [{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a' }, { x: 4, when: 'b' }],
                x => x === 'a'
            ),
            [{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a' }]
        ))
})

describe('filterContributions', () => {
    it('handles empty contributions', () =>
        assert.deepStrictEqual(filterContributions(EMPTY_COMPUTED_CONTEXT, {}), {} as Contributions))

    it('handles empty menu contributions', () =>
        assert.deepStrictEqual(filterContributions(EMPTY_COMPUTED_CONTEXT, { menus: {} }), {
            menus: {},
        } as Contributions))

    it('handles empty array of menu contributions', () =>
        assert.deepStrictEqual(filterContributions(EMPTY_COMPUTED_CONTEXT, { menus: { commandPalette: [] } }), {
            menus: { commandPalette: [] },
        } as Contributions))

    it('handles non-empty contributions', () =>
        assert.deepStrictEqual(
            filterContributions(
                FIXTURE_CONTEXT,
                {
                    actions: [{ id: 'a1', command: 'c' }, { id: 'a2', command: 'c' }, { id: 'a3', command: 'c' }],
                    menus: {
                        [ContributableMenu.CommandPalette]: [
                            { action: 'a1', when: 'a' },
                            { action: 'a2', when: 'b' },
                            { action: 'a3' },
                        ],
                        [ContributableMenu.GlobalNav]: [
                            { action: 'a1', when: 'a' },
                            { action: 'a2' },
                            { action: 'a3', when: 'b' },
                        ],
                    },
                },
                x => x === 'a'
            ),
            {
                actions: [{ id: 'a1', command: 'c' }, { id: 'a2', command: 'c' }, { id: 'a3', command: 'c' }],
                menus: {
                    [ContributableMenu.CommandPalette]: [{ action: 'a1', when: 'a' }, { action: 'a3' }],
                    [ContributableMenu.GlobalNav]: [{ action: 'a1', when: 'a' }, { action: 'a2' }],
                },
            } as Contributions
        ))

    it('throws an error if an error occurs during evaluation', () => {
        const input: Contributions = {
            actions: [{ id: 'a', command: 'c', title: 'a' }],
            menus: { commandPalette: [{ action: 'a', when: 'a' }] },
        }
        assert.throws(() =>
            filterContributions(FIXTURE_CONTEXT, input, () => {
                throw new Error('')
            })
        )
    })
})

// tslint:disable:no-invalid-template-strings
describe('evaluateContributions', () => {
    const TEST_TEMPLATE_EVALUATOR = {
        evaluateTemplate: () => 'x',
        needsEvaluation: (template: string) => template === 'a',
    }

    it('handles empty contributions', () =>
        assert.deepStrictEqual(evaluateContributions(EMPTY_COMPUTED_CONTEXT, {}), {} as Contributions))

    it('handles empty array of command contributions', () =>
        assert.deepStrictEqual(evaluateContributions(EMPTY_COMPUTED_CONTEXT, { actions: [] }), {
            actions: [],
        } as Contributions))

    it('handles non-empty contributions', () => {
        const input: Contributions = {
            actions: [
                {
                    id: 'a1',
                    command: 'c1',
                    commandArguments: ['a', 'b', 'a'],
                    title: 'a',
                    category: 'a',
                    description: 'a',
                    iconURL: 'a',
                    actionItem: {
                        label: 'a',
                        description: 'a',
                        iconDescription: 'a',
                        iconURL: 'a',
                    },
                },
                { id: 'a2', command: 'c2', title: 'a', category: 'b' },
                { id: 'a3', command: 'c3', title: 'b', category: 'b', actionItem: { label: 'a', description: 'b' } },
            ],
        }
        const origInput = JSON.parse(JSON.stringify(input))
        assert.deepStrictEqual(evaluateContributions(FIXTURE_CONTEXT, input, TEST_TEMPLATE_EVALUATOR), {
            actions: [
                {
                    id: 'a1',
                    command: 'c1',
                    commandArguments: ['x', 'b', 'x'],
                    title: 'x',
                    category: 'x',
                    description: 'x',
                    iconURL: 'x',
                    actionItem: {
                        label: 'x',
                        description: 'x',
                        iconDescription: 'x',
                        iconURL: 'x',
                    },
                },
                { id: 'a2', command: 'c2', title: 'x', category: 'b' },
                { id: 'a3', command: 'c3', title: 'b', category: 'b', actionItem: { label: 'x', description: 'b' } },
            ],
        } as Contributions)
        assert.deepStrictEqual(input, origInput, 'input must not be mutated')
    })

    it('supports commandArguments with the first element non-evaluated', () => {
        assert.deepStrictEqual(
            evaluateContributions(
                FIXTURE_CONTEXT,
                {
                    actions: [{ id: 'a', command: 'c', commandArguments: ['b', 'a', 'b', 'a'] }],
                },
                TEST_TEMPLATE_EVALUATOR
            ),
            {
                actions: [{ id: 'a', command: 'c', commandArguments: ['b', 'x', 'b', 'x'] }],
            } as Contributions
        )
    })

    const TEST_THROW_EVALUATOR = {
        evaluateTemplate: () => {
            throw new Error('')
        },
        needsEvaluation: () => true,
    }

    it('does not evaluate the `id` or `command` values', () => {
        assert.deepStrictEqual(
            evaluateContributions(
                FIXTURE_CONTEXT,
                {
                    actions: [{ id: 'a', command: 'c' }],
                },
                TEST_THROW_EVALUATOR
            ),
            {
                actions: [{ id: 'a', command: 'c' }],
            } as Contributions
        )
    })

    it('throws an error if an error occurs during evaluation', () => {
        const input: Contributions = { actions: [{ id: 'a', command: 'c', title: 'a' }] }
        assert.throws(() => evaluateContributions(FIXTURE_CONTEXT, input, TEST_THROW_EVALUATOR))
    })
})
// tslint:enable:no-invalid-template-strings
