import { MockedProvider, MockedResponse } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'

import { WebStory } from '../../../components/WebStory'
import { FetchOwnershipResult, FetchOwnershipVariables } from '../../../graphql-operations'

import { FETCH_OWNERS, FileOwnershipPanel } from './FileOwnershipPanel'

const response: FetchOwnershipResult = {
    node: {
        __typename: 'Repository',
        commit: {
            blob: {
                ownership: [
                    {
                        __typename: 'Ownership',
                        handle: 'alice',
                        person: {
                            __typename: 'Person',
                            email: 'alice@example.com',
                            avatarURL: null,
                            displayName: 'Alice',
                            user: null,
                        },
                        reasons: [
                            {
                                __typename: 'CodeownersFileEntry',
                                title: 'Codeowner',
                                description: 'This person is listed in the CODEOWNERS file',
                            },
                            {
                                __typename: 'RecentContributor',
                                title: 'Contributor',
                                description: 'This person has recently contributed to this file',
                            },
                        ],
                    },
                    {
                        __typename: 'Ownership',
                        handle: 'bob',
                        person: {
                            __typename: 'Person',
                            email: '',
                            avatarURL: null,
                            displayName: 'Bob',
                            user: null,
                        },
                        reasons: [
                            {
                                __typename: 'RecentContributor',
                                title: 'Contributor',
                                description: 'This person has recently contributed to this file',
                            },
                        ],
                    },
                ],
            },
        },
    },
}

const mockResponse: MockedResponse<FetchOwnershipResult> = {
    request: {
        query: getDocumentNode(FETCH_OWNERS),
        variables: {
            repo: 'github.com/sourcegraph/sourcegraph',
            currentPath: 'README.md',
            revision: '',
        } as FetchOwnershipVariables,
    },
    result: {
        data: response,
    },
}

const config: Meta = {
    title: 'web/repo/blob/FileOwnership',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: Story = () => (
    <WebStory>
        {() => (
            <MockedProvider mocks={[mockResponse]}>
                <FileOwnershipPanel repoID="github.com/sourcegraph/sourcegraph" filePath="README.md" />
            </MockedProvider>
        )}
    </WebStory>
)
