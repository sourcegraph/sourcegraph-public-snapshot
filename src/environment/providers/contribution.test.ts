import * as assert from 'assert'
import { Observable, of } from 'rxjs'
import { Subscription } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ContributableMenu, Contributions } from '../../protocol'
import { Context, EMPTY_CONTEXT } from '../context/context'
import { EMPTY_OBSERVABLE_ENVIRONMENT } from '../environment'
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
    commands: [{ command: 'c1.a', title: 'C1.A' }, { command: 'c1.b', title: 'C1.B' }],
    menus: {
        [ContributableMenu.CommandPalette]: [{ command: 'c1.a' }],
        [ContributableMenu.GlobalNav]: [{ command: 'c1.a' }, { command: 'c1.b' }],
    },
}

const FIXTURE_CONTRIBUTIONS_2: Contributions = {
    commands: [{ command: 'c2.a', title: 'C2.A' }, { command: 'c2.b', title: 'C2.B' }],
    menus: {
        [ContributableMenu.CommandPalette]: [{ command: 'c2.a' }],
        [ContributableMenu.EditorTitle]: [{ command: 'c2.a' }, { command: 'c2.b' }],
    },
}

const FIXTURE_CONTRIBUTIONS_MERGED: Contributions = {
    commands: [
        { command: 'c1.a', title: 'C1.A' },
        { command: 'c1.b', title: 'C1.B' },
        { command: 'c2.a', title: 'C2.A' },
        { command: 'c2.b', title: 'C2.B' },
    ],
    menus: {
        [ContributableMenu.CommandPalette]: [{ command: 'c1.a' }, { command: 'c2.a' }],
        [ContributableMenu.GlobalNav]: [{ command: 'c1.a' }, { command: 'c1.b' }],
        [ContributableMenu.EditorTitle]: [{ command: 'c2.a' }, { command: 'c2.b' }],
    },
}

describe('ContributionRegistry', () => {
    function create(context: Observable<Context> = EMPTY_OBSERVABLE_ENVIRONMENT.context): ContributionRegistry {
        return new ContributionRegistry(context)
    }

    it('is initially empty', () => {
        assert.deepStrictEqual(create().entries.value, [])
    })

    it('registers and unregisters contributions', () => {
        const subscriptions = new Subscription()
        const registry = create()
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
        const registry = create()
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
                public getContributions(entries: Observable<ContributionsEntry[]>): Observable<Contributions> {
                    return super.getContributions(entries)
                }
            }(EMPTY_OBSERVABLE_ENVIRONMENT.context)
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    registry.getContributions(
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

        it('emits when context changes and filters on context', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new class extends ContributionRegistry {
                    public constructor() {
                        super(
                            cold<Context>('-a-b-|', {
                                a: new Map([['x', 1], ['y', 2]]),
                                b: new Map([['x', 1], ['y', 1]]),
                            })
                        )
                    }

                    public getContributions(entries: Observable<ContributionsEntry[]>): Observable<Contributions> {
                        return super.getContributions(entries)
                    }
                }()
                expectObservable(
                    registry.getContributions(
                        of([
                            {
                                contributions: { menus: { commandPalette: [{ command: 'c', when: 'x == y' }] } },
                            },
                        ] as ContributionsEntry[])
                    )
                ).toBe('-a-b-|', {
                    a: { menus: { commandPalette: [] } },
                    b: { menus: { commandPalette: [{ command: 'c', when: 'x == y' }] } },
                })
            })
        })

        it('continues after error thrown during evaluation', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const registry = new class extends ContributionRegistry {
                    public constructor() {
                        super(cold<Context>('a', { a: new Map() }))
                    }

                    public getContributions(entries: Observable<ContributionsEntry[]>): Observable<Contributions> {
                        return super.getContributions(entries)
                    }
                }()
                expectObservable(
                    registry.getContributions(
                        of([
                            {
                                // Expression "!" will cause an error to be thrown.
                                contributions: { menus: { commandPalette: [{ command: 'c1', when: '!' }] } },
                            },
                            {
                                contributions: { menus: { commandPalette: [{ command: 'c2' }] } },
                            },
                        ] as ContributionsEntry[])
                    )
                ).toBe('a', {
                    a: { menus: { commandPalette: [{ command: 'c2' }] } },
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
            mergeContributions([FIXTURE_CONTRIBUTIONS_1, FIXTURE_CONTRIBUTIONS_2, {}, { commands: [] }, { menus: {} }]),
            FIXTURE_CONTRIBUTIONS_MERGED
        ))
})

