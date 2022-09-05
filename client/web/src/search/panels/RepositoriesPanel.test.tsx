import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { noop } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { RepositoriesPanel } from './RepositoriesPanel'

describe('RepositoriesPanel', () => {
    test('Both r: and repo: filters are tracked', () => {
        const recentSearches = {
            nodes: [
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:39Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:test"}}}',
                    timestamp: '2020-09-04T18:44:30Z',
                    url: 'https://sourcegraph.test:3443/search?q=repo:test&patternType=literal',
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
            telemetryService: NOOP_TELEMETRY_SERVICE,
            recentlySearchedRepositories: { recentlySearchedRepositoriesLogs: recentSearches },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
        }

        expect(renderWithBrandedContext(<RepositoriesPanel {...props} />).asFragment()).toMatchSnapshot()
    })

    test('consecutive searches with identical repo filters are correctly merged when rendered', () => {
        const recentSearches = {
            nodes: [
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
                    timestamp: '2020-09-08T17:36:52Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:39Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph+test&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:30Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
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
            recentlySearchedRepositories: { recentlySearchedRepositoriesLogs: recentSearches },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        expect(renderWithBrandedContext(<RepositoriesPanel {...props} />).asFragment()).toMatchSnapshot()
    })

    test('Show More button is shown if more pages are available', () => {
        const recentSearches = {
            nodes: [
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
                    timestamp: '2020-09-08T17:36:52Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:39Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:30Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test-two&patternType=literal',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: true,
            },
            totalCount: 6,
        }

        const props = {
            authenticatedUser: null,
            recentlySearchedRepositories: { recentlySearchedRepositoriesLogs: recentSearches },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: noop as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        expect(renderWithBrandedContext(<RepositoriesPanel {...props} />).asFragment()).toMatchSnapshot()
    })

    test('Show More button loads more items', () => {
        const recentSearches1 = {
            nodes: [
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
                    timestamp: '2020-09-08T17:36:52Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:39Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:30Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test-two&patternType=literal',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: true,
            },
            totalCount: 6,
        }

        const recentSearches2 = {
            nodes: [
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
                    timestamp: '2020-09-08T17:36:52Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:39Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:30Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test-two&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
                    timestamp: '2020-09-08T17:36:52Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test-three&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:39Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:r:test-four&patternType=literal',
                },
                {
                    argument:
                        '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
                    timestamp: '2020-09-04T18:44:30Z',
                    url: 'https://sourcegraph.test:3443/search?q=r:test-five&patternType=literal',
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 6,
        }

        const props = {
            className: '',
            authenticatedUser: null,
            recentlySearchedRepositories: { recentlySearchedRepositoriesLogs: recentSearches1 },
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
            fetchMore: (() => ({ recentSearchesLogs: recentSearches2 })) as any,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        const { asFragment } = renderWithBrandedContext(<RepositoriesPanel {...props} />)
        userEvent.click(screen.getByRole('button', { name: /Show more/ }))
        expect(asFragment()).toMatchSnapshot()
    })
})
