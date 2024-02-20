import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TagsPage_TagsQuery } from './page.gql'

export const load: PageLoad = ({ params }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    return {
        tags: client
            .query(TagsPage_TagsQuery, {
                repoName,
                first: 20,
                withBehindAhead: false,
            })
            .then(
                mapOrThrow(result => {
                    if (!result.data?.repository) {
                        throw new Error('Unable to load repository data.')
                    }
                    return result.data.repository.gitRefs
                })
            ),
    }
}
