import type { GitCommitMock } from '$testing/graphql-type-mocks'
import { expect, test } from '$testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/commits`

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
        CommitsPage_CommitsQuery: ({ first, afterCursor }) => {
            const from = afterCursor ? +afterCursor : 0
            const to = from ? (first ?? 20) - 5 : first ?? 20
            return {
                repository: {
                    commit: {
                        ancestors: {
                            nodes: Array.from(
                                { length: to },
                                (_, index): GitCommitMock => ({
                                    id: `commit ${from + index}`,
                                    subject: `Commit ${from + index}`,
                                    body: `Commit ${from + index} body`,
                                })
                            ),
                            pageInfo: {
                                endCursor: !afterCursor ? String(to) : null,
                                hasNextPage: !afterCursor,
                            },
                        },
                    },
                },
            }
        },
    })
})

test.fixme('infinity scroll', async ({ page, utils }) => {
    await page.goto(url)
    // First page of commits is loaded
    const firstCommit = page.getByRole('link', { name: 'Commit 0' })
    await expect(firstCommit).toBeVisible()
    await expect(page.getByRole('link', { name: 'Commit 19' })).toBeVisible()

    // Scroll list, which should load next page
    await utils.scrollYAt(firstCommit, 1000)
    await expect(page.getByRole('link', { name: 'Commit 20' })).toBeVisible()

    // Refreshing should restore commit list and scroll position
    await page.reload()
    await expect(page.getByRole('link', { name: 'Commit 20' })).toBeInViewport()
})

test('no commits', async ({ sg, page }) => {
    sg.mockOperations({
        CommitsPage_CommitsQuery: () => ({
            repository: {
                commit: {
                    ancestors: {
                        nodes: [],
                        pageInfo: {
                            endCursor: null,
                            hasNextPage: false,
                        },
                    },
                },
            },
        }),
    })

    await page.goto(url)
    await expect(page.getByText('No commits found')).toBeVisible()
})

test('error', async ({ sg, page, utils }) => {
    await page.goto(url)

    const firstCommit = page.getByRole('link', { name: 'Commit 0' })
    await expect(firstCommit).toBeVisible()

    sg.mockOperations({
        CommitsPage_CommitsQuery: () => {
            throw new Error('Test error')
        },
    })
    // Scroll list, which should trigger an error
    await utils.scrollYAt(firstCommit, 2000)
    await expect(page.getByText('Test error')).toBeVisible()
})
