import { boolean, withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import {
    CodeIntelConfigurationPolicyPage,
    CodeIntelConfigurationPolicyPageProps,
} from './CodeIntelConfigurationPolicyPage'

const policy: CodeIntelligenceConfigurationPolicyFields = {
    __typename: 'CodeIntelligenceConfigurationPolicy' as const,
    id: '1',
    name: "Eric's feature branches",
    type: GitObjectType.GIT_TREE,
    pattern: 'ef/',
    protected: false,
    retentionEnabled: true,
    retentionDurationHours: 168,
    retainIntermediateCommits: true,
    indexingEnabled: true,
    indexCommitMaxAgeHours: 672,
    indexIntermediateCommits: true,
}

const repoResult = {
    __typename: 'Repository' as const,
    name: 'github.com/sourcegraph/sourcegraph',
}

const branchesResult = {
    ...repoResult,
    branches: {
        totalCount: 3,
        nodes: [{ displayName: 'ef/wip1' }, { displayName: 'ef/wip2' }, { displayName: 'ef/wip3' }],
    },
}

const tagsResult = {
    ...repoResult,
    tags: { totalCount: 2, nodes: [{ displayName: 'v3.0-ref' }, { displayName: 'v3-ref.1' }] },
}

const story: Meta = {
    title: 'web/codeintel/configuration/CodeIntelConfigurationPolicyPage',
    decorators: [story => <div className="p-3 container">{story()}</div>, withKnobs],
    parameters: {
        component: CodeIntelConfigurationPolicyPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelConfigurationPolicyPageProps> = args => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelConfigurationPolicyPage {...props} indexingEnabled={boolean('indexingEnabled', true)} {...args} />
        )}
    </EnterpriseWebStory>
)

const defaults: Partial<CodeIntelConfigurationPolicyPageProps> = {
    getPolicyById: () => of(policy),
    updatePolicy: () => of(),
    searchGitBranches: () => of(branchesResult),
    searchGitTags: () => of(tagsResult),
    repoName: () => of(repoResult),
}

export const GlobalPage = Template.bind({})
GlobalPage.args = {
    ...defaults,
}

export const GlobalProtectedPage = Template.bind({})
GlobalProtectedPage.args = {
    ...defaults,
    getPolicyById: () => of({ ...policy, protected: true }),
}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: '42' },
}
