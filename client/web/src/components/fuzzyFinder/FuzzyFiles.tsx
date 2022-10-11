import { ApolloClient } from '@apollo/client'
import { mdiFileDocumentOutline } from '@mdi/js'

import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { Icon } from '@sourcegraph/wildcard'

import { getWebGraphQLClient } from '../../backend/graphql'
import { createUrlFunction } from '../../fuzzyFinder/WordSensitiveFuzzySearch'
import { FileNamesResult, FileNamesVariables } from '../../graphql-operations'

import { FuzzyFSM, newFuzzyFSMFromValues } from './FuzzyFsm'
import { FuzzyRepoRevision } from './FuzzyRepoRevision'

export const FUZZY_FILES_QUERY = gql`
    query FileNames($repository: String!, $commit: String!) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                id
                fileNames
            }
        }
    }
`

export async function loadFilesFSM(
    apolloClient: ApolloClient<object> | undefined,
    repoRevision: FuzzyRepoRevision,
    createURL: createUrlFunction
): Promise<FuzzyFSM> {
    try {
        const client = apolloClient || (await getWebGraphQLClient())
        const response = await client.query<FileNamesResult, FileNamesVariables>({
            query: getDocumentNode(FUZZY_FILES_QUERY),
            variables: { repository: repoRevision.repositoryName, commit: repoRevision.revision },
        })
        if (response.errors && response.errors.length > 0) {
            return { key: 'failed', errorMessage: JSON.stringify(response.errors) }
        }
        if (response.error) {
            return { key: 'failed', errorMessage: JSON.stringify(response.error) }
        }
        const filenames = response.data.repository?.commit?.fileNames || []
        return newFuzzyFSM(filenames, createURL)
    } catch (error) {
        return { key: 'failed', errorMessage: JSON.stringify(error) }
    }
}

export function newFuzzyFSM(filenames: string[], createUrl: createUrlFunction): FuzzyFSM {
    return newFuzzyFSMFromValues(
        filenames.map(file => ({
            text: file,
            icon: <Icon aria-label={file} svgPath={mdiFileDocumentOutline} />,
        })),
        createUrl
    )
}
