import { ExternalServiceKind } from '$testing/graphql-type-mocks'
import { expect, test } from '$testing/integration'

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
                {
                    canonicalURL: `/${repoName}/-/blob/src/large-file-1.js`,
                },
                {
                    canonicalURL: `/${repoName}/-/blob/src/large-file-2.js`,
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
            binary: false,
        },
        {
            __typename: 'GitBlob',
            name: 'large-file-1.js',
            path: 'src/large-file-1.js',
            canonicalURL: `/${repoName}/-/blob/src/large-file-1.js`,
            isDirectory: false,
            languages: ['JavaScript'],
            richHTML: '',
            content: Array.from({ length: 500 }, (_, i) => `// line ${i + 1};`).join('\n'),
            totalLines: 500,
            binary: false,
        },
        {
            __typename: 'GitBlob',
            name: 'large-file-2.js',
            path: 'src/large-file-2.js',
            canonicalURL: `/${repoName}/-/blob/src/large-file-2.js`,
            isDirectory: false,
            languages: ['JavaScript'],
            richHTML: '',
            content: Array.from({ length: 300 }, (_, i) => `// line ${i + 1};`).join('\n'),
            totalLines: 300,
            binary: false,
        },
        {
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
        TreeEntries: ({ repoName }) => ({
            repository: {
                commit: {
                    tree: {
                        canonicalURL: `/${repoName}/-/tree/src`,
                    },
                },
            },
        }),
        BlobFileViewBlobQuery: ({ path, repoName }) => ({
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

test('load file', async ({ page }) => {
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText(/"file content"/)).toBeVisible()
})

test.describe('file header', () => {
    const url = `/${repoName}/-/blob/src/readme.md`

    test.skip('default editor link', async ({ page }) => {
        await page.goto(url)
        const link = page.getByLabel('Editor')
        await expect(link, 'links to help page').toHaveAttribute('href', '/help/integration/open_in_editor')

        await link.focus()
        const tooltip = page.getByRole('tooltip')
        await expect(tooltip, 'inform user about settings').toHaveText(
            'Add `openInEditor` to your user settings to open files in the editor. Click to learn more.'
        )
    })

    test('editor link', async ({ sg, page }) => {
        const projectsPath = '/Users/USERNAME/Documents'
        sg.mockTypes({
            SettingsCascade: () => ({
                final: JSON.stringify({
                    openInEditor: { editorIds: ['idea'], 'projectPaths.default': projectsPath },
                }),
            }),
        })

        await sg.signIn({ username: 'test' })
        await page.goto(url)
        const link = page.getByLabel('Open in IntelliJ IDEA')
        await expect(link, 'links to correct editor').toHaveAttribute(
            'href',
            `idea://open?file=${projectsPath}/sourcegraph/sourcegraph/src/readme.md&line=1&column=1`
        )
        await link.focus()

        const tooltip = page.getByRole('tooltip')
        await expect(tooltip, 'inform user about settings').toHaveText('Open in IntelliJ IDEA')
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

    // Disabled because flaky in CI
    test.fixme('dropdown menu', async ({ page }) => {
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

    test.describe('breadcrumbs', () => {
        test('links work', async ({ page }) => {
            await page.goto(url)
            const parentBreadcrumb = page.getByRole('link', { name: 'src' })
            await expect(parentBreadcrumb).toBeVisible()
            await parentBreadcrumb.click()
            await page.waitForURL(`${repoName}/-/tree/src`)
            await expect(page.getByRole('link', { name: 'src' })).toBeVisible()
        })

        test('textContent is exactly the path', async ({ page, context }) => {
            await context.grantPermissions(['clipboard-read', 'clipboard-write'])
            await page.goto(url)
            // We specifically check the textContent here because this is what is
            // used to apply highlights. It must exactly equal the path (no additional
            // whitespace) or the highlights will be incorrectly offset.
            const pathContainer = page.locator('css=[data-path-container]').first()
            await expect(pathContainer).toHaveText(/^src\/readme.md$/)
        })

        test('copy path button', async ({ page, context }) => {
            await context.grantPermissions(['clipboard-read', 'clipboard-write'])
            await page.goto(url)
            await page.getByRole('link', { name: 'src' }).hover()
            await page.getByLabel('Copy path to clipboard').click()
            const clipboardText = await page.evaluate('navigator.clipboard.readText()')
            expect(clipboardText, 'path should be copied to clipboard').toBe('src/readme.md')
        })
    })
})

test.describe('repo menu', () => {
    test('click go to root', async ({ page }) => {
        const url = `/${repoName}/-/blob/src/large-file-1.js`
        await page.goto(url)

        await page.getByRole('heading', { name: 'sourcegraph/sourcegraph' }).click()
        await page.getByRole('menuitem', { name: 'Go to repository root' }).click()
        await page.waitForURL(`/${repoName}`)
    })

    test('keyboard shortcut go to root', async ({ page }) => {
        const url = `/${repoName}/-/blob/src/large-file-1.js`
        await page.goto(url)
        // Focus _something_ on the page. Use both mac and linux shortcuts so this works
        // both locally and in CI.
        await page.getByRole('link').first().press('Meta+Backspace')
        await page.getByRole('link').first().press('Control+Backspace')
        await page.waitForURL(`/${repoName}`)
    })
})

test.describe('scroll behavior', () => {
    const url = `/${repoName}/-/blob/src/large-file-1.js`

    test('initial page load', async ({ page }) => {
        await page.goto(url)
        await expect(page.getByText('line 1;'), 'file is scrolled to the top').toBeVisible()
    })

    test('initial page load with selected line', async ({ page }) => {
        await page.goto(url + '?L100')
        const selectedLine = page.getByTestId('selected-line')
        await expect(selectedLine, 'selected line is scrolled into view').toBeVisible()
        await expect(selectedLine).toHaveText(/line 100;/)
    })

    test('go to another file', async ({ page, utils }) => {
        await page.goto(url)
        // Scroll to some arbitrary position
        await utils.scrollYAt(page.getByText('line 1;'), 1000)

        // Open sidebar
        await page.getByLabel('Open sidebar').click()

        await page.getByRole('link', { name: 'large-file-2.js' }).click()
        await expect(page.getByText('line 1;')).toBeVisible()
    })

    test.skip('select a line', async ({ page, utils }) => {
        await page.goto(url)

        // Scrolls to line 64 at the top (found out by inspecting the test)
        await utils.scrollYAt(page.getByText('line 1;'), 1000)
        const line64 = page.getByText('line 64;')
        await expect(line64).toBeVisible()
        const position = await line64.boundingBox()

        // Select line
        await page.getByText('70', { exact: true }).click()
        await expect(page.getByTestId('selected-line')).toHaveText(/line 70;/)

        // Compare positions
        expect((await line64.boundingBox())?.y, 'selecting a line preserves scroll position').toBe(position?.y)
    })

    test.skip('[back] preserve scroll position', async ({ page, utils }) => {
        await page.goto(url)
        const line1 = page.getByText('line 1;')
        await expect(line1).toBeVisible()

        // Scrolls to line 64 at the top (found out by inspecting the test)
        await utils.scrollYAt(line1, 1000)
        const line64 = page.getByText('line 64;')
        await expect(line64).toBeVisible()
        const position = await line64.boundingBox()

        // Open sidebar
        await page.locator('#sidebar-panel').getByRole('button').click()

        await page.getByRole('link', { name: 'large-file-2.js' }).click()
        await expect(line1).toBeVisible()

        await page.goBack()
        expect((await page.getByText('line 64;').boundingBox())?.y, 'restores scroll position on back navigation').toBe(
            position?.y
        )
    })

    test.skip('[forward] preserve scroll position', async ({ page, utils }) => {
        await page.goto(url)

        // Open sidebar
        await page.locator('#sidebar-panel').getByRole('button').click()

        await page.getByRole('link', { name: 'large-file-2.js' }).click()

        const firstLine = page.getByText('line 1;')
        await expect(firstLine).toBeVisible()

        // Scrolls to line 64 at the top (found out by inspecting the test)
        await utils.scrollYAt(firstLine, 1000)
        const line64 = page.getByText('line 64;')
        await expect(line64).toBeVisible()
        const position = await line64.boundingBox()

        await page.goBack()
        await expect(page.getByText('/ large-file-1.js')).toBeVisible()
        await page.goForward()
        await expect(page.getByText('/ large-file-2.js')).toBeVisible()

        expect((await line64.boundingBox())?.y, 'restores scroll navigation on forward navigation').toBe(position?.y)
    })

    test.skip('[back] preserve scroll position with selected line', async ({ page, utils }) => {
        await page.goto(url + '?L100')
        const line100 = page.getByText('line 100;')
        await expect(line100).toBeVisible()

        // Scrolls to line 210 at the top (found out by inspecting the test)
        await utils.scrollYAt(line100, 2000)
        const line210 = page.getByText('line 210;')
        await expect(line210).toBeVisible()
        const position = await line210.boundingBox()

        // Open sidebar
        await page.locator('#sidebar-panel').getByRole('button').click()

        await page.getByRole('link', { name: 'large-file-2.js' }).click()
        await expect(page.getByText('line 1;')).toBeVisible()

        // This should restore the previous scroll position, not go to the selected line
        await page.goBack()
        expect((await line210.boundingBox())?.y, 'restores scroll position on back navigation').toBe(position?.y)
    })
})

test('non-existent file', async ({ page, sg }) => {
    sg.mockOperations({
        BlobFileViewBlobQuery: ({}) => ({
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
        BlobFileViewBlobQuery: ({}) => {
            throw new Error('Blob error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText(/Blob error/).first()).toBeVisible()
})

test.skip('error loading highlights data', async ({ page, sg }) => {
    sg.mockOperations({
        BlobFileViewHighlightedFileQuery: ({}) => {
            throw new Error('Highlights error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', { name: 'index.js' })).toBeVisible()
    await expect(page.getByText(/"file content"/)).toBeVisible()
    await expect(page.getByText(/Highlights error/).first()).toBeVisible()
})
