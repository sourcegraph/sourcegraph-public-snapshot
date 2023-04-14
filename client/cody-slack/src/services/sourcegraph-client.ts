import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'

import { DEFAULT_APP_SETTINGS, ENVIRONMENT_CONFIG } from '../constants'

const { SOURCEGRAPH_ACCESS_TOKEN } = ENVIRONMENT_CONFIG

export const sourcegraphClient = new SourcegraphGraphQLAPIClient(
    DEFAULT_APP_SETTINGS.serverEndpoint,
    SOURCEGRAPH_ACCESS_TOKEN
)

export const intentDetector = new SourcegraphIntentDetectorClient(sourcegraphClient)
