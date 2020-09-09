import React from 'react'
import { of, NEVER } from 'rxjs'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'

const { add } = storiesOf('web/search/panels/RecentSearchesPanel', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [800] },
    })
    .addDecorator(story => <div style={{ width: '800px' }}>{story()}</div>)

const emptyRecentSearches = {
    totalCount: 0,
    nodes: [],
    pageInfo: {
        endCursor: null,
        hasNextPage: false,
    },
}

const populatedRecentSearches = {
    nodes: [
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
            timestamp: '2020-09-08T17:36:52Z',
            url: 'https://sourcegraph.test:3443/search?q=test&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 5, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 1, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "^test"}}}',
            timestamp: '2020-09-08T17:26:05Z',
            url: 'https://sourcegraph.test:3443/search?q=%5Etest&patternType=regexp',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 5, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 1, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "^test"}}}',
            timestamp: '2020-09-08T17:20:11Z',
            url: 'https://sourcegraph.test:3443/search?q=%5Etest&patternType=regexp',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 5, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 1, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "^test"}}}',
            timestamp: '2020-09-08T17:20:05Z',
            url: 'https://sourcegraph.test:3443/search?q=%5Etest&patternType=regexp',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 26, "space": 2, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 3, "count_non_default": 1}, "field_lang": {"count": 1, "count_alias": 0, "count_negated": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 2, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "lang:cpp try {:[my_match]}"}}}',
            timestamp: '2020-09-08T17:12:53Z',
            url:
                'https://sourcegraph.test:3443/search?q=lang:cpp+try+%7B:%5Bmy_match%5D%7D&patternType=structural&onboardingTour=true',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 26, "space": 2, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 3, "count_non_default": 1}, "field_lang": {"count": 1, "count_alias": 0, "count_negated": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 2, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "lang:cpp try {:[my_match]}"}}}',
            timestamp: '2020-09-08T17:11:46Z',
            url:
                'https://sourcegraph.test:3443/search?q=lang:cpp+try+%7B:%5Bmy_match%5D%7D&patternType=structural&onboardingTour=true',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 86, "space": 4, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 4, "count_non_default": 3}, "field_lang": {"count": 1, "count_alias": 0, "count_negated": 0}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 1, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_type": {"count": 1, "value_diff": 0, "value_file": 0, "value_commit": 1, "value_symbol": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 1, "count_pattern": 1, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:^github\\\\.com/sourcegraph/sourcegraph$ PanelContainer lang:typescript  type:commit"}}}',
            timestamp: '2020-09-04T20:31:57Z',
            url:
                'https://sourcegraph.test:3443/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+PanelContainer+lang:typescript++type:commit&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 86, "space": 4, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 4, "count_non_default": 3}, "field_lang": {"count": 1, "count_alias": 0, "count_negated": 0}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 1, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_type": {"count": 1, "value_diff": 0, "value_file": 0, "value_commit": 1, "value_symbol": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 1, "count_pattern": 1, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:^github\\\\.com/sourcegraph/sourcegraph$ PanelContainer lang:typescript  type:commit"}}}',
            timestamp: '2020-09-04T20:27:02Z',
            url:
                'https://sourcegraph.test:3443/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+PanelContainer+lang:typescript++type:commit&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 86, "space": 4, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 4, "count_non_default": 3}, "field_lang": {"count": 1, "count_alias": 0, "count_negated": 0}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 1, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_type": {"count": 1, "value_diff": 0, "value_file": 0, "value_commit": 1, "value_symbol": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 1, "count_pattern": 1, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:^github\\\\.com/sourcegraph/sourcegraph$ PanelContainer lang:typescript  type:commit"}}}',
            timestamp: '2020-09-04T20:24:56Z',
            url:
                'https://sourcegraph.test:3443/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+PanelContainer+lang:typescript++type:commit&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 74, "space": 3, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 3, "count_non_default": 2}, "field_lang": {"count": 1, "count_alias": 0, "count_negated": 0}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 1, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 1, "count_pattern": 1, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:^github\\\\.com/sourcegraph/sourcegraph$ PanelContainer lang:typescript "}}}',
            timestamp: '2020-09-04T20:23:44Z',
            url:
                'https://sourcegraph.test:3443/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+PanelContainer+lang:typescript+&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 57, "space": 1, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 2, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 1, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 2, "count_regexp": 0, "count_literal": 1, "count_pattern": 1, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:^github\\\\.com/sourcegraph/sourcegraph$ PanelContainer"}}}',
            timestamp: '2020-09-04T20:23:38Z',
            url:
                'https://sourcegraph.test:3443/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+PanelContainer&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 43, "space": 1, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 1, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 0, "count_pattern": 1, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:^github\\\\.com/sourcegraph/sourcegraph$ "}}}',
            timestamp: '2020-09-04T20:23:30Z',
            url:
                'https://sourcegraph.test:3443/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 28, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 0, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "repo:sourcegraph/sourcegraph"}}}',
            timestamp: '2020-09-04T20:23:23Z',
            url: 'https://sourcegraph.test:3443/search?q=repo:sourcegraph/sourcegraph&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
            timestamp: '2020-09-04T20:23:09Z',
            url: 'https://sourcegraph.test:3443/search?q=test&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
            timestamp: '2020-09-04T20:23:08Z',
            url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
            timestamp: '2020-09-04T20:23:07Z',
            url: 'https://sourcegraph.test:3443/search?q=test&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 13, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 1}, "field_repo": {"count": 1, "value_glob": 0, "value_pipe": 0, "count_alias": 1, "value_regexp": 0, "count_negated": 0, "value_at_sign": 0, "value_rev_star": 0, "value_rev_caret": 0, "value_rev_colon": 0}, "field_default": {"count": 0, "count_regexp": 0, "count_literal": 0, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "r:sourcegraph"}}}',
            timestamp: '2020-09-04T20:23:06Z',
            url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
        },
        {
            argument:
                '{"mode": "plain", "code_search": {"query_data": {"empty": false, "query": {"chars": {"count": 4, "space": 0, "non_ascii": 0, "double_quote": 0, "single_quote": 0}, "fields": {"count": 1, "count_non_default": 0}, "field_default": {"count": 1, "count_regexp": 0, "count_literal": 1, "count_pattern": 0, "count_double_quote": 0, "count_single_quote": 0}}, "combined": "test"}}}',
            timestamp: '2020-09-04T20:23:06Z',
            url: 'https://sourcegraph.test:3443/search?q=test&patternType=literal',
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
            url: 'https://sourcegraph.test:3443/search?q=r:sourcegraph&patternType=literal',
        },
    ],
    pageInfo: {
        endCursor: null,
        hasNextPage: true,
    },
    totalCount: 436,
}

const props = {
    authenticatedUser: null,
    fetchRecentSearches: () => of(populatedRecentSearches),
}

add('Populated', () => <WebStory>{() => <RecentSearchesPanel {...props} />}</WebStory>)

add('Empty', () => (
    <WebStory>{() => <RecentSearchesPanel {...props} fetchRecentSearches={() => of(emptyRecentSearches)} />}</WebStory>
))

add('Loading', () => <WebStory>{() => <RecentSearchesPanel {...props} fetchRecentSearches={() => NEVER} />}</WebStory>)
