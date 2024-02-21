import { expect, test } from '../../../../../../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/blob/src/index.js`

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
                    canonicalURL: `/${repoName}/-/blob/src/index.js`,
                },
            ],
        },
        {
            __typename: 'GitBlob',
            name: 'index.js',
            path: 'src/index.js',
            canonicalURL: `/${repoName}/-/blob/src/index.js`,
            isDirectory: false,
            languages: ['JavaScript'],
            richHTML: '',
            content: '"file content"',
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
        BlobPageQuery: ({}) => ({
            repository: {
                commit: {
                    blob: {
                        canonicalURL: `/${repoName}/-/blob/src/index.js`,
                    },
                },
            },
        }),
    })
})

test('load file', async ({ page }) => {
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText(/"file content"/)).toBeVisible()
})

test('non-existent file', async ({ page, sg }) => {
    sg.mockOperations({
        BlobPageQuery: ({}) => ({
            repository: {
                commit: {
                    blob: null,
                },
            },
        }),
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText('File not found')).toBeVisible()
})

test('error loading file data', async ({ page, sg }) => {
    sg.mockOperations({
        BlobPageQuery: ({}) => {
            throw new Error('Blob error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText(/Blob error/).first()).toBeVisible()
})

test('error loading highlights data', async ({ page, sg }) => {
    sg.mockOperations({
        BlobSyntaxHighlightQuery: ({}) => {
            throw new Error('Highlights error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText(/"file content"/)).toBeVisible()
    await expect(page.getByText(/Highlights error/).first()).toBeVisible()
})
