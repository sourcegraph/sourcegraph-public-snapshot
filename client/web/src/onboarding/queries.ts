import { gql } from '@sourcegraph/http-client'

export const SITE_CONFIG_QUERY = gql`
    query SiteConfig {
        site {
            configuration {
                id
                effectiveContents
            }
        }

        externalServices {
            nodes {
                id
                displayName
                unrestrictedAccess
            }
        }
    }
`

export const LICENSE_KEY_MUTATION = gql`
    mutation UpdateSiteConfiguration($lastID: Int!, $input: String!) {
        updateSiteConfiguration(lastID: $lastID, input: $input)
    }
`
