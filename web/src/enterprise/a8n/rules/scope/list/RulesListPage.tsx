import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React from 'react'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { HeroPage } from '../../../../../components/HeroPage'
import { RulesAreaContext } from '../ScopeRulesArea'
import { useRulesForScope } from '../useRulesForScope'

interface Props extends Pick<RulesAreaContext, 'scope' | 'rulesURL'> {}

const LOADING: 'loading' = 'loading'

/**
 * A page showing a list of rules for a particular scope.
 */
export const RulesListPage: React.FunctionComponent<Props> = ({ scope, ...props }) => {
    const rulesOrError = useRulesForScope(scope)
    if (rulesOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(rulesOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={rulesOrError.message} />
    }
    return (
        <div className="rules-list-page">
            <RulesList {...props} rulesOrError={rulesOrError} itemClassName="text-truncate" />
        </div>
    )
}
