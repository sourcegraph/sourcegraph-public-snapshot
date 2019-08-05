import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { queryAndFragmentForThreadOrIssueOrChangeset } from '../../../threadlike/util/graphql'

const LOADING: 'loading' = 'loading'

const { fragment, query } = queryAndFragmentForThreadOrIssueOrChangeset()

/**
 * A React hook that observes all threads (queried from the GraphQL API).
 */
export const useThreads = (open = true): typeof LOADING | GQL.IThreadOrIssueOrChangesetConnection | ErrorLike => {
    const [result, setResult] = useState<typeof LOADING | GQL.IThreadOrIssueOrChangesetConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query Threads($open: Boolean) {
                    threadOrIssueOrChangesets(open: $open) {
                        nodes {
                            ${query}
                        }
                        totalCount
                    }
                }
                ${fragment}
            `,
            { open }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => data.threadOrIssueOrChangesets),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [open])
    return result
}
