import { screen } from '@testing-library/react'
import { noop } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { RecentFilesPanel } from './RecentFilesPanel'

describe('RecentFilesPanel', () => {
    test('duplicate files are only shown once', () => {
        const recentFiles = {
            nodes: [
                {
                    argument: '{"filePath": "go.mod", "repoName": "ghe.sgdev.org/sourcegraph/gorilla-mux"}',
                    timestamp: '2020-09-10T22:55:30Z',
                    url: 'https://sourcegraph.test:3443/ghe.sgdev.org/sourcegraph/gorilla-mux/-/blob/go.mod',
                },
                {
                    argument: '{"filePath": ".eslintrc.js", "repoName": "github.com/sourcegraph/sourcegraph"}',
                    timestamp: '2020-09-10T22:55:18Z',
                    url: 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/.eslintrc.js',
                },
                {
                    argument: '{"filePath": "go.mod", "repoName": "ghe.sgdev.org/sourcegraph/gorilla-mux"}',
                    timestamp: '2020-09-10T22:55:06Z',
                    url: 'https://sourcegraph.test:3443/ghe.sgdev.org/sourcegraph/gorilla-mux/-/blob/go.mod',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 3,
        }

        const props = {
            authenticatedUser: null,
            recentFilesFragment: { recentFilesLogs: recentFiles },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        renderWithBrandedContext(<RecentFilesPanel {...props} />)
        const listItems = screen.getAllByTestId('recent-files-item')
        expect(listItems).toHaveLength(2)
        expect(listItems[0]).toHaveTextContent('ghe.sgdev.org/sourcegraph/gorilla-mux › go.mod')
        expect(listItems[1]).toHaveTextContent('github.com/sourcegraph/sourcegraph › .eslintrc.js')
    })

    test('files with missing data can extract it from the URL if available', () => {
        const recentFiles = {
            nodes: [
                {
                    argument: '{"filePath": ".eslintrc.js", "repoName": "github.com/sourcegraph/sourcegraph"}',
                    timestamp: '2020-09-10T22:55:18Z',
                    url: 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/.eslintrc.js',
                },
                {
                    argument: '{}',
                    timestamp: '2020-09-10T22:55:06Z',
                    url: 'https://sourcegraph.test:3443/ghe.sgdev.org/sourcegraph/gorilla-mux/-/blob/go.mod',
                },
                {
                    argument: '{}',
                    timestamp: '2020-09-10T22:55:06Z',
                    url: 'https://sourcegraph.test:3443/bitbucket.sgdev.org/SOURCEGRAPH/jsonrpc2',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 2,
        }

        const props = {
            authenticatedUser: null,
            recentFilesFragment: { recentFilesLogs: recentFiles },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        renderWithBrandedContext(<RecentFilesPanel {...props} />)
        const listItems = screen.getAllByTestId('recent-files-item')
        expect(listItems).toHaveLength(2)
        expect(listItems[0]).toHaveTextContent('github.com/sourcegraph/sourcegraph › .eslintrc.js')
        expect(listItems[1]).toHaveTextContent('ghe.sgdev.org/sourcegraph/gorilla-mux › go.mod')
    })

    test('Show More button shown when more items can be loaded', () => {
        const recentFiles = {
            nodes: [
                {
                    argument: '{"filePath": ".eslintrc.js", "repoName": "github.com/sourcegraph/sourcegraph"}',
                    timestamp: '2020-09-10T22:55:18Z',
                    url: 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/.eslintrc.js',
                },
                {
                    argument: '{}',
                    timestamp: '2020-09-10T22:55:06Z',
                    url: 'https://sourcegraph.test:3443/ghe.sgdev.org/sourcegraph/gorilla-mux/-/blob/go.mod',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: true,
            },
            totalCount: 2,
        }

        const props = {
            authenticatedUser: null,
            recentFilesFragment: { recentFilesLogs: recentFiles },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        renderWithBrandedContext(<RecentFilesPanel {...props} />)
        expect(screen.getByTestId('recent-files-panel-show-more')).toBeInTheDocument()
    })

    test('Show More button not shown when more items cannot be loaded', () => {
        const recentFiles = {
            nodes: [
                {
                    argument: '{"filePath": ".eslintrc.js", "repoName": "github.com/sourcegraph/sourcegraph"}',
                    timestamp: '2020-09-10T22:55:18Z',
                    url: 'https://sourcegraph.test:3443/github.com/sourcegraph/sourcegraph/-/blob/.eslintrc.js',
                },
                {
                    argument: '{}',
                    timestamp: '2020-09-10T22:55:06Z',
                    url: 'https://sourcegraph.test:3443/ghe.sgdev.org/sourcegraph/gorilla-mux/-/blob/go.mod',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 2,
        }

        const props = {
            authenticatedUser: null,
            recentFilesFragment: { recentFilesLogs: recentFiles },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        renderWithBrandedContext(<RecentFilesPanel {...props} />)
        expect(screen.queryByTestId('recent-files-panel-show-more')).not.toBeInTheDocument()
    })
})
