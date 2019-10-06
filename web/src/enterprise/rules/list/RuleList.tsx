import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H, { locationsAreEqual } from 'history'
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
import { RuleListItem } from './RuleListItem'

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
    location: H.Location
    history: H.History
}

/**
 * A list of rules in a rule container.
 */
export const RuleList: React.FunctionComponent<Props> = ({
    container,
    match,
    className = '',
    location,
    history,
    ...props
}) => {
    const [rules, onRulesUpdate] = useRules(container)

    const dismissNewRuleForm = useCallback(() => history.replace(match.url), [history, match.url])

    return (
        <div className={`rule-list ${className}`}>
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
                    // eslint-disable-next-line react/jsx-no-bind
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
            ) : rules.nodes.length > 0 ? (
                <ul className="list-unstyled">
                    {rules.nodes.map(rule => (
                        <li key={rule.id}>
                            <RuleListItem
                                {...props}
                                rule={rule}
                                onRuleUpdate={onRulesUpdate}
                                isEditing={location.pathname === rule.url}
                                stopEditingUrl={match.url}
                                className="mb-4"
                                history={history}
                            />
                        </li>
                    ))}
                </ul>
            ) : (
                <div className="mt-3 text-muted">No rules yet.</div>
            )}
        </div>
    )
}
