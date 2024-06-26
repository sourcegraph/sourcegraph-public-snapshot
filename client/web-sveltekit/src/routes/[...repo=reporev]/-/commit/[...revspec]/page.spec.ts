import { expect, test } from '../../../../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/commit/1234567890abcdef`

test.beforeEach(async ({ sg }) => {
    sg.mockOperations({
        ResolveRepoRevision: () => ({
            repositoryRedirect: {
                __typename: 'Repository',
                mirrorInfo: {
                    cloned: true,
                    cloneInProgress: false,
                },
            },
        }),
    })
})

test('commit not found', async ({ page, sg }) => {
    if (process.env.BAZEL_SKIP_TESTS?.includes('commit not found')) {
        // Some tests are working with `pnpm run test` but not in Bazel.
        // To get CI working we are skipping these tests for now.
        // https://github.com/sourcegraph/sourcegraph/pull/62560#issuecomment-2128313393
        return
    }
    sg.mockOperations({
        ResolveRepoRevision: () => ({
            repositoryRedirect: {
                mirrorInfo: {
                    cloned: true,
                    cloneInProgress: false,
                },
            },
        }),
        CommitPage_CommitQuery: () => ({
            repository: {
                commit: null,
            },
        }),
    })
    await page.goto(url)
    await expect(page.getByText(/Commit not found/)).toBeVisible()
})

test('error loading commit information', async ({ page, sg }) => {
    if (process.env.BAZEL_SKIP_TESTS?.includes('error loading commit information')) {
        // Some tests are working with `pnpm run test` but not in Bazel.
        // To get CI working we are skipping these tests for now.
        // https://github.com/sourcegraph/sourcegraph/pull/62560#issuecomment-2128313393
        return
    }
    sg.mockOperations({
        CommitPage_CommitQuery: () => {
            throw new Error('Test error')
        },
    })
    await page.goto(url)
    await expect(page.getByText(/Test error/)).toBeVisible()
})

test('error loading diff information', async ({ page, sg }) => {
    sg.mockOperations({
        CommitPage_DiffQuery: () => {
            throw new Error('Test error')
        },
    })
    await page.goto(url)
    await expect(page.getByText(/Test error/)).toBeVisible()
})

test('shows previous diffs when error occurs', async ({ page, sg }) => {
    let callCount = 0
    sg.mockOperations({
        CommitPage_DiffQuery: () => {
            if (callCount === 1) {
                throw new Error('Test error')
            }
            callCount++
            return {
                repository: {
                    comparison: {
                        fileDiffs: {
                            nodes: [
                                {
                                    __typename: 'FileDiff',
                                    oldPath: null,
                                    newPath: '<new path>',
                                },
                            ],
                            pageInfo: {
                                hasNextPage: true,
                                endCursor: 'cursor',
                            },
                        },
                    },
                },
            }
        },
    })
    await page.goto(url)
    await expect(page.getByText('<new path>')).toBeVisible()
    await expect(page.getByText('Test error')).toBeVisible()
})
