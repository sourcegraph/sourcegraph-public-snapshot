import { boolean, withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../components/WebStory'
import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { POLICY_CONFIGURATION_BY_ID } from '../hooks/usePolicyConfigurationById'
import { PREVIEW_GIT_OBJECT_FILTER } from '../hooks/usePreviewGitObjectFilter'
import { PREVIEW_REPOSITORY_FILTER } from '../hooks/usePreviewRepositoryFilter'

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
    repositoryPatterns: [],
    protected: false,
    retentionEnabled: true,
    retentionDurationHours: 168,
    retainIntermediateCommits: true,
    indexingEnabled: true,
    indexCommitMaxAgeHours: 672,
    indexIntermediateCommits: true,
    repository: null,
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

const previewRepositoryFilterRequest = {
    request: {
        query: getDocumentNode(PREVIEW_REPOSITORY_FILTER),
        variables: {
            pattern: 'github.com/sourcegraph/sourcegraph',
        },
    },
    result: {
        data: {
            previewRepositoryFilter: {
                nodes: [
                    {
                        name: 'github.com/sourcegraph/sourcegraph',
                    },
                    {
                        name: '*',
                    },
                ],
                totalCount: 3,
                totalMatches: 2,
                limit: null,
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
    <WebStory
        mocks={[
            policyRequest,
            branchRequest,
            branchRequest,
            tagRequest,
            tagRequest,
            commitRequest,
            previewRepositoryFilterRequest,
            previewRepositoryFilterRequest,
        ]}
    >
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
RepositoryPage.parameters = {
    // Keep snapshots for one variant
    chromatic: { disableSnapshots: false },
}
