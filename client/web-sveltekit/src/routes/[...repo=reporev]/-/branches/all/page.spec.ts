import { expect, test } from '$testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/branches/all`

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
        AllBranchesPage_BranchesQuery: () => ({
            repository: {
                branches: {
                    nodes: [{ displayName: 'main' }, { displayName: 'feature/branch' }],
                    pageInfo: {
                        // Needed to prevent infinity scroll from trying to load more pages
                        hasNextPage: false,
                    },
                },
            },
        }),
    })
})

test('list branches', async ({ page }) => {
    await page.goto(url)

    await expect(page.getByRole('link', { name: 'main' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'feature/branch' })).toBeVisible()
})

test('error loading branches', async ({ page, sg }) => {
    sg.mockOperations({
        AllBranchesPage_BranchesQuery: () => {
            throw new Error('Test error')
        },
    })
    await page.goto(url)
    await expect(page.getByText(/Test error/)).toBeVisible()
})
