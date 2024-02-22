import { expect, test } from '../../../../../../testing/integration'

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
