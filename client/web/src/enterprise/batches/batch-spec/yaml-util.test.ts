import { describe, expect, it } from 'vitest'

import {
    excludeRepo,
    haveMatchingWorkspaces,
    insertFieldIntoLibraryItem,
    insertQueryIntoLibraryItem,
    quoteYAMLString,
} from './yaml-util'

const SPEC_WITH_ONE_REPOSITORY = `name: hello-world
on:
    - repository: repo1
`

const SPEC_WITH_ONE_REPOSITORY_AND_STEPS = `name: hello-world
on:
    - repository: repo1
steps:
    - run: echo Hello World | tee -a $(find -name README.md)
      container: ubuntu:18.04
`

const SPEC_WITH_QUERY = `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
`

const SPEC_WITH_QUERY_AND_STEPS = `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
steps:
    - run: echo Hello World | tee -a $(find -name README.md)
      container: ubuntu:18.04
`

const SPEC_WITH_BOTH = `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
    - repository: repo2
    - repository: repo3
`

const SPEC_WITH_BOTH_AND_STEPS = `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
    - repository: repo2
    - repository: repo3
steps:
    - run: echo Hello World | tee -a $(find -name README.md)
      container: ubuntu:18.04
`

const SPEC_WITH_IMPORT = `name: hello-world
on:
    - repository: repo1
importChangesets:
    - repository: repo2
      externalIDs:
        - 123
`

const SPEC_WITH_IMPORT_AND_STEPS = `name: hello-world
on:
    - repository: repo1
steps:
    - run: echo Hello World | tee -a $(find -name README.md)
      container: ubuntu:18.04
importChangesets:
    - repository: repo2
      externalIDs:
        - 123
`

const SAMPLE_SPECS: { original: string; expected: string | 0; repo: string; branch: string }[] = [
    // Spec with only one "repository" directive, repo to remove doesn't match => no change
    {
        repo: 'no-match',
        branch: 'doesnt-matter',
        original: SPEC_WITH_ONE_REPOSITORY,
        expected: 0,
    },

    // Spec with multiple "repository" directives, repo to remove doesn't match => no change
    {
        repo: 'no-match',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repository: repo1
    - repository: repo2
    - repository: repo3
    - repository: repo4
      branch: doesnt-matter
`,
        expected: 0,
    },

    // Spec with multiple "repository" directives, repo to remove matches => remove it
    {
        repo: 'repo1',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repository: repo1
    - repository: repo2
    - repository: repo3
      branch: doesnt-matter
`,
        expected: `name: hello-world
on:
    - repository: repo2
    - repository: repo3
      branch: doesnt-matter
`,
    },

    // Spec with multiple "repository" directives, repo to remove matches case insensitive => remove it
    {
        repo: 'repo1',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repository: REPO1
    - repository: repo2
`,
        expected: `name: hello-world
on:
    - repository: repo2
`,
    },

    // Spec with multiple "repository" directives + branches, repo + branch to remove
    // matches => remove it
    {
        repo: 'repo1',
        branch: 'branch2',
        original: `name: hello-world
on:
    - repository: repo1
      branch: branch1
    - repository: repo1
      branch: branch2
    - repository: repo1
      branch: branch3
    - repository: repo2
    - repository: repo3
      branch: doesnt-matter
`,
        expected: `name: hello-world
on:
    - repository: repo1
      branch: branch1
    - repository: repo1
      branch: branch3
    - repository: repo2
    - repository: repo3
      branch: doesnt-matter
`,
    },

    // Spec with multiple "repository" directives + branches, repo + branch to remove
    // matches case insensitive => remove it
    {
        repo: 'repo1',
        branch: 'branch2',
        original: `name: hello-world
on:
    - repository: REPO1
      branch: BRANCH1
    - repository: REPO1
      branch: BRANCH2
`,
        expected: `name: hello-world
on:
    - repository: REPO1
      branch: BRANCH1
`,
    },

    // Spec with multiple "repository" directives + branches, repo + branch to remove
    // doesn't match => no change
    {
        repo: 'repo1',
        branch: 'no-match',
        original: `name: hello-world
on:
    - repository: repo1
      branch: branch1
    - repository: repo1
      branch: branch2
    - repository: repo1
      branch: branch3
    - repository: repo2
      branch: no-match
`,
        expected: 0,
    },

    // Spec with "repositoriesMatchingQuery" => append "-repo:" with escaped dot
    {
        repo: 'github.com/repo1',
        branch: 'doesnt-matter',
        original: SPEC_WITH_QUERY,
        expected: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md -repo:github\\.com/repo1
`,
    },

    // Spec with "repositoriesMatchingQuery" with the query captured in quotes => append
    // "-repo:" without any escaping
    {
        repo: 'github.com/repo1',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repositoriesMatchingQuery: "file:README.md"
        `,
        expected: `name: hello-world
on:
    - repositoriesMatchingQuery: "file:README.md -repo:github.com/repo1"
        `,
    },

    // Spec with "repositoriesMatchingQuery" and multiple "repository" directives but repo
    // to remove doesn't match => just append "-repo:"
    {
        repo: 'repo1',
        branch: 'doesnt-matter',
        original: SPEC_WITH_BOTH,
        expected: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md -repo:repo1
    - repository: repo2
    - repository: repo3
`,
    },

    // Spec with "repositoriesMatchingQuery" and multiple "repository" directives and repo
    // to remove matches => append "-repo:" and remove directive
    {
        repo: 'repo1',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
    - repository: repo0
    - repository: repo1
    - repository: repo2
`,
        expected: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md -repo:repo1
    - repository: repo0
    - repository: repo2
`,
    },
]

