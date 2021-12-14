import { gql } from '../graphql/graphql'

const settingsCascadeFragment = gql`
    fragment SettingsCascadeFields on SettingsCascade {
        subjects {
            __typename
            ... on Org {
                id
                name
                displayName
            }
            ... on User {
                id
                username
                displayName
            }
            ... on Site {
                id
                siteID
                allowSiteSettingsEdits
            }
            latestSettings {
                id
                contents
            }
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
