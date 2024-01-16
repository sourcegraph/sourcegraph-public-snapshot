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
        CommitsQuery: ({ first, afterCursor }) => {
            const from = afterCursor ? +afterCursor : 0
            const to = from ? (first ?? 20) - 5 : first ?? 20
            return {
                node: {
                    __typename: 'Repository',
                    id: '1',
                    commit: {
                        id: '1',
                        ancestors_paginated: {
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

test('forward pagination', async ({ page }) => {
    await page.goto(`/${repoName}/-/commits`)
    // First page of commits is loaded
    await expect(page.getByRole('link', { name: 'Commit 0' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Commit 19' })).toBeVisible()

    // Next page of commits is loaded
    await page.getByRole('link', { name: 'Next' }).click()

    await expect(page.getByRole('link', { name: 'Commit 20' })).toBeVisible()
})

test('backward pagination', async ({ page }) => {
    await page.goto(`/${repoName}/-/commits?$after=20`)

    // Second page of commits is loaded
    await expect(page.getByRole('link', { name: 'Commit 20' })).toBeVisible()

    // First page of commits is loaded
    await page.getByRole('link', { name: 'Previous' }).click()

    await expect(page.getByRole('link', { name: 'Commit 0' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Commit 19' })).toBeVisible()
})
