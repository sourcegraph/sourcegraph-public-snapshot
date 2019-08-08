import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { useRules } from '../useRules'
import { NewRuleForm } from './NewRuleForm'
import { RuleRow } from './RuleRow'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps {
    container: Pick<GQL.RuleContainer, 'id'>

    className?: string
}

/**
 * A list of rules in a rule container.
 */
export const RulesList: React.FunctionComponent<Props> = ({ container, className = '', ...props }) => {
    const [rules, onRulesUpdate] = useRules(container)

    const [isShowingNewRuleForm, setIsShowingNewRuleForm] = useState(false)
    const toggleIsShowingNewRuleForm = useCallback(() => setIsShowingNewRuleForm(!isShowingNewRuleForm), [
        isShowingNewRuleForm,
    ])

    return (
        <div className={`rules-list ${className}`}>
            <div className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="mb-0">Rules</h2>
                <button type="button" className="btn btn-success" onClick={toggleIsShowingNewRuleForm}>
                    New rule
                </button>
            </div>
            {isShowingNewRuleForm && (
                <NewRuleForm
                    container={container}
                    onDismiss={toggleIsShowingNewRuleForm}
                    onRuleCreate={onRulesUpdate}
                    className="my-3 p-3 border"
                />
            )}
            {rules === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(rules) ? (
                <div className="alert alert-danger mt-3">{rules.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header">
                        <span className="text-muted">
                            {rules.totalCount} {pluralize('rule', rules.totalCount)}
                        </span>
                    </div>
                    {rules.nodes.length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {rules.nodes.map(rule => (
                                <li key={rule.id} className="list-group-item p-3">
                                    <RuleRow {...props} rule={rule} onRuleUpdate={onRulesUpdate} />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-3 text-muted">No rules yet.</div>
                    )}
                </div>
            )}
        </div>
    )
}
