import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { ActorFragment } from '../../../actor/graphql'
import { queryGraphQL } from '../../../backend/graphql'
import { ThreadFragment } from '../util/graphql'

export const ThreadConnectionFiltersFragment = gql`
    fragment ThreadConnectionFiltersFragment on ThreadConnectionFilters {
        repository {
            repository {
                id
                name
            }
            count
            isApplied
        }
        label {
            label {
                name
                color
            }
            labelName
            count
            isApplied
        }
        openCount
        closedCount
    }
`

export const ThreadConnectionFragment = gql`
    fragment ThreadConnectionFragment on ThreadConnection {
        nodes {
            ...ThreadFragment
        }
        totalCount
        filters {
            ...ThreadConnectionFiltersFragment
        }
    }
`

export const ThreadOrThreadPreviewConnectionFragment = gql`
    fragment ThreadOrThreadPreviewConnectionFragment on ThreadOrThreadPreviewConnection {
        nodes {
            ... on Thread {
                ...ThreadFragment
            }
            ... on ThreadPreview {
                ...ThreadPreviewFragment
            }
        }
        totalCount
        filters {
            ...ThreadConnectionFiltersFragment
        }
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IThreadConnection | ErrorLike

/**
 * A React hook that observes threads queried from the GraphQL API.
 */
export const useThreads = (filters: GQL.IThreadFilters | null): Result => {
    const filtersJSON = JSON.stringify(filters)
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query Threads($filters: ThreadFilters) {
                    threads(filters: $filters) {
                        ...ThreadConnectionFragment
                    }
                }
                ${ThreadConnectionFragment}
                ${ThreadConnectionFiltersFragment}
                ${ThreadFragment}
                ${ActorFragment}
            `,
            { filters: JSON.parse(filtersJSON) }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => data.threads),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [filtersJSON])
    return result
}
