import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { LabelFragment } from '../useLabels'

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ILabelConnection | ErrorLike

/**
 * A React hook that observes all labels that are applied to the given labelable (queried from the
 * GraphQL API).
 *
 * @param labelable The labelable whose labels to observe.
 */
export const useLabelableLabels = (labelable: Pick<GQL.Labelable, '__typename' | 'id'>): [Result, () => void] => {
    if (labelable.__typename !== 'Thread') {
        throw new Error('only Thread Labelable is supported')
    }

    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query LabelableLabels($labelable: ID!) {
                    node(id: $labelable) {
                        __typename
                        ... on Thread {
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
            { labelable: labelable.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Thread') {
                        throw new Error('not a thread')
                    }
                    return data.node.labels
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [labelable, updateSequence])
    return [result, incrementUpdateSequence]
}
