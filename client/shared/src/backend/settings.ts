import { gql } from '@sourcegraph/http-client'

const settingsCascadeFragment = gql`
    fragment SettingsCascadeFields on SettingsCascade {
        subjects {
            __typename
            ...OrgSettingFields
            ...UserSettingFields
            ...SiteSettingFields
            ...DefaultSettingFields
        }
        final
    }

    fragment OrgSettingFields on Org {
        __typename
        latestSettings {
            id
            contents
        }
        id
        settingsURL
        viewerCanAdminister

        name
        displayName
    }

    fragment UserSettingFields on User {
        __typename
        latestSettings {
            id
            contents
        }
        id
        settingsURL
        viewerCanAdminister

        username
        displayName
    }

    fragment SiteSettingFields on Site {
        __typename
        latestSettings {
            id
            contents
        }
        id
        settingsURL
        viewerCanAdminister

        siteID
        allowSiteSettingsEdits
    }

    fragment DefaultSettingFields on DefaultSettings {
        __typename
        latestSettings {
            id
            contents
        }
        id
        settingsURL
        viewerCanAdminister
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
