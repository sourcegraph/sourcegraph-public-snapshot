import { test, expect } from '../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'

test.beforeEach(({ sg }) => {
    sg.fixture([
        {
            __typename: 'Repository',
            id: '1',
            name: 'github.com/sourcegraph/sourcegraph',
            mirrorInfo: {
                cloned: true,
                cloneInProgress: false,
            },
        },
    ])
})

test.describe('cloned repository', () => {
    test.beforeEach(async ({ sg, page }) => {
        sg.mock({
            Query: () => ({
                repositoryRedirect: {
                    __typename: 'Repository',
                    id: '1',
                },
            }),
        })
        await page.goto(`/${repoName}`)
    })

    test('shows repo name in header', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'sourcegraph/sourcegraph' })).toBeVisible()
    })

    test('has search button', async ({ page }) => {
        await page.getByRole('button', { name: 'Search' }).click()
        await expect(page.getByRole('textbox')).toHaveText('repo:github.com/sourcegraph/sourcegraph ')
    })
})

test('clone in progress', async ({ sg, page }) => {
    sg.mock({
        Query: () => ({
            repositoryRedirect: {
                __typename: 'Repository',
                id: '1',
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
    sg.mock({
        Query: () => ({
            repositoryRedirect: {
                __typename: 'Repository',
                id: '1',
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