const FIXTURE_CONTEXT = () =>
    new Map<string, any>(
        Object.entries({
            a: true,
            b: false,
        })
    )

describe('contextFilter', () => {
    it('filters', () =>
        assert.deepStrictEqual(
            contextFilter(
                FIXTURE_CONTEXT(),
                [{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a' }, { x: 4, when: 'b' }],
                x => x === 'a'
            ),
            [{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a' }]
        ))
})

describe('filterContributions', () => {
    it('handles empty contributions', () =>
        assert.deepStrictEqual(filterContributions(EMPTY_CONTEXT, {}), {} as Contributions))

    it('handles empty menu contributions', () =>
        assert.deepStrictEqual(filterContributions(EMPTY_CONTEXT, { menus: {} }), { menus: {} } as Contributions))

    it('handles empty array of menu contributions', () =>
        assert.deepStrictEqual(filterContributions(EMPTY_CONTEXT, { menus: { commandPalette: [] } }), {
            menus: { commandPalette: [] },
        } as Contributions))

    it('handles non-empty contributions', () =>
        assert.deepStrictEqual(
            filterContributions(
                FIXTURE_CONTEXT(),
                {
                    commands: [{ command: 'c1' }, { command: 'c2' }, { command: 'c3' }],
                    menus: {
                        [ContributableMenu.CommandPalette]: [
                            { command: 'c1', when: 'a' },
                            { command: 'c2', when: 'b' },
                            { command: 'c3' },
                        ],
                        [ContributableMenu.GlobalNav]: [
                            { command: 'c1', when: 'a' },
                            { command: 'c2' },
                            { command: 'c3', when: 'b' },
                        ],
                    },
                },
                x => x === 'a'
            ),
            {
                commands: [{ command: 'c1' }, { command: 'c2' }, { command: 'c3' }],
                menus: {
                    [ContributableMenu.CommandPalette]: [{ command: 'c1', when: 'a' }, { command: 'c3' }],
                    [ContributableMenu.GlobalNav]: [{ command: 'c1', when: 'a' }, { command: 'c2' }],
                },
            } as Contributions
        ))

    it('throws an error if an error occurs during evaluation', () => {
        const input: Contributions = {
            commands: [{ command: 'c', title: 'a' }],
            menus: { commandPalette: [{ command: 'c', when: 'a' }] },
        }
        assert.throws(() =>
            filterContributions(FIXTURE_CONTEXT(), input, () => {
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
        assert.deepStrictEqual(evaluateContributions(EMPTY_CONTEXT, {}), {} as Contributions))

    it('handles empty array of command contributions', () =>
        assert.deepStrictEqual(evaluateContributions(EMPTY_CONTEXT, { commands: [] }), {
            commands: [],
        } as Contributions))

    it('handles non-empty contributions', () => {
        const input: Contributions = {
            commands: [
                {
                    command: 'c1',
                    title: 'a',
                    category: 'a',
                    description: 'a',
                    iconURL: 'a',
                    actionItem: {
                        label: 'a',
                        description: 'a',
                        group: 'a',
                        iconDescription: 'a',
                        iconURL: 'a',
                    },
                },
                { command: 'c2', title: 'a', category: 'b' },
                { command: 'c3', title: 'b', category: 'b', actionItem: { label: 'a', description: 'b' } },
            ],
        }
        const origInput = JSON.parse(JSON.stringify(input))
        assert.deepStrictEqual(evaluateContributions(FIXTURE_CONTEXT(), input, TEST_TEMPLATE_EVALUATOR), {
            commands: [
                {
                    command: 'c1',
                    title: 'x',
                    category: 'x',
                    description: 'x',
                    iconURL: 'x',
                    actionItem: {
                        label: 'x',
                        description: 'x',
                        group: 'x',
                        iconDescription: 'x',
                        iconURL: 'x',
                    },
                },
                { command: 'c2', title: 'x', category: 'b' },
                { command: 'c3', title: 'b', category: 'b', actionItem: { label: 'x', description: 'b' } },
            ],
        } as Contributions)
        assert.deepStrictEqual(input, origInput, 'input must not be mutated')
    })

    const TEST_THROW_EVALUATOR = {
        evaluateTemplate: () => {
            throw new Error('')
        },
        needsEvaluation: () => true,
    }

    it('throws an error if an error occurs during evaluation', () => {
        const input: Contributions = { commands: [{ command: 'c', title: 'a' }] }
        assert.throws(() => evaluateContributions(FIXTURE_CONTEXT(), input, TEST_THROW_EVALUATOR))
    })
})
// tslint:enable:no-invalid-template-strings
