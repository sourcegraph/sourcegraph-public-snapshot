import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

export const LabelFragment = gql`
    fragment LabelFragment on Label {
        id
        name
        description
        color
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ILabelConnection | ErrorLike

/**
 * A React hook that observes labels queried from the GraphQL API.
 *
 * @param repository The repository in which to observe the labels defined.
 */
export const useLabels = (repository: Pick<GQL.IRepository, 'id'>): Result => {
    const [labels, setLabels] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query LabelsInRepository($repository: ID!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            labels {
                                nodes {
                                    ...LabelFragment
                                }
                                totalCount
                            }
                        }
                    }
                }
                ${LabelFragment}
            `,
            { repository: repository.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Repository') {
                        throw new Error('not a repository')
                    }
                    return data.node.labels
                })
            )
            .pipe(startWith(LOADING))
            .subscribe(setLabels, err => setLabels(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository])
    return labels
}
