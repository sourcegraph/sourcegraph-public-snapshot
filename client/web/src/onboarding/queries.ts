import { gql } from '@sourcegraph/http-client'

export const SITE_CONFIG_QUERY = gql`
    query SiteConfig {
        site {
            configuration {
                effectiveContents
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
