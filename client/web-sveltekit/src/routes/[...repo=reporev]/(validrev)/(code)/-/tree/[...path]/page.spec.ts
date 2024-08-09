import { expect, test } from '$testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/tree/src`

test.beforeEach(({ sg }) => {
    sg.fixture([
        {
            __typename: 'Repository',
            id: '1',
            name: repoName,
            mirrorInfo: {
                cloned: true,
                cloneInProgress: false,
            },
        },
        {
            __typename: 'GitTree',
            name: 'src',
            path: 'src',
            canonicalURL: `/${repoName}/-/tree/src`,
            isDirectory: true,
            isRoot: false,
            entries: [
                {
                    canonicalURL: `/${repoName}/-/tree/src/notes`,
                },
                {
                    canonicalURL: `/${repoName}/-/blob/src/index.js`,
                },
                {
                    canonicalURL: `/${repoName}/-/blob/src/README.md`,
                },
            ],
        },
        {
            __typename: 'GitTree',
            name: 'notes',
            path: 'src/notes',
            canonicalURL: `/${repoName}/-/tree/src/notes`,
            isDirectory: true,
            isRoot: false,
        },
        {
            __typename: 'GitBlob',
            name: 'index.js',
            path: 'src/index.js',
            canonicalURL: `/${repoName}/-/blob/src/index.js`,
            isDirectory: false,
            languages: ['JavaScript'],
            richHTML: '',
            content: 'var hello = "world"',
        },
        {
            __typename: 'GitBlob',
            canonicalURL: `/${repoName}/-/blob/src/README.md`,
            name: 'README.md',
            path: 'src/README.md',
            isDirectory: false,
            richHTML: 'Example readme content',
        },
    ])

    sg.mockOperations({
        ResolveRepoRevision: () => ({
            repositoryRedirect: {
                id: '1',
            },
        }),
        TreeEntries: ({}) => ({
            repository: {
                commit: {
                    tree: {
                        canonicalURL: `/${repoName}/-/tree/src`,
                    },
                },
            },
        }),
        TreePageCommitInfoQuery: ({}) => ({
            repository: {
                commit: {
                    tree: {
                        canonicalURL: `/${repoName}/-/tree/src`,
                    },
                },
            },
        }),
        TreePageReadmeQuery: ({ path }) => ({
            repository: {
                commit: {
                    blob: {
                        canonicalURL: `/${repoName}/-/blob/${path}`,
                    },
                },
            },
        }),
    })
})

test('list files in a directory', async ({ page }) => {
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'src' })).toBeVisible()

    await expect(page.getByRole('cell', { name: 'notes' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'index.js' })).toBeVisible()
})

test('shows README if available', async ({ page, sg }) => {
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'README.md' })).toBeVisible()
    await expect(page.getByText('Example readme content')).toBeVisible()

    // Not available

    sg.mockOperations({
        TreeEntries: ({}) => ({
            repository: {
                commit: {
                    tree: {
                        canonicalURL: `/${repoName}/-/tree/src`,
                        entries: [
                            {
                                canonicalURL: `/${repoName}/-/blob/src/index.js`,
                            },
                        ],
                    },
                },
            },
        }),
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'src' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'README.md' })).not.toBeVisible()
})

test('empty tree', async ({ page, sg }) => {
    sg.mockOperations({
        TreeEntries: ({}) => ({
            repository: {
                commit: {
                    tree: {
                        canonicalURL: `/${repoName}/-/tree/src`,
                        entries: [],
                    },
                },
            },
        }),
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'src' })).toBeVisible()
    await expect(page.getByText('This directory is empty')).toBeVisible()
})

test('non-existent tree', async ({ page, sg }) => {
    sg.mockOperations({
        TreeEntries: ({}) => ({
            repository: {
                commit: {
                    tree: null,
                },
            },
        }),
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'src' })).toBeVisible()
    await expect(page.getByText('Directory not found')).toBeVisible()
})

test('error loading tree data', async ({ page, sg }) => {
    sg.mockOperations({
        TreeEntries: ({}) => {
            throw new Error('Sentinel error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'src' })).toBeVisible()
    await expect(page.getByText(/Unable to load directory.*Sentinel error/)).toBeVisible()
})

test('error loading commit data', async ({ page, sg }) => {
    sg.mockOperations({
        TreePageCommitInfoQuery: ({}) => {
            throw new Error('Commit info error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'src' })).toBeVisible()
    await expect(page.getByText(/Commit info error/)).toBeVisible()
})
