import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { RuleScope } from '../types'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all rules queried from the GraphQL API defined in a particular scope.
 *
 * @param scope The scope in which to observe the rules defined.
 */
export const useRulesDefinedInScope = (scope: RuleScope): typeof LOADING | GQL.IRule[] | ErrorLike => {
    const [rulesOrError, setRulesOrError] = useState<typeof LOADING | GQL.IRule[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query RulesDefinedIn($project: ID!) {
                    node(id: $project) {
                        __typename
                        ... on Project {
                            rules {
                                nodes {
                                    id
                                    name
                                    url
                                }
                            }
                        }
                    }
                }
            `,
            { project: scope.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Project') {
                        throw new Error('not a project')
                    }
                    return data.node.rules.nodes
                }),
                startWith(LOADING)
            )
            .subscribe(setRulesOrError, err => setRulesOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [scope])
    return rulesOrError
}
