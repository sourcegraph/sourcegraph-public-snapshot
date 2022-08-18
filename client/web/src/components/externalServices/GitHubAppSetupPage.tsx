import React, { useCallback } from 'react'

import { gql, useMutation, useQuery } from '@apollo/client'

import { PageHeader, Link, Text, LoadingSpinner } from '@sourcegraph/wildcard'

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

const GET_SITE_CONFIG = gql`
    query GetSite {
        site {
            configuration {
                id
                effectiveContents
            }
        }
    }
`

const SET_SITE_CONFIG = gql`
    mutation UpdateSiteConfiguration($lastID: Int!, $input: String!) {
        updateSiteConfiguration(lastID: $lastID, input: $input)
    }
`

export const GitHubAppSetupPage: React.FunctionComponent = ({ ...props }) => {
    const { loading, error, data } = useQuery(GET_SITE_CONFIG)
    const [updateSettings] = useMutation(SET_SITE_CONFIG)

    let config = {}
    if (!loading) {
        config = JSON.parse(data.site.configuration.effectiveContents)
    }

    const doSettingsUpdate = useCallback(
        async updatedConfig => {
            try {
                await updateSettings({
                    variables: { lastID: parseInt(data.site.configuration.id), input: JSON.stringify(updatedConfig) },
                })
            } catch {
                console.log('SettingsUpdateFailed')
            }
        },
        [updateSettings, config, data]
    )

    return (
        <div>
            <PageTitle title="Configure GitHub App" />
            <PageHeader path={[{ text: 'GitHub App' }]} headingElement="h2" />
            {loading && <LoadingSpinner />}
            {!loading && data && (
                <EditGitHubAppForm
                    value={config}
                    initialValue={config}
                    doUpdate={doSettingsUpdate}
                    after={
                        window.context.sourcegraphDotComMode && (
                            <Text className="mt-4">
                                <Link to="https://about.sourcegraph.com/contact">Contact support</Link> to delete your
                                account.
                            </Text>
                        )
                    }
                />
            )}
        </div>
    )
}
