import { boolean, withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import React from 'react'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'

import { WebStory } from '../../../components/WebStory'
import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import {
    CodeIntelConfigurationPolicyPage,
    CodeIntelConfigurationPolicyPageProps,
} from './CodeIntelConfigurationPolicyPage'
import { POLICY_CONFIGURATION_BY_ID } from './usePoliciesConfigurations'
import { PREVIEW_GIT_OBJECT_FILTER } from './useSearchGit'

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

const policyRequest = {
    request: {
        query: getDocumentNode(POLICY_CONFIGURATION_BY_ID),
    },
    result: {
        data: {
            node: {
                ...policy,
            },
        },
    },
}

const branchRequest = {
    request: {
        query: getDocumentNode(PREVIEW_GIT_OBJECT_FILTER),
    },
    result: {
        data: {
            node: {
                ...repoResult,
                previewGitObjectFilter: [
                    { name: 'ef/wip1', rev: 'deadbeef01' },
                    { name: 'ef/wip2', rev: 'deadbeef02' },
                    { name: 'ef/wip3', rev: 'deadbeef03' },
                ],
            },
        },
    },
}

const tagRequest = {
    request: {
        query: getDocumentNode(PREVIEW_GIT_OBJECT_FILTER),
    },
    result: {
        data: {
            node: {
                ...repoResult,
                previewGitObjectFilter: [
                    { name: 'v3.0-ref', rev: 'deadbeef04' },
                    { name: 'v3-ref.1', rev: 'deadbeef05' },
                ],
            },
        },
    },
}

const commitRequest = {
    request: {
        query: getDocumentNode(PREVIEW_GIT_OBJECT_FILTER),
    },
    result: {
        data: {
            node: {
                ...repoResult,
                previewGitObjectFilter: [],
            },
        },
    },
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
    <WebStory mocks={[policyRequest, branchRequest, branchRequest, tagRequest, tagRequest, commitRequest]}>
        {props => (
            <CodeIntelConfigurationPolicyPage {...props} indexingEnabled={boolean('indexingEnabled', true)} {...args} />
        )}
    </WebStory>
)

const defaults: Partial<CodeIntelConfigurationPolicyPageProps> = {}

export const GlobalPage = Template.bind({})
GlobalPage.args = {
    ...defaults,
}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: '42' },
}
