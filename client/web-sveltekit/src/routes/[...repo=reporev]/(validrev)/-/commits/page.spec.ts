import type { GitCommitMock } from '$testing/graphql-type-mocks'

import { expect, test } from '../../../../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'

test.beforeEach(async ({ sg }) => {
    sg.mockOperations({
        ResolveRepoRevison: () => ({
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
                node: {
                    __typename: 'Repository',
                    id: '1',
                    commit: {
                        id: '1',
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

test('infinity scroll', async ({ page }) => {
    await page.goto(`/${repoName}/-/commits`)
    // First page of commits is loaded
    const firstCommit = page.getByRole('link', { name: 'Commit 0' })
    await expect(firstCommit).toBeVisible()
    await expect(page.getByRole('link', { name: 'Commit 19' })).toBeVisible()

    // Position mouse over list of commits so that whell events will scroll
    // the list
    const { x, y } = (await firstCommit.boundingBox()) ?? { x: 0, y: 0 }
    await page.mouse.move(x, y)

    // Scroll list, which should load next page
    await page.mouse.wheel(0, 1000)
    await expect(page.getByRole('link', { name: 'Commit 20' })).toBeVisible()

    // Refreshing should restore commit list and scroll position
    await page.reload()
    await expect(page.getByRole('link', { name: 'Commit 20' })).toBeInViewport()
})
