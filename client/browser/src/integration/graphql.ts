import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { sharedGraphQlResults } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import { BrowserGraphQlOperations } from '../graphql-operations'

/**
 * Predefined results for GraphQL requests that are made on almost every page.
 */
export const commonBrowserGraphQlResults: Partial<BrowserGraphQlOperations & SharedGraphQlOperations> = {
    ...sharedGraphQlResults,
    logEvent: () => ({
        logEvent: {
            alwaysNil: null,
        },
    }),
}
