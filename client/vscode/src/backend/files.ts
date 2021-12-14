import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { FileNamesResult, FileNamesVariables } from '../graphql-operations'

import { requestGraphQLFromVSCode } from './requestGraphQl'

const fileNamesQuery = gql`
    query FileNames($repository: String!, $revision: String!) {
        repository(name: $repository) {
            commit(rev: $revision) {
                fileNames
            }
        }
    }
`

export async function getFiles(variables: FileNamesVariables): Promise<string[]> {
    const result = await requestGraphQLFromVSCode<FileNamesResult, FileNamesVariables>(fileNamesQuery, variables)

    if (result.data?.repository?.commit) {
        return result.data.repository.commit.fileNames
    }

    // TODO error handling
    throw new Error(`Failed to fetch file names for ${variables.repository}@${variables.revision}`)
}