const SAMPLE_COMPARISON_SPECS: { spec1: string; spec2: string; matches: boolean | 'UNKNOWN' }[] = [
    // Not parseable => UNKNOWN
    { spec1: 'wut', spec2: 'huh', matches: 'UNKNOWN' },
    { spec1: SPEC_WITH_BOTH, spec2: 'huh', matches: 'UNKNOWN' },
    { spec1: 'wut', spec2: SPEC_WITH_BOTH, matches: 'UNKNOWN' },
    // Identical specs => matches
    { spec1: SPEC_WITH_ONE_REPOSITORY, spec2: SPEC_WITH_ONE_REPOSITORY, matches: true },
    { spec1: SPEC_WITH_QUERY, spec2: SPEC_WITH_QUERY, matches: true },
    { spec1: SPEC_WITH_BOTH, spec2: SPEC_WITH_BOTH, matches: true },
    { spec1: SPEC_WITH_IMPORT, spec2: SPEC_WITH_IMPORT, matches: true },
    // Different directives => no match
    { spec1: SPEC_WITH_ONE_REPOSITORY, spec2: SPEC_WITH_QUERY, matches: false },
    { spec1: SPEC_WITH_ONE_REPOSITORY, spec2: SPEC_WITH_BOTH, matches: false },
    { spec1: SPEC_WITH_QUERY, spec2: SPEC_WITH_BOTH, matches: false },
    // Added/removed import => no match
    { spec1: SPEC_WITH_ONE_REPOSITORY, spec2: SPEC_WITH_IMPORT, matches: false },
    // Only different steps => matches
    { spec1: SPEC_WITH_ONE_REPOSITORY, spec2: SPEC_WITH_ONE_REPOSITORY_AND_STEPS, matches: true },
    { spec1: SPEC_WITH_QUERY, spec2: SPEC_WITH_QUERY_AND_STEPS, matches: true },
    { spec1: SPEC_WITH_BOTH, spec2: SPEC_WITH_BOTH_AND_STEPS, matches: true },
    { spec1: SPEC_WITH_IMPORT, spec2: SPEC_WITH_IMPORT_AND_STEPS, matches: true },
]

