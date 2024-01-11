import { test, expect } from '../../../../testing/integration'

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
        {
            __typename: 'GitBlob',
            path: 'index.js',
            name: 'index.js',
            canonicalURL: `/${repoName}/-/blob/index.js`,
            isDirectory: false,
            languages: ['JavaScript'],
            richHTML: '',
            content: 'var hello = "world"',
        },
        {
            __typename: 'GitBlob',
            canonicalURL: `/${repoName}/-/blob/README.md`,
            name: 'README.md',
            path: 'README.md',
            isDirectory: false,
            richHTML: 'Example readme content',
        },
    ])

    sg.mock({
        Query: () => ({
            repositoryRedirect: {
                __typename: 'Repository',
                id: '1',
            },
        }),
        GitCommit: info => ({
            oid: 'test',
            tree: {
                isRoot: true,
                canonicalURL: `/${repoName}/-/tree/`,
                isDirectory: true,
                entries: [
                    {
                        __typename: 'GitBlob',
                        canonicalURL: `/${repoName}/-/blob/index.js`,
                    },
                    {
                        __typename: 'GitBlob',
                        canonicalURL: `/${repoName}/-/blob/README.md`,
                    },
                ],
            },
            blob: {
                canonicalURL: `/${repoName}/-/blob/README.md}`,
            },
        }),
    })
})

test('file sidebar', async ({ page, sg }) => {
    const readmeEntry = page.getByRole('treeitem', { name: 'README.md' })

    await page.goto(`/${repoName}`)
    await expect(readmeEntry).toBeVisible()

    // Close file sidebar
    await page.getByRole('button', { name: 'Hide sidebar' }).click()
    await expect(readmeEntry).toBeHidden()

    // Open sidebar
    await page.getByRole('button', { name: 'Show sidebar' }).click()

    sg.mock({
        GitCommit: () => ({
            blob: {
                richHTML: 'Example readme content',
            },
        }),
    })

    // Go to a file
    await readmeEntry.click()
    await expect(page).toHaveURL(`/${repoName}/-/blob/README.md`)
    // Verify that entry is selected
    await expect(page.getByRole('treeitem', { name: 'README.md', selected: true })).toBeVisible()

    sg.mock({
        GitCommit: () => ({
            blob: {
                richHTML: 'index.js',
            },
        }),
    })

    // Go other file
    await page.getByRole('treeitem', { name: 'index.js' }).click()
    await expect(page).toHaveURL(`/${repoName}/-/blob/index.js`)
    // Verify that entry is selected
    await expect(page.getByRole('treeitem', { name: 'index.js', selected: true })).toBeVisible()
})

test('repo readme', async ({ page }) => {
    // Shows the readme file if there is one in the repo root
    await page.goto(`/${repoName}`)
    await expect(page.getByRole('heading', { name: 'README.md' })).toBeVisible()
    await expect(page.getByText('Example readme content')).toBeVisible()
})

test('repo description', async ({ page, sg }) => {
    // Shows the repo description if there is no readme file in the repo root
    sg.mock({
        Repository: () => ({
            description: 'Example description',
        }),
        GitCommit: () => ({
            tree: {
                isRoot: true,
                entries: [],
            },
        }),
    })

    await page.goto(`/${repoName}`)
    await expect(page.getByRole('heading', { name: 'Description' })).toBeVisible()
    await expect(page.getByText('Example description')).toBeVisible()
})

test('history panel', async ({ page, sg }) => {
    sg.mock(
        {
            GitCommit: () => ({
                subject: 'Test commit',
            }),
        },
        'GitHistoryQuery'
    )

    await page.goto(`/${repoName}`)

    // Open history panel
    await page.getByRole('tab', { name: 'History' }).click()
    await expect(page.getByText('Test commit')).toBeVisible()

    // Close history panel
    await page.getByRole('tab', { name: 'History' }).click()
    await expect(page.getByText('Test commit')).toBeHidden()
})
