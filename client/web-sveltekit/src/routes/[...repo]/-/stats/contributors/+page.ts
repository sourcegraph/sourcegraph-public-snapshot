import { map } from 'rxjs/operators'

import type { PageLoad } from './$types'

import { isErrorLike } from '$lib/common'
import type { PagedRepositoryContributorsResult, PagedRepositoryContributorsVariables } from '$lib/graphql-operations'
import { CONTRIBUTORS_QUERY } from '$lib/loader/repo'
import { getPaginationParams } from '$lib/Paginator.svelte'
import { asStore } from '$lib/utils'

const pageSize = 20

const emptyPage: Extract<PagedRepositoryContributorsResult['node'], { __typename: 'Repository' }>['contributors'] = {
    totalCount: 0,
    nodes: [] as any[],
    pageInfo: {
        hasNextPage: false,
        hasPreviousPage: false,
        endCursor: null,
        startCursor: null,
    },
}

export const load: PageLoad = ({ url, parent }) => {
    const afterDate = url.searchParams.get('after') ?? ''
    const { first, last, before, after } = getPaginationParams(url.searchParams, pageSize)

    const contributors = asStore(
        parent().then(({ resolvedRevision, platformContext }) =>
            !isErrorLike(resolvedRevision)
                ? platformContext
                      .requestGraphQL<PagedRepositoryContributorsResult, PagedRepositoryContributorsVariables>({
                          request: CONTRIBUTORS_QUERY,
                          variables: {
                              afterDate,
                              repo: resolvedRevision.repo.id,
                              revisionRange: '',
                              path: '',
                              first,
                              last,
                              after,
                              before,
                          },
                          mightContainPrivateInfo: true,
                      })
                      .pipe(
                          map(result =>
                              result.data?.node?.__typename === 'Repository' ? result.data.node.contributors : emptyPage
                          )
                      )
                      .toPromise()
                : null
        )
    )

    return {
        after: afterDate,
        contributors,
    }
}
