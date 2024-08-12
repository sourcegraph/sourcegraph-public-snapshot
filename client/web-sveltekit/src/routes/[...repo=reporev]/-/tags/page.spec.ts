import { test, expect } from '$testing/integration'

const repoName = 'sourcegraph/sourcegraph'

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

test('list tags', async ({ sg, page }) => {
    sg.mockOperations({
        TagsPage_TagsQuery: () => ({
            repository: {
                gitRefs: {
                    nodes: [{ displayName: 'v1.0.0', url: `/${repoName}@v1.0.0` }, { displayName: 'v1.0.1' }],
                    pageInfo: {
                        // Needed to prevent infinity scroll from trying to load more pages
                        hasNextPage: false,
                    },
                    totalCount: 42,
                },
            },
        }),
    })

    await page.goto(`/${repoName}/-/tags`)
    await expect(page.getByRole('link', { name: 'v1.0.0' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'v1.0.1' })).toBeVisible()
    await expect(page.getByText('42 tags total')).toBeVisible()

    // Click on a tag
    await page.getByRole('link', { name: 'v1.0.0' }).click()
    await expect(page).toHaveURL(`/${repoName}@v1.0.0`)
})

test('no tags', async ({ sg, page }) => {
    sg.mockOperations({
        TagsPage_TagsQuery: () => ({
            repository: {
                gitRefs: {
                    nodes: [],
                    totalCount: 0,
                },
            },
        }),
    })

    await page.goto(`/${repoName}/-/tags`)
    await expect(page.getByText('No tags found')).toBeVisible()
})

test('error', async ({ sg, page }) => {
    sg.mockOperations({
        TagsPage_TagsQuery: () => {
            throw new Error('Test error')
        },
    })

    await page.goto(`/${repoName}/-/tags`)
    await expect(page.getByText('Test error')).toBeVisible()
})
