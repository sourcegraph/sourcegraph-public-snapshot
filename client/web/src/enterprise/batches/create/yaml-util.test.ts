import { excludeRepo } from './yaml-util'

const SAMPLE_SPECS: { original: string; expected: string | 0; repo: string; branch: string }[] = [
    // Spec with only one "repository" directive, repo to remove doesn't match => no change
    {
        repo: 'no-match',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repository: repo1
`,
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

    // Spec with "repositoriesMatchingQuery" => append "-repo:"
    {
        repo: 'repo1',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
`,
        expected: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md -repo:repo1
`,
    },

    // Spec with "repositoriesMatchingQuery" and multiple "repository" directives but repo
    // to remove doesn't match => just append "-repo:"
    {
        repo: 'repo1',
        branch: 'doesnt-matter',
        original: `name: hello-world
on:
    - repositoriesMatchingQuery: file:README.md
    - repository: repo2
    - repository: repo3
`,
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
})
