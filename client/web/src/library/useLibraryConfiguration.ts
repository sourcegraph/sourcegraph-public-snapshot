import { logger } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'

import { type LibraryConfigurationResult } from '../graphql-operations'

type LibraryConfiguration = Pick<LibraryConfigurationResult, 'viewerCanChangeLibraryItemVisibilityToPublic'>

const DEFAULT_LIBRARY_CONFIGURATION: LibraryConfiguration = {
    viewerCanChangeLibraryItemVisibilityToPublic: false,
}

const libraryConfigurationQuery = gql`
    query LibraryConfiguration {
        viewerCanChangeLibraryItemVisibilityToPublic
    }
`

/**
 * A React hook to get the configuration for the saved searches library and prompt library.
 */
export function useLibraryConfiguration(): LibraryConfiguration {
    const { data } = useQuery<LibraryConfigurationResult>(libraryConfigurationQuery, {
        onError(error) {
            logger.error('Failed to fetch library configuration:', error)
        },
    })
    return data ?? DEFAULT_LIBRARY_CONFIGURATION
}
