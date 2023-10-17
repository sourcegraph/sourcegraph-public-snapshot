import { gql } from '@sourcegraph/http-client'

export const SITE_CONFIG_QUERY = gql`
    query SiteConfig {
        site {
            configuration {
                id
                effectiveContents
                licenseInfo {
                    tags
                    userCount
                    expiresAt
                }
            }
        }

        externalServices {
            nodes {
                id
                displayName
            }
        }
    }
`

export const LICENSE_KEY_MUTATION = gql`
    mutation UpdateSiteConfiguration($lastID: Int!, $input: String!) {
        updateSiteConfiguration(lastID: $lastID, input: $input)
    }
`
