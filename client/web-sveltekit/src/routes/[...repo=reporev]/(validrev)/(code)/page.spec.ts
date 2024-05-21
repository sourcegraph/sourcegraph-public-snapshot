import { test, expect, type Page } from '../../../../testing/integration'

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
                        canonicalURL: `/${repoName}/-/tree/src`,
                    },
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
            __typename: 'GitTree',
            path: 'src',
            name: 'src',
            canonicalURL: `/${repoName}/-/tree/src`,
            isDirectory: true,
            isRoot: false,
            entries: [
                {
                    canonicalURL: `/${repoName}/-/blob/src/notes.txt`,
                },
            ],
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
        {
            __typename: 'GitBlob',
            canonicalURL: `/${repoName}/-/blob/src/notes.txt`,
            name: 'notes.txt',
            path: 'src/notes.txt',
            isDirectory: false,
            content: 'Some notes',
            richHTML: '',
        },
    ])

    sg.mockOperations({
        ResolveRepoRevision: () => ({
            repositoryRedirect: {
                id: '1',
            },
        }),
        TreeEntries: () => ({
            repository: {
                id: '1',
            },
        }),
        RepoPageReadmeQuery: ({ path }) => ({
            repository: {
                id: '1',
                commit: {
                    blob: {
                        canonicalURL: `/${repoName}/-/blob/${path}`,
                    },
                },
            },
        }),
    })
})

test.describe('file sidebar', () => {
    async function openSidebar(page: Page): Promise<void> {
        return page.getByLabel('Open sidebar').click()
    }

    test('basic functionality', async ({ page }) => {
        const readmeEntry = page.getByRole('treeitem', { name: 'README.md' })

        await page.goto(`/${repoName}`)

        // Open sidebar
        await page.getByLabel('Open sidebar').click()
        await expect(readmeEntry).toBeVisible()

        // Go to a file
        await readmeEntry.click()
        await expect(page).toHaveURL(`/${repoName}/-/blob/README.md`)
        // Verify that entry is selected
        await expect(page.getByRole('treeitem', { name: 'README.md', selected: true })).toBeVisible()

        // Go to other file
        await page.getByRole('treeitem', { name: 'index.js' }).click()
        await expect(page).toHaveURL(`/${repoName}/-/blob/index.js`)
        // Verify that entry is selected
        await expect(page.getByRole('treeitem', { name: 'index.js', selected: true })).toBeVisible()

        // Close file sidebar
        await page.getByLabel('Close sidebar').click()
        await expect(readmeEntry).toBeHidden()
    })

    test('error handling root', async ({ page, sg }) => {
        sg.mockOperations({
            TreeEntries: () => {
                throw new Error('Sidebar error')
            },
        })

        await page.goto(`/${repoName}`)
        await openSidebar(page)
        await expect(page.getByText(/Sidebar error/)).toBeVisible()
    })

    test('error handling children', async ({ page, sg }) => {
        await page.goto(`/${repoName}`)
        await openSidebar(page)

        const treeItem = page.getByRole('treeitem', { name: 'src' })
        // For some reason we need to wait for the tree to be rendered
        // before we mock the GraphQL response to throw an error
        await expect(treeItem).toBeVisible()

        sg.mockOperations({
            TreeEntries: () => {
                throw new Error('Child error')
            },
        })
        // Clicks the toggle button next to the tree entry, to expand the tree
        // and _not_ follow the link
        await treeItem.getByRole('button').click()
        await expect(page.getByText(/Child error/)).toBeVisible()
    })

    test('error handling non-existing directory -> root', async ({ page, sg }) => {
        // Here we expect the sidebar to show an error message, and after navigigating
        // to an existing directory, the directory contents
        sg.mockOperations({
            TreeEntries: () => {
                throw new Error('Sidebar error')
            },
        })

        await page.goto(`/${repoName}/-/tree/non-existing-directory`)
        await openSidebar(page)
        await expect(page.getByText(/Sidebar error/).first()).toBeVisible()

        sg.mockOperations({
            TreeEntries: () => ({
                repository: {
                    id: '1',
                },
            }),
        })

        await page.goto(`/${repoName}`)
        await openSidebar(page)
        await expect(page.getByRole('treeitem', { name: 'README.md' })).toBeVisible()
    })
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
