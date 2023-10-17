import type { Meta, StoryFn, Decorator } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'

import { REPO_CHANGESETS_STATS } from './backend'
import { RepoBatchChangesButton } from './RepoBatchChangesButton'

const decorator: Decorator = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/batches/repo',
    decorators: [decorator],
}

export default config
let openValue = 0
let mergedValue = 0

export const RepoButton: StoryFn = args => (
    <WebStory>
        {() => {
            openValue = args.open
            mergedValue = args.merged

            return (
                <MockedTestProvider
                    link={
                        new WildcardMockLink([
                            {
                                request: {
                                    query: getDocumentNode(REPO_CHANGESETS_STATS),
                                    variables: MATCH_ANY_PARAMETERS,
                                },
                                result: {
                                    data: {
                                        repository: {
                                            __typename: 'Repository',
                                            changesetsStats: { open: openValue, merged: mergedValue },
                                        },
                                    },
                                },
                                nMatches: Number.POSITIVE_INFINITY,
                            },
                        ])
                    }
                >
                    <RepoBatchChangesButton repoName="Awesome Repo" />
                </MockedTestProvider>
            )
        }}
    </WebStory>
)
RepoButton.argTypes = {
    open: {
        control: { type: 'number' },
    },
    merged: {
        control: { type: 'number' },
    },
}
RepoButton.args = {
    open: 2,
    merged: 47,
}

RepoButton.storyName = 'RepoButton'
