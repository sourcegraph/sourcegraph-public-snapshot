import { test, expect } from '../../../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'

test.beforeEach(({ sg }) => {
    sg.fixture([
        {
            __typename: 'Repository',
            id: '1',
            name: repoName,
            description: 'Example description',
            mirrorInfo: {
                cloned: true,
                cloneInProgress: false,
            },
            commit: {
                id: '2',
            },
        },
        {
            __typename: 'GitCommit',
            id: '2',
            tree: {
                isRoot: true,
                canonicalURL: `/${repoName}/-/tree/`,
                isDirectory: true,
                entries: [
                    {
                        canonicalURL: `/${repoName}/-/blob/index.js`,
                    },
                    {
                        canonicalURL: `/${repoName}/-/blob/README.md`,
                    },
                ],
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

    sg.mockOperations({
        ResolveRepoRevison: () => ({
            repositoryRedirect: {
                id: '1',
            },
        }),
        TreeEntries: () => ({
            repository: {
                id: '1',
            },
        }),
        RepoPageReadmeQuery: ({ repoID, path }) => ({
            node: {
                id: repoID,
                commit: {
                    blob: {
                        canonicalURL: `/${repoName}/-/blob/${path}`,
                    },
                },
            },
        }),
    })
})

test('file sidebar', async ({ page }) => {
    const readmeEntry = page.getByRole('treeitem', { name: 'README.md' })

    await page.goto(`/${repoName}`)
    await expect(readmeEntry).toBeVisible()

    // Close file sidebar
    await page.getByRole('button', { name: 'Hide sidebar' }).click()
    await expect(readmeEntry).toBeHidden()

    // Open sidebar
    await page.getByRole('button', { name: 'Show sidebar' }).click()

    // Go to a file
    await readmeEntry.click()
    await expect(page).toHaveURL(`/${repoName}/-/blob/README.md`)
    // Verify that entry is selected
    await expect(page.getByRole('treeitem', { name: 'README.md', selected: true })).toBeVisible()

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
    sg.mockOperations({
        TreeEntries: () => ({
            repository: {
                commit: {
                    tree: {
                        isRoot: true,
                        entries: [],
                    },
                },
            },
        }),
    })

    await page.goto(`/${repoName}`)
    await expect(page.getByRole('heading', { name: 'Description' })).toBeVisible()
    await expect(page.getByText('Example description')).toBeVisible()
})

test('history panel', async ({ page, sg }) => {
    sg.mockOperations({
        GitHistoryQuery: () => ({
            repository: {
                id: '1',
                commit: {
                    ancestors: {
                        nodes: [{ subject: 'Test commit' }, { subject: 'Test commit 2' }],
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                },
            },
        }),
    })

    await page.goto(`/${repoName}`)

    // Open history panel
    await page.getByRole('tab', { name: 'History' }).click()
    await expect(page.getByText('Test commit', { exact: true })).toBeVisible()

    // Close history panel
    await page.getByRole('tab', { name: 'History' }).click()
    await expect(page.getByText('Test commit')).toBeHidden()
})
