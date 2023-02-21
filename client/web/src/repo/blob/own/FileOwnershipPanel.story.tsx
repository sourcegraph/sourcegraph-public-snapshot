import { MockedProvider, MockedResponse } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'

import { WebStory } from '../../../components/WebStory'
import { FetchOwnershipResult } from '../../../graphql-operations'

import { FETCH_OWNERS, FileOwnershipPanel } from './FileOwnershipPanel'

const response: FetchOwnershipResult = {
    node: {
        __typename: 'Repository',
        commit: {
            blob: {
                ownership: {
                    nodes: [
                        {
                            __typename: 'Ownership',
                            owner: {
                                __typename: 'Person',
                                email: 'alice@example.com',
                                avatarURL: null,
                                displayName: 'Alice',
                                user: null,
                            },
                            reasons: [
                                {
                                    __typename: 'CodeownersFileEntry',
                                    title: 'CodeOwner',
                                    description: 'This person is listed in the CODEOWNERS file',
                                },
                            ],
                        },
                        {
                            __typename: 'Ownership',
                            owner: {
                                __typename: 'Person',
                                email: '',
                                avatarURL: null,
                                displayName: 'Bob',
                                user: null,
                            },
                            reasons: [
                                {
                                    __typename: 'CodeownersFileEntry',
                                    title: 'CodeOwner',
                                    description: 'This person is listed in the CODEOWNERS file',
                                },
                            ],
                        },
                    ],
                },
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
        },
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
