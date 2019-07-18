import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a rule queried from the GraphQL API by ID.
 *
 * @param scope The scope in which to observe the rules.
 */
export const useRuleByID = (id: GQL.ID): typeof LOADING | GQL.IRule | null | ErrorLike => {
    const [ruleOrError, setRuleOrError] = useState<typeof LOADING | GQL.IRule | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query RuleByID($rule: ID!) {
                    node(id: $rule) {
                        __typename
                        ... on Rule {
                            id
                            name
                            description
                            settings
                            url
                        }
                    }
                }
            `,
            { rule: id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Rule') {
                        return null
                    }
                    return data.node
                }),
                startWith(LOADING)
            )
            .subscribe(setRuleOrError, err => setRuleOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [id])
    return ruleOrError
}
