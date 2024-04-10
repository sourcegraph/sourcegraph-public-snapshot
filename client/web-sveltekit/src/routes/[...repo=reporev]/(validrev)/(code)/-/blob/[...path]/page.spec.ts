import { ExternalServiceKind } from '../../../../../../../testing/graphql-type-mocks'
import { expect, test } from '../../../../../../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/blob/src/index.js`
const revision = '123'

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
                commit: {
                    oid: revision,
                },
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

test.describe('file header', () => {
    const url = `/${repoName}/-/blob/src/readme.md`

    test.beforeEach(({ sg }) => {
        sg.mockOperations({
            BlobPageQuery: ({}) => ({
                repository: {
                    commit: {
                        blob: {
                            __typename: 'GitBlob',
                            name: 'readme.md',
                            path: 'src/readme.md',
                            canonicalURL: `/${repoName}/-/blob/src/readme.md`,
                            isDirectory: false,
                            languages: ['Markdown'],
                            richHTML: '<h1>file content</h1>',
                            content: '# file content',
                            externalURLs: [
                                {
                                    url: 'https://example.com',
                                    serviceKind: ExternalServiceKind.GITHUB,
                                },
                            ],
                            binary: false,
                            byteSize: 12345,
                            totalLines: 42,
                        },
                    },
                },
            }),
        })
    })

    test('code host link', async ({ page }) => {
        await page.goto(url)
        const link = page.getByLabel('Open in code host')
        await expect(link, 'links to correct code host').toHaveAttribute('href', 'https://example.com')
        await expect(link, 'show code host name').toHaveText('GitHub')
    })

    test('permalink', async ({ page }) => {
        await page.goto(url)
        const link = page.getByRole('link', { name: 'Permalink' })
        await expect(link, 'links to correct revision').toHaveAttribute(
            'href',
            `/${repoName}@${revision}/-/blob/src/readme.md`
        )
    })

    test('dropdown menu', async ({ page }) => {
        await page.goto(url)

        async function openDropdown() {
            await test.step('open dropdown (if necessary)', async () => {
                if (!(await page.getByRole('menuitem', { name: 'View raw' }).isVisible())) {
                    await page.getByLabel('Show more actions').click()
                }
            })
        }

        await openDropdown()
        await expect(page.getByRole('menuitem', { name: 'View raw' }), 'dropdown menu opens').toBeVisible()

        await expect(
            page.getByRole('menuitem', { name: 'View raw' }),
            '"view raw" links to correct URL'
        ).toHaveAttribute('href', `/${repoName}/-/raw/src/readme.md`)

        const lineWrappingOption = page.getByRole('menuitem', { name: 'Enable wrapping long lines' })
        await expect(lineWrappingOption, 'line wrapping is disabled for formatted view').toBeDisabled()

        await test.step('switch to code view', () => page.getByLabel('Code', { exact: true }).click())
        await openDropdown()
        await expect(lineWrappingOption, 'line wrapping is enabled for code view').toBeEnabled()

        await lineWrappingOption.click()
        await openDropdown()
        await expect(
            page.getByRole('menuitem', { name: 'Disable wrapping long lines' }),
            'line wrapping option was updated'
        ).toBeVisible()
    })

    test('view modes', async ({ page }) => {
        await page.goto(url)
        // Rendered markdown is shown by default
        await expect(page.getByLabel('Formatted'), "'Formatted' is selected by default").toBeChecked()
        await expect(page.getByRole('heading', { name: 'file content' })).toBeVisible()

        // Switch to raw content view
        const codeOption = page.getByLabel('Code', { exact: true })
        await codeOption.click()
        await expect(codeOption, "'Code' is selected").toBeChecked()
        await expect(page.getByText(/# file content/)).toBeVisible()
    })

    test('meta data', async ({ page }) => {
        await page.goto(url)
        await expect(page.getByText('12.35 KB')).toBeVisible()
        await expect(page.getByText('42 lines')).toBeVisible()
    })
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
