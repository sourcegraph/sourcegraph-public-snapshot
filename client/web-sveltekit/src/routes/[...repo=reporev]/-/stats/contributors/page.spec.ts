import { test, expect } from '$testing/integration'

const repoName = 'sourcegraph/sourcegraph'
const url = `/${repoName}/-/stats/contributors`

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
        ContributorsPage_ContributorsQuery: ({ after, before }) => {
            const allNodes = Array.from({ length: 15 }, (_, index) => ({
                _id: (index + 1).toString(),
                person: { displayName: `Person ${index + 1}` },
            }))

            let startCursor = '1'
            let endCursor = '5'
            let nodes: any[] = allNodes.slice(0, 5)

            if (after) {
                const index = allNodes.findIndex(node => node._id === after)
                startCursor = allNodes[index + 1]._id
                endCursor = allNodes[index + 5]._id
                nodes = allNodes.slice(index + 1, index + 6)
            } else if (before) {
                const index = allNodes.findIndex(node => node._id === before)
                startCursor = allNodes[index - 1]._id
                endCursor = allNodes[index - 5]._id
                nodes = allNodes.slice(index - 5, index)
            }

            const pageInfo = {
                startCursor,
                endCursor,
                hasNexPage: endCursor !== allNodes[nodes.length - 1]._id,
                hasPreviousPage: startCursor !== allNodes[0]._id,
            }

            return {
                repository: {
                    contributors: {
                        nodes,
                        pageInfo,
                        totalCount: allNodes.length,
                    },
                },
            }
        },
    })
})

// Disabled because flaky in CI
test.fixme('paginate contributors', async ({ page }) => {
    await page.goto(url)
    await expect(page.getByRole('row')).toHaveCount(5)

    // Go to next page
    await page.getByRole('link', { name: 'Next' }).click()
    await expect(page.getByText('Person 6')).toBeVisible()

    // Go to next page
    await page.getByRole('link', { name: 'Next' }).click()
    await expect(page.getByText('Person 11')).toBeVisible()

    // Go to previous page
    await page.getByRole('link', { name: 'Previous' }).click()
    await expect(page.getByText('Person 6')).toBeVisible()
})

test('no contributors', async ({ sg, page }) => {
    sg.mockOperations({
        ContributorsPage_ContributorsQuery: () => ({
            repository: {
                contributors: {
                    nodes: [],
                    totalCount: 0,
                },
            },
        }),
    })

    await page.goto(url)
    await expect(page.getByText('No contributors found')).toBeVisible()
})

test('error', async ({ sg, page }) => {
    sg.mockOperations({
        ContributorsPage_ContributorsQuery: () => {
            throw new Error('Test error')
        },
    })

    await page.goto(url)
    await expect(page.getByText(/Test error/)).toBeVisible()
})
