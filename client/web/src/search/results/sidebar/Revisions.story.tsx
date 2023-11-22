import type { MockedProviderProps } from '@apollo/client/testing'
import type { Meta } from '@storybook/react'

import { type RevisionsProps, TabIndex } from '@sourcegraph/branded'
import sidebarStyles from '@sourcegraph/branded/src/search-ui/results/sidebar/SearchSidebar.module.scss'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { H2 } from '@sourcegraph/wildcard'

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

const meta: Meta = {
    title: 'web/search/results/sidebar/Revisions',
    component: Revisions,
    argTypes: { onFilterClick: { action: 'onFilterClick' } },
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default meta

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
                            <H2>{title}</H2>
                            <div className={sidebarStyles.sidebar}>
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
