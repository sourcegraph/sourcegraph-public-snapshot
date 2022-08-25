import { gql } from '@sourcegraph/http-client'

export const GET_SOURCEGRAPH_VERSION = gql`
    query GetSourcegraphVersion {
        site {
            productVersion
        }
    }
`
