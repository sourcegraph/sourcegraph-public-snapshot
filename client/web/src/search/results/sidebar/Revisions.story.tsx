import { MockedProviderProps } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'
import React from 'react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'

import { Revisions, RevisionsProps, TabIndex } from './Revisions'
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
import sidebarStyles from './SearchSidebar.module.scss'

export default {
    title: 'web/search/results/sidebar/Revisions',
    component: Revisions,
    decorators: [
        Story => (
            <div className={sidebarStyles.searchSidebar}>
                <Story />
            </div>
        ),
    ],
    argTypes: { onFilterClick: { action: 'onFilterClick' } },
} as Meta

const Template: Story<RevisionsProps & Partial<Pick<MockedProviderProps, 'mocks'>>> = ({ mocks, ...props }) => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={mocks}>
                <Revisions {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const EmptyBranches = Template.bind({})
EmptyBranches.args = {
    ...MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: EMPTY_MOCKS,
}
export const EmptyTags = Template.bind({})
EmptyTags.args = {
    ...EmptyBranches.args,
    _initialTab: TabIndex.TAGS,
}

export const FewResultsBranches = Template.bind({})
FewResultsBranches.args = {
    ...MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: FEW_RESULTS_MOCKS,
}
export const FewResultsTags = Template.bind({})
FewResultsTags.args = {
    ...FewResultsBranches.args,
    _initialTab: TabIndex.TAGS,
}

export const ManyResultsBranches = Template.bind({})
ManyResultsBranches.args = {
    ...MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: DEFAULT_MOCKS,
}

export const ManyResultsTags = Template.bind({})
ManyResultsTags.args = {
    ...ManyResultsBranches.args,
    _initialTab: TabIndex.TAGS,
}

export const SearchBranches = Template.bind({})
SearchBranches.args = {
    ...FILTERED_MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: FILTERED_MOCKS,
}
export const SearchTags = Template.bind({})
SearchTags.args = {
    ...SearchBranches.args,
    _initialTab: TabIndex.TAGS,
}

export const EmptySearchBranches = Template.bind({})
EmptySearchBranches.args = {
    ...FILTERED_MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: EMPTY_FILTERED_MOCKS,
}
export const EmptySearchTags = Template.bind({})
EmptySearchTags.args = {
    ...EmptySearchBranches.args,
    _initialTab: TabIndex.TAGS,
}

export const NetworkErrorBranches = Template.bind({})
NetworkErrorBranches.args = {
    ...MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: NETWORK_ERROR_MOCKS,
}
export const NetworkErrorTags = Template.bind({})
NetworkErrorTags.args = {
    ...NetworkErrorBranches.args,
    _initialTab: TabIndex.TAGS,
}

export const GraphqlErrorBranches = Template.bind({})
GraphqlErrorBranches.args = {
    ...MOCK_PROPS,
    _initialTab: TabIndex.BRANCHES,
    mocks: GRAPHQL_ERROR_MOCKS,
}
export const GraphqlErrorTags = Template.bind({})
GraphqlErrorTags.args = {
    ...GraphqlErrorBranches.args,
    _initialTab: TabIndex.TAGS,
}
