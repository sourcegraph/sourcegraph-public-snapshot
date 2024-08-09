import { expect, test } from '$testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/branches`

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
        BranchesPage_OverviewQuery: () => ({
            repository: {
                defaultBranch: {
                    id: '1',
                    displayName: 'main',
                },
                branches: {
                    nodes: [{ displayName: 'main', id: '1' }, { displayName: 'feature/branch' }],
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
        BranchesPage_OverviewQuery: () => {
            throw new Error('Test error')
        },
    })
    await page.goto(url)
    await expect(page.getByText(/Test error/)).toBeVisible()
})
