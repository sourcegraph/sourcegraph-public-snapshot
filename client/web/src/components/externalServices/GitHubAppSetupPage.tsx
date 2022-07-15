import React from 'react'

import { gql } from '@sourcegraph/http-client'
import { PageHeader, Link, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../PageTitle'

import { EditGitHubAppForm } from './EditGitHubAppForm'

export const EditUserProfilePageGQLFragment = gql`
    fragment EditUserProfilePage on User {
        id
        username
        displayName
        avatarURL
        viewerCanChangeUsername
        createdAt
    }
`

export const GitHubAppSetupPage: React.FunctionComponent = ({ ...props }) => (
    <div>
        <PageTitle title="Configure GitHub App" />
        <PageHeader path={[{ text: 'GitHub App' }]} headingElement="h2" />
        <EditGitHubAppForm
            after={
                window.context.sourcegraphDotComMode && (
                    <Text className="mt-4">
                        <Link to="https://about.sourcegraph.com/contact">Contact support</Link> to delete your account.
                    </Text>
                )
            }
        />
    </div>
)
