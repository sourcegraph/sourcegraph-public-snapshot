import { boolean, withKnobs } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CodeIntelConfigurationPage, CodeIntelConfigurationPageProps } from './CodeIntelConfigurationPage'

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
        protected: true,
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
        protected: false,
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
        protected: false,
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
        protected: false,
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

const story: Meta = {
    title: 'web/codeintel/configuration/CodeIntelConfigurationPage',
    decorators: [story => <div className="p-3 container">{story()}</div>, withKnobs],
    parameters: {
        component: CodeIntelConfigurationPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelConfigurationPageProps> = args => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelConfigurationPage {...props} indexingEnabled={boolean('indexingEnabled', true)} {...args} />
        )}
    </EnterpriseWebStory>
)

const defaults: Partial<CodeIntelConfigurationPageProps> = {
    getPolicies: () => of([]),
    getConfigurationForRepository: () => of(repositoryConfiguration),
    getInferredConfigurationForRepository: () => of(inferredRepositoryConfiguration),
    updateConfigurationForRepository: () => of(),
    deletePolicyById: () => of(),
}

export const EmptyGlobalPage = Template.bind({})
EmptyGlobalPage.args = {
    ...defaults,
}

export const GlobalPage = Template.bind({})
GlobalPage.args = {
    ...defaults,
    getPolicies: (repositoryId?: string) => of(repositoryId ? policies : globalPolicies),
}

export const EmptyRepositoryPage = Template.bind({})
EmptyRepositoryPage.args = {
    ...defaults,
    repo: { id: 'sourcegraph' },
}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: 'sourcegraph' },
    getPolicies: (repositoryId?: string) => of(repositoryId ? policies : globalPolicies),
}
