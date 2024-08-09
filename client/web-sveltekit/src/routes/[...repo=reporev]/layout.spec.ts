import { ExternalServiceKind } from '$testing/graphql-type-mocks'
import { test, expect } from '$testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'

test.beforeEach(({ sg }) => {
    sg.fixture([
        {
            __typename: 'Repository',
            id: '1',
            mirrorInfo: {
                cloned: true,
                cloneInProgress: false,
            },
        },
    ])
})

test.describe('cloned repository', () => {
    test.beforeEach(async ({ sg, page }) => {
        sg.mockOperations({
            ResolveRepoRevision: ({ repoName }) => ({
                repositoryRedirect: {
                    id: '1',
                    name: repoName,
                    commit: {
                        oid: '123456789',
                    },
                },
            }),
        })
        await page.goto(`/${repoName}`)
    })

    test('shows repo name in header', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'sourcegraph/sourcegraph' })).toBeVisible()
    })

    // TODO: Better test to ensure that we are testing the search input
    test.fixme('has prepopulated search bar', async ({ page }) => {
        await expect(page.getByText('repo:^github\\.com/sourcegraph')).toBeVisible()
    })
})

test('clone in progress', async ({ sg, page }) => {
    sg.mockOperations({
        ResolveRepoRevision: ({ repoName }) => ({
            repositoryRedirect: {
                id: '1',
                name: repoName,
                mirrorInfo: {
                    cloneInProgress: true,
                    cloneProgress: 'Test clone message',
                },
            },
        }),
    })

    await page.goto(`/${repoName}`)

    // Shows repo name
    await expect(page.getByRole('heading', { name: 'sourcegraph/sourcegraph' })).toBeVisible()
    // Shows clone progress message
    await expect(page.getByText('Test clone message')).toBeVisible()
})

test('not cloned', async ({ sg, page }) => {
    sg.mockOperations({
        ResolveRepoRevision: ({ repoName }) => ({
            repositoryRedirect: {
                id: '1',
                name: repoName,
                mirrorInfo: {
                    cloned: false,
                    cloneInProgress: false,
                },
            },
        }),
    })

    await page.goto(`/${repoName}`)

    // Shows repo name
    await expect(page.getByRole('heading', { name: 'sourcegraph/sourcegraph' })).toBeVisible()
    // Shows queue message
    await expect(page.getByText('queued for cloning')).toBeVisible()
})

test('not found', async ({ sg, page }) => {
    sg.mockOperations({
        ResolveRepoRevision: () => ({
            repositoryRedirect: null,
        }),
    })

    await page.goto(`/${repoName}`)

    // Shows not found error message
    await expect(page.getByRole('heading', { name: 'Repository not found' })).toBeVisible()
})

test.describe('repo menu', () => {
    test.beforeEach(async ({ sg, page }) => {
        sg.mockOperations({
            ResolveRepoRevision: ({ repoName }) => ({
                repositoryRedirect: {
                    id: '1',
                    name: repoName,
                    commit: {
                        oid: '123456789',
                    },
                    externalURLs: [
                        {
                            serviceKind: ExternalServiceKind.GITHUB,
                            url: 'https://github.com/sourcegraph/sourcegraph',
                        },
                    ],
                },
            }),
        })
        await page.goto(`/${repoName}`)
    })

    test('click switch repo', async ({ page }) => {
        await page.getByRole('heading', { name: 'sourcegraph/sourcegraph' }).click()
        await page.getByRole('menuitem', { name: 'Switch repo' }).click()
        await expect(page.getByPlaceholder('Find repositories...')).toBeVisible()
    })

    test('settings url', async ({ page }) => {
        await page.getByRole('heading', { name: 'sourcegraph/sourcegraph' }).click()
        const url = await page.getByRole('menuitem', { name: 'Settings' }).getAttribute('href')
        expect(url).toEqual(`/${repoName}/-/settings`)
    })

    test('github url', async ({ page }) => {
        await page.getByRole('heading', { name: 'sourcegraph/sourcegraph' }).click()
        const url = await page.getByRole('menuitem', { name: 'Hosted on GitHub' }).getAttribute('href')
        expect(url).toEqual(`https://github.com/sourcegraph/sourcegraph`)
    })
})
