import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

export const RuleFragment = gql`
    fragment RuleFragment on Rule {
        id
        name
        description
        definition {
            raw
            parsed
        }
        createdAt
        updatedAt
        url
        viewerCanUpdate
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IRuleConnection | ErrorLike

/**
 * A React hook that observes rules queried from the GraphQL API.
 *
 * @param container The container whose rules to observe.
 */
export const useRules = (container: Pick<GQL.RuleContainer, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)

    const [rules, setRules] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query Rules($container: ID!) {
                    ruleContainer(id: $container) {
                        __typename
                        rules {
                            nodes {
                                ...RuleFragment
                            }
                            totalCount
                        }
                    }
                }
                ${RuleFragment}
            `,
            { container: container.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.ruleContainer) {
                        throw new Error('no rule container')
                    }
                    return data.ruleContainer.rules
                })
            )
            .pipe(startWith(LOADING))
            .subscribe(setRules, err => setRules(asError(err)))
        return () => subscription.unsubscribe()
    }, [container, updateSequence])

    const onUpdate = useCallback(() => {
        // Fetch from server.
        setUpdateSequence(updateSequence + 1)
    }, [updateSequence])
    return [rules, onUpdate]
}
