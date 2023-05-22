import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'

import { DEFAULT_APP_SETTINGS, ENVIRONMENT_CONFIG } from '../constants'

const { SOURCEGRAPH_ACCESS_TOKEN } = ENVIRONMENT_CONFIG

export const sourcegraphClient = new SourcegraphGraphQLAPIClient({
    serverEndpoint: DEFAULT_APP_SETTINGS.serverEndpoint,
    accessToken: SOURCEGRAPH_ACCESS_TOKEN,
    customHeaders: {},
})

export const intentDetector = new SourcegraphIntentDetectorClient(sourcegraphClient)
