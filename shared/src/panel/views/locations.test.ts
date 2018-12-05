import assert from 'assert'
import { GroupedLocations, groupLocations } from './locations'

type TestLocation = string

type TestGroup = string

const LOCATIONS: TestLocation[] = ['a/a/0', 'b/a/0', 'a/a/1', 'a/b/0']

const GROUP_KEYS: ((location: TestLocation) => TestGroup | undefined)[] = [
    loc => loc.split('/')[0],
    loc => loc.split('/')[1],
]

describe('groupLocations', () => {
    it('groups 1 levels', () =>
        assert.deepStrictEqual(
            groupLocations<TestLocation, TestGroup>(LOCATIONS, null, GROUP_KEYS.slice(0, 1), LOCATIONS[0]),
            {
                groups: [[{ key: 'a', count: 3 }, { key: 'b', count: 1 }]],
                selectedGroups: ['a'],
                visibleLocations: ['a/a/0', 'a/a/1', 'a/b/0'],
            } as GroupedLocations<TestLocation, TestGroup>
        ))

    it('groups 2 levels', () =>
        assert.deepStrictEqual(groupLocations<TestLocation, TestGroup>(LOCATIONS, null, GROUP_KEYS, LOCATIONS[0]), {
            groups: [
                [{ key: 'a', count: 3 }, { key: 'b', count: 1 }],
                [{ key: 'a', count: 2 }, { key: 'b', count: 1 }],
            ],
            selectedGroups: ['a', 'a'],
            visibleLocations: ['a/a/0', 'a/a/1'],
        } as GroupedLocations<TestLocation, TestGroup>))

    it('supports initial selectedGroups', () =>
        assert.deepStrictEqual(
            groupLocations<TestLocation, TestGroup>(LOCATIONS, ['b', 'a'], GROUP_KEYS, LOCATIONS[0]),
            {
                groups: [[{ key: 'a', count: 3 }, { key: 'b', count: 1 }], [{ key: 'a', count: 1 }]],
                selectedGroups: ['b', 'a'],
                visibleLocations: ['b/a/0'],
            } as GroupedLocations<TestLocation, TestGroup>
        ))

    it('handles selectedGroups element that does not exist', () =>
        assert.deepStrictEqual(
            groupLocations<TestLocation, TestGroup>(LOCATIONS, ['b', 'x'], GROUP_KEYS, LOCATIONS[0]),
            {
                groups: [[{ key: 'a', count: 3 }, { key: 'b', count: 1 }], [{ key: 'a', count: 1 }]],
                selectedGroups: ['b', 'x'],
                visibleLocations: [],
            } as GroupedLocations<TestLocation, TestGroup>
        ))

    it('resolves selectedGroups undefined element', () =>
        assert.deepStrictEqual(groupLocations<TestLocation, TestGroup>(LOCATIONS, ['a'], GROUP_KEYS, LOCATIONS[0]), {
            groups: [
                [{ key: 'a', count: 3 }, { key: 'b', count: 1 }],
                [{ key: 'a', count: 2 }, { key: 'b', count: 1 }],
            ],
            selectedGroups: ['a', 'a'],
            visibleLocations: ['a/a/0', 'a/a/1'],
        } as GroupedLocations<TestLocation, TestGroup>))
})
