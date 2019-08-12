import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import SyncIcon from 'mdi-react/SyncIcon'
import React, { useCallback } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'
import { Modal } from 'reactstrap'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { EditRuleForm } from '../form/EditRuleForm'
import { NewRuleForm } from '../form/NewRuleForm'
import { useRules } from '../useRules'
import { RuleRow } from './RuleRow'

const FormModal: React.FunctionComponent<{ toggle: () => void }> = ({ toggle, children }) => (
    <Modal
        isOpen={true}
        backdrop={true}
        autoFocus={false}
        centered={true}
        scrollable={true}
        fade={false}
        toggle={toggle}
        className="container mx-auto"
    >
        {children}
    </Modal>
)

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps {
    container: Pick<GQL.RuleContainer, 'id'>

    /**
     * The base URL of the area.
     */
    match: { url: string }

    className?: string
    history: H.History
}

/**
 * A list of rules in a rule container.
 */
export const RulesList: React.FunctionComponent<Props> = ({ container, match, className = '', history, ...props }) => {
    const [rules, onRulesUpdate] = useRules(container)

    const dismissNewRuleForm = useCallback(() => history.replace(match.url), [history, match.url])

    return (
        <div className={`rules-list ${className}`}>
            <div className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="mb-0">Rules</h2>
                <Link to={`${match.url}/new`} className="btn btn-success">
                    New rule
                </Link>
            </div>
            <div className="border border-success p-3 my-4 d-flex align-items-start">
                <SyncIcon className="flex-0 mr-2" />
                <div className="flex-1">
                    <h4 className="mb-0">Continuously applying rules</h4>
                    <p className="mb-0">
                        The rules run when any base branch changes or when a new repository matches.
                        {/* TODO!(sqs): this isnt true */}
                    </p>
                </div>
            </div>
            <Switch>
                <Route path={`${match.url}/new`}>
                    <FormModal
                        toggle={dismissNewRuleForm} // TODO!(sqs) prompt to save
                    >
                        <NewRuleForm
                            {...props}
                            container={container}
                            onDismiss={dismissNewRuleForm}
                            onRuleCreate={onRulesUpdate}
                            className="p-4"
                            history={history}
                        />
                    </FormModal>
                </Route>
                <Route
                    path={`${match.url}/:ruleID`}
                    // tslint:disable-next-line: jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ ruleID: string }>) => {
                        const rule =
                            rules !== LOADING && !isErrorLike(rules)
                                ? rules.nodes.find(r => r.id === routeComponentProps.match.params.ruleID)
                                : undefined
                        return rule ? (
                            <FormModal toggle={dismissNewRuleForm}>
                                <EditRuleForm
                                    {...props}
                                    rule={{ ...rule, definition: rule.definition.raw }}
                                    onRuleUpdate={onRulesUpdate}
                                    onDismiss={dismissNewRuleForm}
                                    className="p-4"
                                    history={history}
                                />
                            </FormModal>
                        ) : (
                            <div className="alert alert-danger mb-3">Rule to edit not found</div>
                        )
                    }}
                />
            </Switch>
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
