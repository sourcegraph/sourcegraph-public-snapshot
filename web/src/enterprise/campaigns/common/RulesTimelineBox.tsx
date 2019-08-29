import React, { useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { RulesIcon } from '../../rules/icons'
import { Link } from 'react-router-dom'
import { useLocalStorage } from '../../../util/useLocalStorage'

interface Props {
    ruleContainer: GQL.RuleContainer
}

export const RulesTimelineBox: React.FunctionComponent<Props> = ({ ruleContainer }) => {
    const [showAllRules, setShowAllRules] = useLocalStorage('RulesTimelineBox.showAllRules', true)
    const toggleShowAllRules = useCallback(() => setShowAllRules(!showAllRules), [setShowAllRules, showAllRules])
    return (
        <div className="bg-body border mt-5">
            <div className="d-flex align-items-center w-100 p-4">
                <RulesIcon className="icon-inline h5 mb-0 mr-3" />
                {ruleContainer.rules.totalCount > 0 ? (
                    <>
                        <p className="flex-1 mb-0">Rules ran successfully</p>
                        <button type="button" className="btn btn-link" onClick={toggleShowAllRules}>
                            {showAllRules ? 'Hide' : 'Show'} all rules
                        </button>
                    </>
                ) : (
                    <div className="flex-1" />
                )}
                <Link to={`${ruleContainer.url}/manage`} className="btn btn-secondary">
                    Manage rules
                </Link>
            </div>
            {showAllRules && (
                <ul className="list-group list-group-flush">
                    {ruleContainer.rules.nodes.map(rule => (
                        <li key={rule.id} className="list-group-item px-5">
                            {rule.name}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}
