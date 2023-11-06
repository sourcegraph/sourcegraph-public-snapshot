import { describe, expect, test } from '@jest/globals'

import { collectMetrics } from './metrics'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

describe('collectMetrics', () => {
    test('operators', () => {
        expect(collectMetrics('not e and a or b and c not d')).toMatchInlineSnapshot(`
            {
              "count_or": 1,
              "count_and": 2,
              "count_not": 2
            }
        `)
    })

    test('predicates', () => {
        expect(collectMetrics('repo:contains.path(foo) r:contains.file(path:foo content:bar)')).toMatchInlineSnapshot(`
            {
              "count_repo_contains_path": 1,
              "count_repo_contains_file": 1
            }
        `)
    })

    test('only patterns', () => {
        expect(collectMetrics('i want google search context:global')).toMatchInlineSnapshot(`
            {
              "count_only_patterns": 1,
              "count_only_patterns_three_or_more": 1
            }
        `)
    })

    test('personal context', () => {
        expect(collectMetrics('i want google search context:personal')).toMatchInlineSnapshot(`
            {
              "count_non_global_context": 1
            }
        `)
    })

    test('exhaustive count:all', () => {
        expect(collectMetrics('I SAID EVERYTHING count:all')).toMatchInlineSnapshot(`
            {
              "count_count_all": 1
            }
        `)
    })

    test('select', () => {
        expect(collectMetrics('select:commit.diff.added type:diff fixme')).toMatchInlineSnapshot(`
            {
              "count_select_commit_diff_added": 1
            }
        `)
    })

    test('uninteresting', () => {
        expect(collectMetrics('repo:foo file:bar whatever')).toMatchInlineSnapshot('{}')
    })
})
