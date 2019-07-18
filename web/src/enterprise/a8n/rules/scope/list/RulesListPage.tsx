import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { HeroPage } from '../../../../../components/HeroPage'
import { RulesList } from '../../list/RulesList'
import { RulesAreaContext } from '../ScopeRulesArea'
import { useRulesDefinedInScope } from '../useRulesDefinedInScope'

interface Props extends Pick<RulesAreaContext, 'scope' | 'rulesURL'> {
    newRuleURL: string
}

const LOADING: 'loading' = 'loading'

/**
 * A page showing a list of rules for a particular scope.
 */
export const RulesListPage: React.FunctionComponent<Props> = ({ newRuleURL, scope, rulesURL, ...props }) => {
    const rulesOrError = useRulesDefinedInScope(scope)
    if (rulesOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(rulesOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={rulesOrError.message} />
    }
    return (
        <div className="rules-list-page">
            <div className="d-flex align-items-center my-4">
                <Link to={newRuleURL} className="btn btn-success">
                    New rule
                </Link>
            </div>
            <RulesList {...props} rulesURL={rulesURL} rules={rulesOrError} itemClassName="text-truncate" />
        </div>
    )
}
