import { gql } from '@sourcegraph/http-client'

export const AUTHZ_PROVIDERS = gql`
    query AuthzProviders {
        authzProviders {
            serviceID
            serviceType
        }
    }
`
