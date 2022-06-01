import { MockedProviderProps } from '@apollo/client/testing'
import { Meta } from '@storybook/react'

import { RevisionsProps, TabIndex } from '@sourcegraph/search-ui'
// eslint-disable-next-line no-restricted-imports
import sidebarStyles from '@sourcegraph/search-ui/src/results/sidebar/SearchSidebar.module.scss'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'

import { Revisions } from './Revisions'
import {
    EMPTY_MOCKS,
    FEW_RESULTS_MOCKS,
    FILTERED_MOCKS,
    MOCK_PROPS,
    FILTERED_MOCK_PROPS,
    DEFAULT_MOCKS,
    EMPTY_FILTERED_MOCKS,
    NETWORK_ERROR_MOCKS,
    GRAPHQL_ERROR_MOCKS,
} from './Revisions.mocks'

export default {
    title: 'web/search/results/sidebar/Revisions',
    component: Revisions,
    argTypes: { onFilterClick: { action: 'onFilterClick' } },
    parameters: {
        chromatic: { disableSnapshot: false },
    },
} as Meta

const examples: (RevisionsProps & Partial<Pick<MockedProviderProps, 'mocks'>> & { title: string })[] = [
    TabIndex.BRANCHES,
    TabIndex.TAGS,
]
    .map(_initialTab => {
        const tabName = _initialTab === TabIndex.BRANCHES ? 'branches' : 'tags'
        return [
            {
                title: `Empty ${tabName}`,
                ...MOCK_PROPS,
                _initialTab,
                mocks: EMPTY_MOCKS,
            },
            {
                title: `Few results ${tabName}`,
                ...MOCK_PROPS,
                _initialTab,
                mocks: FEW_RESULTS_MOCKS,
            },
            {
                title: `Many results ${tabName}`,
                ...MOCK_PROPS,
                _initialTab,
                mocks: DEFAULT_MOCKS,
            },
            {
                title: `Search ${tabName}`,
                ...FILTERED_MOCK_PROPS,
                _initialTab,
                mocks: FILTERED_MOCKS,
            },
            {
                title: `Empty search ${tabName}`,
                ...FILTERED_MOCK_PROPS,
                _initialTab,
                mocks: EMPTY_FILTERED_MOCKS,
            },
            {
                title: `Network error ${tabName}`,
                ...MOCK_PROPS,
                _initialTab,
                mocks: NETWORK_ERROR_MOCKS,
            },
            {
                title: `Network error ${tabName}`,
                ...MOCK_PROPS,
                _initialTab,
                mocks: GRAPHQL_ERROR_MOCKS,
            },
        ]
    })
    .flat()

export function RevisionsSection() {
    return (
        <WebStory>
            {() => (
                <>
                    {examples.map(({ mocks, title, ...props }) => (
                        <div
                            key={title}
                            style={{ border: '1px solid #AAA', borderRadius: '3px', padding: '1rem', margin: '1rem' }}
                        >
                            <Typography.H2>{title}</Typography.H2>
                            <div className={sidebarStyles.searchSidebar}>
                                <MockedTestProvider mocks={mocks}>
                                    <Revisions {...props} />
                                </MockedTestProvider>
                            </div>
                        </div>
                    ))}
                </>
            )}
        </WebStory>
    )
}
