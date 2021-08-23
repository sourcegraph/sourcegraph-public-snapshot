import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CodeIntelConfigurationPolicyPage } from './CodeIntelConfigurationPolicyPage'

const policy: CodeIntelligenceConfigurationPolicyFields = {
    __typename: 'CodeIntelligenceConfigurationPolicy' as const,
    id: '1',
    name: "Eric's feature branches",
    type: GitObjectType.GIT_TREE,
    pattern: 'ef/',
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

const { add } = storiesOf('web/codeintel/configuration/CodeIntelConfigurationPolicyPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

for (const { repo, indexingEnabled } of [
    { repo: undefined, indexingEnabled: true },
    { repo: undefined, indexingEnabled: false },
    { repo: { id: '42' }, indexingEnabled: true },
    { repo: { id: '42' }, indexingEnabled: false },
]) {
    add(
        `${repo ? 'Repository' : 'Global'}ConfigurationPolicyIndexing${indexingEnabled ? 'Enabled' : 'Disabled'}`,
        () => (
            <EnterpriseWebStory>
                {props => (
                    <CodeIntelConfigurationPolicyPage
                        {...props}
                        repo={repo}
                        indexingEnabled={indexingEnabled}
                        getPolicyById={() => of(policy)}
                        repoName={() => of(repoResult)}
                        searchGitBranches={() => of(branchesResult)}
                        searchGitTags={() => of(tagsResult)}
                        updatePolicy={() => of()}
                    />
                )}
            </EnterpriseWebStory>
        )
    )
}
