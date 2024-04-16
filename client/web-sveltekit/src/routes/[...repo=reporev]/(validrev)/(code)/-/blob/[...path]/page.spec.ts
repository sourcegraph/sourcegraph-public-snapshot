import {ExternalServiceKind} from '../../../../../../../testing/graphql-type-mocks'
import {expect, test} from '../../../../../../../testing/integration'

const repoName = 'github.com/sourcegraph/sourcegraph'
const url = `/${repoName}/-/blob/src/index.js`
const revision = '123'

const settings = '{\n  "openInEditor": {\n    "projectPaths.default": "/Users/michael/WebstormProjects",\n    "editorIds": [\n      "idea"\n    ],\n  }\n}'

test.beforeEach(({sg}) => {
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
        // {
        //     __typename: 'SettingsCascade',
        //     id: '1',
        //     subjects: [
        //         {
        //             __typename: 'DefaultSettings',
        //             id: 'TestDefaultSettingsID',
        //             settingsURL: null,
        //             viewerCanAdminister: false,
        //             latestSettings: {
        //                 id: 0,
        //                 contents: settings,
        //             },
        //         },
        //     ],
        //     final: settings,
        // }
    ])

    sg.mockOperations({
        ViewerSettings: () => ({
            viewerSettings: {
                __typename: 'SettingsCascade',
                subjects: [
                    {
                        __typename: 'DefaultSettings',
                        id: 'TestDefaultSettingsID',
                        settingsURL: null,
                        viewerCanAdminister: false,
                        latestSettings: {
                            id: 0,
                            contents: settings,
                        },
                    },
                ],
                final: settings,
            },
        }),
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

test('load file', async ({page}) => {
    await page.goto(url)
    await expect(page.getByRole('heading', {name: 'index.js'})).toBeVisible()
    await expect(page.getByText(/"file content"/)).toBeVisible()
})

test.describe('file header', () => {
    const url = `/${repoName}/-/blob/src/readme.md`

    test.beforeEach(({sg}) => {
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

    test('default editor link', async ({page}) => {
        await page.goto(url)
        const link = page.getByLabel('Editor')
        await expect(link, 'links to correct code host').toHaveAttribute('href', '/help/integration/open_in_editor')

        await link.focus();
        const tooltip = page.getByRole('tooltip');
        await expect(tooltip, 'inform user about settings').toHaveText('Add `openInEditor` to your user settings to open files in the editor. Click to learn more.')
    })

    test('editor link', async ({sg, page}) => {
        // sg.fixture([{
        //     __typename: 'SettingsCascade',
        //     subjects: [
        //         {
        //             __typename: 'DefaultSettings',
        //             id: 'TestDefaultSettingsID',
        //             settingsURL: null,
        //             viewerCanAdminister: false,
        //             latestSettings: {
        //                 id: 0,
        //                 contents: settings,
        //             },
        //         },
        //     ],
        //     final: settings,
        // }])

        sg.mockTypes({
            ViewerSettings: () => ({
                viewerSettings: {
                    __typename: 'SettingsCascade',
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            id: 'TestDefaultSettingsID',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: settings,
                            },
                        },
                    ],
                    final: settings,
                },
            }),
        })

        sg.mockOperations({
            ViewerSettings: () => ({
                viewerSettings: {
                    __typename: 'SettingsCascade',
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            id: 'TestDefaultSettingsID',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: settings,
                            },
                        },
                    ],
                    final: settings,
                },
            }),
        })
        // sg.fixture([
        //     {
        //         __typename: 'SettingsCascade',
        //         id: '1',
        //         subjects: [
        //             {
        //                 'latestSettings': {
        //                     'id': 1,
        //                     'contents': '{\n  "openInEditor": {\n    "projectPaths.default": "/Users/michael/WebstormProjects",\n    "editorIds": [\n      "idea"\n    ],\n  }\n}'
        //                 }
        //             },
        //         ]
        //     }
        // ])
        // sg.fixture([
        //     {
        //         __typename: 'SettingsCascade',
        //         subjects: [
        // {
        //     "latestSettings": {
        //         "id": 0,
        //         "contents": "{\"experimentalFeatures\": {}}"
        //     }
        // },
        // {
        //     "latestSettings": {
        //         "id": 3534,
        //         "contents": "{\n  \"notices\": [\n    {\n      \"dismissible\": true,\n      \"location\": \"top\",\n      \"message\": \"ℹ️ Are you looking for infra access to troubleshoot problems? Please visit [go/s2-ops](http://go/s2-ops) for next steps. Reach out to [#discuss-cloud-ops](https://sourcegraph.slack.com/archives/discuss-cloud-ops) if you need help.\"\n    },\n    {\n      \"dismissible\": false,\n      \"location\": \"home\",\n      \"message\": \"🚨 Note that this deployment has private code and sensitive customers information - for demos, please use [demo.sourcegraph.com](https://demo.sourcegraph.com) instead.\"\n    },\n    {\n      \"dismissible\": true,\n      \"location\": \"top\",\n      \"message\": \"🚨 Note that this deployment has private code and sensitive customers information - for demos, please use [demo.sourcegraph.com](https://demo.sourcegraph.com) instead.\"\n    }\n  ],\n  // add settings here (Ctrl+Space to see hints)\n  // experimentalFeatures.batchChangesExecution\n  \"experimentalFeatures\": {\n    \"batchChangesExecution\": true,\n    \"codeInsights\": true,\n    \"codeMonitoring\": true,\n    \"showSearchContext\": true,\n    \"showSearchContextManagement\": true,\n    \"showSearchNotebook\": true,\n    \"searchContextsQuery\": true,\n    \"codeMonitoringWebHooks\": true,\n    \"fuzzyFinderAll\": true,\n    \"enableSearchStack\": true,\n    \"coolCodeIntel\": true,\n    \"proactiveSearchResultsAggregations\": true,\n    // Enabled by Olaf Geirsson 2022/12/05\n    \"codeNavigation\": \"selection-driven\",\n    // Enabled by SQS on 2022-12-19\n    \"applySearchQuerySuggestionOnEnter\": true,\n    // Enabled by @michael.lin 2023-09-13\n    \"searchJobs\": true,\n    // Set by Dax 2023-10-25\n    \"searchQueryInput\": \"v2\",\n    // \"newSearchNavigationUI\": false,\n    \"newSearchResultsUI\": true,\n    \"newSearchResultFiltersPanel\": true,\n    \"keywordSearch\": true,\n  },\n\n  \"codeIntel.lsif\": true,\n  \"codeIntel.blobKeyboardNavigation\": \"token\",\n  \"search.scopes\": [\n    {\n      \"name\": \"Mike Test code\",\n      \"value\": \"file:(test|spec)\"\n    }\n  ],\n  \"alerts.ShowPatchUpdates\": false\n}"
        //     }
        // },
        // {
        //     "latestSettings": {
        //         "id": 3359,
        //         "contents": "{\n  \"openInEditor\": {\n    \"projectPaths.default\": \"/Users/michael/WebstormProjects\",\n    \"editorIds\": [\n      \"idea\"\n    ],\n  }\n}"
        //     }
        // },
        //             {
        //                 latestSettings: {
        //                     id: 0,
        //                     contents: '{\n  "openInEditor": {\n    "projectPaths.default": "/Users/michael/WebstormProjects",\n    "editorIds": [\n      "idea"\n    ],\n  }\n}'
        //                 }
        //             }
        //         ]
        //     },
        // ])

        sg.signIn({username: 'test'})
        await page.goto(url)
        const link = page.getByLabel('Editor')
        await expect(link, 'links to correct code host').toHaveAttribute('href', '/help/integration/open_in_editor')

        await link.focus();
        const tooltip = page.getByRole('tooltip');
        await expect(tooltip, 'inform user about settings').toHaveText('Add `openInEditor` to your user settings to open files in the editor. Click to learn more.')
    })

    test('code host link', async ({page}) => {
        await page.goto(url)
        const link = page.getByLabel('Open in code host')
        await expect(link, 'links to correct code host').toHaveAttribute('href', 'https://example.com')
        await expect(link, 'show code host name').toHaveText('GitHub')
    })

    test('permalink', async ({page}) => {
        await page.goto(url)
        const link = page.getByRole('link', {name: 'Permalink'})
        await expect(link, 'links to correct revision').toHaveAttribute(
            'href',
            `/${repoName}@${revision}/-/blob/src/readme.md`
        )
    })

    test('dropdown menu', async ({page}) => {
        await page.goto(url)

        async function openDropdown() {
            await test.step('open dropdown (if necessary)', async () => {
                if (!(await page.getByRole('menuitem', {name: 'View raw'}).isVisible())) {
                    await page.getByLabel('Show more actions').click()
                }
            })
        }

        await openDropdown()
        await expect(page.getByRole('menuitem', {name: 'View raw'}), 'dropdown menu opens').toBeVisible()

        await expect(
            page.getByRole('menuitem', {name: 'View raw'}),
            '"view raw" links to correct URL'
        ).toHaveAttribute('href', `/${repoName}/-/raw/src/readme.md`)

        const lineWrappingOption = page.getByRole('menuitem', {name: 'Enable wrapping long lines'})
        await expect(lineWrappingOption, 'line wrapping is disabled for formatted view').toBeDisabled()

        await test.step('switch to code view', () => page.getByLabel('Code', {exact: true}).click())
        await openDropdown()
        await expect(lineWrappingOption, 'line wrapping is enabled for code view').toBeEnabled()

        await lineWrappingOption.click()
        await openDropdown()
        await expect(
            page.getByRole('menuitem', {name: 'Disable wrapping long lines'}),
            'line wrapping option was updated'
        ).toBeVisible()
    })

    test('view modes', async ({page}) => {
        await page.goto(url)
        // Rendered markdown is shown by default
        await expect(page.getByLabel('Formatted'), '\'Formatted\' is selected by default').toBeChecked()
        await expect(page.getByRole('heading', {name: 'file content'})).toBeVisible()

        // Switch to raw content view
        const codeOption = page.getByLabel('Code', {exact: true})
        await codeOption.click()
        await expect(codeOption, '\'Code\' is selected').toBeChecked()
        await expect(page.getByText(/# file content/)).toBeVisible()
    })

    test('meta data', async ({page}) => {
        await page.goto(url)
        await expect(page.getByText('12.35 KB')).toBeVisible()
        await expect(page.getByText('42 lines')).toBeVisible()
    })
})

test('non-existent file', async ({page, sg}) => {
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
    await expect(page.getByRole('heading', {name: 'index.js'})).toBeVisible()
    await expect(page.getByText('File not found')).toBeVisible()
})

test('error loading file data', async ({page, sg}) => {
    sg.mockOperations({
        BlobPageQuery: ({}) => {
            throw new Error('Blob error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', {name: 'index.js'})).toBeVisible()
    await expect(page.getByText(/Blob error/).first()).toBeVisible()
})

test('error loading highlights data', async ({page, sg}) => {
    sg.mockOperations({
        BlobSyntaxHighlightQuery: ({}) => {
            throw new Error('Highlights error')
        },
    })
    await page.goto(url)
    await expect(page.getByRole('heading', {name: 'index.js'})).toBeVisible()
    await expect(page.getByText(/"file content"/)).toBeVisible()
    await expect(page.getByText(/Highlights error/).first()).toBeVisible()
})
