import { gql } from '@sourcegraph/http-client'

const settingsCascadeFragment = gql`
    fragment SettingsCascadeFields on SettingsCascade {
        subjects {
            __typename
            ... on Org {
                name
                displayName
            }
            ... on User {
                username
                displayName
            }
            ... on Site {
                siteID
                allowSiteSettingsEdits
            }
            latestSettings {
                id
                contents
            }
            id
            settingsURL
            viewerCanAdminister
        }
        final
    }
`

export const viewerSettingsQuery = gql`
    query ViewerSettings {
        viewerSettings {
            ...SettingsCascadeFields
        }
    }
    ${settingsCascadeFragment}
`
