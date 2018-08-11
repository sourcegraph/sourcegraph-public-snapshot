import * as assert from 'assert'
import { Observable } from 'rxjs'
import { Subscription } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ContributableMenu, Contributions } from '../../protocol'
import { ContributionRegistry, ContributionsEntry, mergeContributions } from './contribution'

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
    it('is initially empty', () => {
        assert.deepStrictEqual(new ContributionRegistry().entries.value, [])
    })

    it('registers and unregisters contributions', () => {
        const subscriptions = new Subscription()
        const registry = new ContributionRegistry()
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
        const registry = new ContributionRegistry()
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
        it('returns stream of results of registrations', () => {
            const registry = new class extends ContributionRegistry {
                public getContributions(entries: Observable<ContributionsEntry[]>): Observable<Contributions> {
                    return super.getContributions(entries)
                }
            }()
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
