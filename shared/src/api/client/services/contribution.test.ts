import { Observable, of } from 'rxjs'
import { Subscription } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { ContributableMenu, Contributions, EvaluatedContributions } from '../../protocol'
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

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

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
    test('is initially empty', () => {
        expect(
            new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({})).entries.value
        ).toEqual([])
    })

    test('registers and unregisters contributions', () => {
        const subscriptions = new Subscription()
        const registry = new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
        const entry1: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_1 }
        const entry2: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_2 }

        const unregister1 = subscriptions.add(registry.registerContributions(entry1))
        expect(registry.entries.value).toEqual([entry1])

        const unregister2 = subscriptions.add(registry.registerContributions(entry2))
        expect(registry.entries.value).toEqual([entry1, entry2])

        unregister1.unsubscribe()
        expect(registry.entries.value).toEqual([entry2])

        unregister2.unsubscribe()
        expect(registry.entries.value).toEqual([])
    })

    test('replaces contributions', () => {
        const registry = new ContributionRegistry(of(EMPTY_MODEL), { data: of(EMPTY_SETTINGS_CASCADE) }, of({}))
        const entry1: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_1 }
        const entry2: ContributionsEntry = { contributions: FIXTURE_CONTRIBUTIONS_2 }

        const unregister1 = registry.registerContributions(entry1)
        expect(registry.entries.value).toEqual([entry1])

        const unregister2 = registry.replaceContributions(unregister1, entry2)
        expect(registry.entries.value).toEqual([entry2])

        unregister1.unsubscribe()
        expect(registry.entries.value).toEqual([entry2])

        unregister2.unsubscribe()
        expect(registry.entries.value).toEqual([])
    })

    describe('contributions observable', () => {
        test('emits stream of results of registrations', () => {
            const registry = new class extends ContributionRegistry {
                public getContributionsFromEntries(
                    entries: Observable<ContributionsEntry[]>
                ): Observable<EvaluatedContributions> {
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

        test('supports registration of an observable', () => {
            const registry = new class extends ContributionRegistry {
                public getContributionsFromEntries(
                    entries: Observable<ContributionsEntry[]>
                ): Observable<EvaluatedContributions> {
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

        test('emits when context changes and filters on context', () => {
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
                    ): Observable<EvaluatedContributions> {
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

        test('continues after error thrown during evaluation', () => {
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
                    ): Observable<EvaluatedContributions> {
                        return super.getContributionsFromEntries(entries, scope, undefined, () => void 0 /* noop log */)
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
    test('handles an empty array', () => expect(mergeContributions([])).toEqual({}))
    test('handles a single item', () =>
        expect(mergeContributions([FIXTURE_CONTRIBUTIONS_1 as EvaluatedContributions])).toEqual(
            FIXTURE_CONTRIBUTIONS_1
        ))
    test('handles multiple items', () =>
        expect(
            mergeContributions([
                FIXTURE_CONTRIBUTIONS_1 as EvaluatedContributions,
                FIXTURE_CONTRIBUTIONS_2 as EvaluatedContributions,
                {},
                { actions: [] },
                { menus: {} },
            ])
        ).toEqual(FIXTURE_CONTRIBUTIONS_MERGED))
})

const FIXTURE_CONTEXT = new Map<string, any>(
    Object.entries({
        a: true,
        b: false,
    })
)

describe('contextFilter', () => {
    test('filters', () =>
        expect(
            contextFilter(
                FIXTURE_CONTEXT,
                [{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a' }, { x: 4, when: 'b' }],
                x => x === 'a'
            )
        ).toEqual([{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a' }]))
})

describe('filterContributions', () => {
    test('handles empty contributions', () =>
        expect(filterContributions(EMPTY_COMPUTED_CONTEXT, {})).toEqual({} as Contributions))

    test('handles empty menu contributions', () =>
        expect(filterContributions(EMPTY_COMPUTED_CONTEXT, { menus: {} })).toEqual({
            menus: {},
        } as Contributions))

    test('handles empty array of menu contributions', () =>
        expect(filterContributions(EMPTY_COMPUTED_CONTEXT, { menus: { commandPalette: [] } })).toEqual({
            menus: { commandPalette: [] },
        } as Contributions))

    test('handles non-empty contributions', () =>
        expect(
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
            )
        ).toEqual({
            actions: [{ id: 'a1', command: 'c' }, { id: 'a2', command: 'c' }, { id: 'a3', command: 'c' }],
            menus: {
                [ContributableMenu.CommandPalette]: [{ action: 'a1', when: 'a' }, { action: 'a3' }],
                [ContributableMenu.GlobalNav]: [{ action: 'a1', when: 'a' }, { action: 'a2' }],
            },
        } as Contributions))

    test('throws an error if an error occurs during evaluation', () => {
        const input: Contributions = {
            actions: [{ id: 'a', command: 'c', title: 'a' }],
            menus: { commandPalette: [{ action: 'a', when: 'a' }] },
        }
        expect(() =>
            filterContributions(FIXTURE_CONTEXT, input, () => {
                throw new Error('')
            })
        ).toThrow()
    })
})

// tslint:disable:no-invalid-template-strings
describe('evaluateContributions', () => {
    const TEST_TEMPLATE_EVALUATOR = {
        evaluateTemplate: () => 'x',
        needsEvaluation: (template: string) => template === 'a',
    }

    test('handles empty contributions', () =>
        expect(evaluateContributions(EMPTY_COMPUTED_CONTEXT, {})).toEqual({} as Contributions))

    test('handles empty array of command contributions', () =>
        expect(evaluateContributions(EMPTY_COMPUTED_CONTEXT, { actions: [] })).toEqual({
            actions: [],
        } as Contributions))

    test('handles non-empty contributions', () => {
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
        expect(evaluateContributions(FIXTURE_CONTEXT, input, TEST_TEMPLATE_EVALUATOR)).toEqual({
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
        expect(input).toEqual(origInput)
    })

    test('supports commandArguments with the first element non-evaluated', () => {
        expect(
            evaluateContributions(
                FIXTURE_CONTEXT,
                {
                    actions: [{ id: 'a', command: 'c', commandArguments: ['b', 'a', 'b', 'a'] }],
                },
                TEST_TEMPLATE_EVALUATOR
            )
        ).toEqual({
            actions: [{ id: 'a', command: 'c', commandArguments: ['b', 'x', 'b', 'x'] }],
        } as Contributions)
    })

    const TEST_THROW_EVALUATOR = {
        evaluateTemplate: () => {
            throw new Error('')
        },
        needsEvaluation: () => true,
    }

    test('does not evaluate the `id` or `command` values', () => {
        expect(
            evaluateContributions(
                FIXTURE_CONTEXT,
                {
                    actions: [{ id: 'a', command: 'c' }],
                },
                TEST_THROW_EVALUATOR
            )
        ).toEqual({
            actions: [{ id: 'a', command: 'c' }],
        } as Contributions)
    })

    test('throws an error if an error occurs during evaluation', () => {
        const input: Contributions = { actions: [{ id: 'a', command: 'c', title: 'a' }] }
        expect(() => evaluateContributions(FIXTURE_CONTEXT, input, TEST_THROW_EVALUATOR)).toThrow()
    })

    test('evaluates `actionItem.pressed` if present', () => {
        const input: Contributions = { actions: [{ id: 'a', command: 'c', title: 'a', actionItem: { pressed: 'a' } }] }
        expect(evaluateContributions(FIXTURE_CONTEXT, input)).toEqual({
            actions: [{ id: 'a', command: 'c', title: 'a', actionItem: { pressed: true } }],
        })
    })
})
// tslint:enable:no-invalid-template-strings
