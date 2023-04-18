import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'

import { ENVIRONMENT_CONFIG } from './constants'

const { SOURCEGRAPH_ACCESS_TOKEN, SOURCEGRAPH_SERVER_ENDPOINT } = ENVIRONMENT_CONFIG

export const sourcegraphClient = new SourcegraphGraphQLAPIClient(SOURCEGRAPH_SERVER_ENDPOINT, SOURCEGRAPH_ACCESS_TOKEN)
