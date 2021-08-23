import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CodeIntelConfigurationPage } from './CodeIntelConfigurationPage'

const trim = (value: string) => {
    const firstSignificantLine = value
        .split('\n')
        .map(line => ({ length: line.length, trimmedLength: line.trimStart().length }))
        .find(({ trimmedLength }) => trimmedLength !== 0)
    if (!firstSignificantLine) {
        return value
    }

    const { length, trimmedLength } = firstSignificantLine
    return value
        .split('\n')
        .map(line => line.slice(length - trimmedLength))
        .join('\n')
        .trim()
}

const globalPolicies: CodeIntelligenceConfigurationPolicyFields[] = [
    {
        __typename: 'CodeIntelligenceConfigurationPolicy' as const,
        id: 'g1',
        name: 'Default major release retention',
        type: GitObjectType.GIT_TAG,
        pattern: '.0.0',
        retentionEnabled: true,
        retentionDurationHours: 168,
        retainIntermediateCommits: false,
        indexingEnabled: false,
        indexCommitMaxAgeHours: 672,
        indexIntermediateCommits: false,
    },
    {
        __typename: 'CodeIntelligenceConfigurationPolicy' as const,
        id: 'g2',
        name: 'Default brach retention',
        type: GitObjectType.GIT_TREE,
        pattern: '',
        retentionEnabled: true,
        retentionDurationHours: 2016,
        retainIntermediateCommits: false,
        indexingEnabled: false,
        indexCommitMaxAgeHours: 4032,
        indexIntermediateCommits: false,
    },
]

const policies: CodeIntelligenceConfigurationPolicyFields[] = [
    {
        __typename: 'CodeIntelligenceConfigurationPolicy' as const,
        id: 'id1',
        name: 'All branches created by Eric',
        type: GitObjectType.GIT_TREE,
        pattern: 'ef/',
        retentionEnabled: true,
        retentionDurationHours: 8064,
        retainIntermediateCommits: true,
        indexingEnabled: true,
        indexCommitMaxAgeHours: 40320,
        indexIntermediateCommits: true,
    },
    {
        __typename: 'CodeIntelligenceConfigurationPolicy' as const,
        id: 'id2',
        name: 'All branches created by Erik',
        type: GitObjectType.GIT_TREE,
        pattern: 'es/',
        retentionEnabled: true,
        retentionDurationHours: 8064,
        retainIntermediateCommits: true,
        indexingEnabled: true,
        indexCommitMaxAgeHours: 40320,
        indexIntermediateCommits: true,
    },
]

const repositoryConfiguration = {
    __typename: 'Repository' as const,
    indexConfiguration: {
        configuration: trim(`
            {
                "shared_steps": [],
                "index_jobs": [
                    {
                        "steps": [
                            {
                                "root": "",
                                "image": "sourcegraph/lsif-node:autoindex",
                                "commands": [
                                    "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto",
                                    "yarn --ignore-engines"
                                ]
                            },
                            {
                                "root": "client/web",
                                "image": "sourcegraph/lsif-node:autoindex",
                                "commands": [
                                    "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto",
                                    "npm install"
                                ]
                            }
                        ],
                        "local_steps": [
                            "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"
                        ],
                        "root": "client/web",
                        "indexer": "sourcegraph/lsif-node:autoindex",
                        "indexer_args": [
                            "lsif-tsc",
                            "-p",
                            "."
                        ],
                        "outfile": ""
                    }
                ]
            }
        `),
    },
}

const inferredRepositoryConfiguration = {
    __typename: 'Repository' as const,
    indexConfiguration: {
        inferredConfiguration: trim(`
            {
                "shared_steps": [],
                "index_jobs": [
                    {
                        "steps": [
                            {
                                "root": "lib",
                                "image": "sourcegraph/lsif-go:latest",
                                "commands": [
                                    "go mod download"
                                ]
                            }
                        ],
                        "local_steps": [],
                        "root": "lib",
                        "indexer": "sourcegraph/lsif-go:latest",
                        "indexer_args": [
                            "lsif-go",
                            "--no-animation"
                        ],
                        "outfile": ""
                    }
                ]
            }
        `),
    },
}

const { add } = storiesOf('web/codeintel/configuration/CodeIntelConfigurationPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Empty', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelConfigurationPage
                {...props}
                repo={{ id: '42' }}
                getPolicies={() => of([])}
                getConfigurationForRepository={() => of(repositoryConfiguration)}
                getInferredConfigurationForRepository={() => of(inferredRepositoryConfiguration)}
                deletePolicyById={() => of()}
                updateConfigurationForRepository={() => of()}
            />
        )}
    </EnterpriseWebStory>
))

for (const { repo, indexingEnabled } of [
    { repo: undefined, indexingEnabled: true },
    { repo: undefined, indexingEnabled: false },
    { repo: { id: '42' }, indexingEnabled: true },
    { repo: { id: '42' }, indexingEnabled: false },
]) {
    add(`${repo ? 'Repository' : 'Global'}ConfigurationIndexing${indexingEnabled ? 'Enabled' : 'Disabled'}`, () => (
        <EnterpriseWebStory>
            {props => (
                <CodeIntelConfigurationPage
                    {...props}
                    repo={repo}
                    indexingEnabled={indexingEnabled}
                    getPolicies={(repositoryId?: string) => (repositoryId ? of(policies) : of(globalPolicies))}
                    getConfigurationForRepository={() => of(repositoryConfiguration)}
                    getInferredConfigurationForRepository={() => of(inferredRepositoryConfiguration)}
                    deletePolicyById={() => of()}
                    updateConfigurationForRepository={() => of()}
                />
            )}
        </EnterpriseWebStory>
    ))
}