describe('Batch spec yaml utils', () => {
    describe('excludeRepo', () => {
        it('should succeed and exclude the repo from the spec if it can', () => {
            for (const { original, expected, repo, branch } of SAMPLE_SPECS) {
                expect(excludeRepo(original, repo, branch)).toEqual({
                    success: true,
                    spec: expected === 0 ? original : expected,
                })
            }
        })

        it('should fail and return an error if it cannot parse the spec', () => {
            expect(excludeRepo('invalid', 'repo1', 'doesnt-matter')).toEqual({
                success: false,
                error: 'Spec not parseable',
                spec: 'invalid',
            })
        })
    })

    describe('haveMatchingWorkspaces', () => {
        it('should return the correct comparison result for a pair of batch specs', () => {
            for (const { spec1, spec2, matches } of SAMPLE_COMPARISON_SPECS) {
                // Order shouldn't matter
                expect(haveMatchingWorkspaces(spec1, spec2)).toEqual(matches)
                expect(haveMatchingWorkspaces(spec2, spec1)).toEqual(matches)
            }
        })
    })

    describe('insertFieldIntoLibraryItem', () => {
        it('should correctly overwrite the name in a given spec', () => {
            for (const spec of [SPEC_WITH_ONE_REPOSITORY, SPEC_WITH_IMPORT_AND_STEPS]) {
                expect(insertFieldIntoLibraryItem(spec, 'new-name', 'name', true)).toEqual(
                    spec.replace('hello-world', 'new-name')
                )
            }
        })
        it('should correctly quote special names', () => {
            for (const newName of ['bad: colons', 'true', 'false', '1.23']) {
                expect(insertFieldIntoLibraryItem(SPEC_WITH_ONE_REPOSITORY, newName, 'name', true)).toEqual(
                    SPEC_WITH_ONE_REPOSITORY.replace('hello-world', `"${newName}"`)
                )
            }
        })
        it('should not quote edge-cases that do not need quoting', () => {
            for (const newName of ['"asdf"', "'asdf'", 'hello-"asdf"', 'zero', 'on', 'off', 'yes', 'no']) {
                expect(insertFieldIntoLibraryItem(SPEC_WITH_ONE_REPOSITORY, newName, 'name', true)).toEqual(
                    SPEC_WITH_ONE_REPOSITORY.replace('hello-world', newName)
                )
            }
        })
        it('should correctly overwrite the `on` field in a given spec', () => {
            for (const spec of [SPEC_WITH_ONE_REPOSITORY, SPEC_WITH_IMPORT_AND_STEPS]) {
                expect(insertFieldIntoLibraryItem(spec, '- repository: foo-bar\n', 'on', false)).toEqual(
                    spec.replace('repo1', 'foo-bar')
                )
            }
        })
        it('should not quote value when quotable flag is false', () => {
            for (const newName of ['bad: colons', 'true', 'false', '1.23']) {
                expect(insertFieldIntoLibraryItem(SPEC_WITH_ONE_REPOSITORY, newName, 'name', false)).toEqual(
                    SPEC_WITH_ONE_REPOSITORY.replace('hello-world', newName)
                )
            }
        })
    })

    describe('insertQueryIntoLibraryItem', () => {
        it('should add simple query', () => {
            const spec = insertQueryIntoLibraryItem(SPEC_WITH_QUERY, 'context:global hello patternType:standard', false)
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: context:global hello patternType:standard\n'
            )
        })

        it('should add quoted query', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global "hello" patternType:standard',
                false
            )
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: context:global "hello" patternType:standard\n'
            )
        })

        it('should add unbalanced quoted query', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global "hello patternType:standard',
                false
            )
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: context:global "hello patternType:standard\n'
            )
        })

        it('should add query with colon', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global hello: world patternType:standard',
                false
            )
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: "context:global hello: world patternType:standard"\n'
            )
        })

        it('should add query with colon and quotes', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global hello: "world" patternType:standard',
                false
            )
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: "context:global hello: \\"world\\" patternType:standard"\n'
            )
        })

        it('should add query with colon and unbalanced quotes', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global hello: "world patternType:standard',
                false
            )
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: "context:global hello: \\"world patternType:standard"\n'
            )
        })

        it('should add query with colon and double unbalanced quotes', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global "hello": "world patternType:standard',
                false
            )
            expect(spec).toEqual(
                'name: hello-world\n' +
                    'on:\n' +
                    '    - repositoriesMatchingQuery: "context:global \\"hello\\": \\"world patternType:standard"\n'
            )
        })

        it('should comment out existing query when commentExistingQuery is true', () => {
            const spec = insertQueryIntoLibraryItem(
                SPEC_WITH_QUERY,
                'context:global hello: "world" patternType:standard',
                true
            )
            expect(spec).toEqual(`name: hello-world
on:
    - repositoriesMatchingQuery: "context:global hello: \\"world\\" patternType:standard"
# - repositoriesMatchingQuery: file:README.md
`)
        })
    })

    describe('quoteYAMLString', () => {
        it('should add double quote a numeric value', () => {
            const quotedString = quoteYAMLString('1024')
            expect(quotedString).toEqual('"1024"')
        })

        it('should not quote a string without special characters', () => {
            const unQuotedString = quoteYAMLString('random-name')
            expect(unQuotedString).toEqual('random-name')
        })

        it('should double quote and escape special characters if contained in the value', () => {
            const quotedString = quoteYAMLString(
                String.raw`fork:yes repo:^github\.com/foo/bar$ file:package.json "scaling-palm-tree": "..."`
            )
            expect(quotedString).toEqual(
                String.raw`"fork:yes repo:^github\\.com/foo/bar$ file:package.json \"scaling-palm-tree\": \"...\""`
            )
        })
    })
})
