import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { from, fromEvent, Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, map } from 'rxjs/operators'
import { Key } from 'ts-key-enum'
import { Link } from '../../../../shared/src/components/Link'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps, SettingsSubject } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { eventLogger } from '../../tracking/eventLogger'

export interface SavedQueryFields {
    id: string
    description: string
    query: string
    subject: GQL.ID
    notify: boolean
    notifySlack: boolean
    userID: number | null
    orgID: number | null
    slackWebhookURL: string | null
    orgName: string
}

interface Props extends SettingsCascadeProps {
    authenticatedUser: GQL.IUser | null
    defaultValues?: Partial<SavedQueryFields>
    title?: string
    submitLabel: string
    onSubmit: (fields: SavedQueryFields) => Observable<void>
    onDidCommit: () => void
    onDidCancel: () => void
}

interface State {
    values: SavedQueryFields

    subjectOptions: SettingsSubject[]
    isSubmitting: boolean
    isFocused: boolean
    error?: any
    sawUnsupportedNotifyQueryWarning: boolean
}

export class SavedQueryForm extends React.Component<Props, State> {
    private handleDescriptionChange = this.createInputChangeHandler('description')
    private handleNotifyChange = this.createInputChangeHandler('notify')
    private handleNotifySlackChange = this.createInputChangeHandler('notifySlack') // This needs to update the subject to correspond to the org that is selected now.

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const { defaultValues } = props

        this.state = {
            values: {
                id: (defaultValues && defaultValues.id) || '',
                query: (defaultValues && defaultValues.query) || '',
                description: (defaultValues && defaultValues.description) || '',
                subject: (defaultValues && defaultValues.subject) || '',
                notify: !!(defaultValues && defaultValues.notify),
                notifySlack: !!(defaultValues && defaultValues.notifySlack),
                userID:
                    (defaultValues && defaultValues.userID) ||
                    (defaultValues && defaultValues.orgID
                        ? null
                        : this.props.authenticatedUser && this.props.authenticatedUser.databaseID) ||
                    null,
                orgID: (defaultValues && defaultValues.orgID) || null,
                slackWebhookURL: (defaultValues && defaultValues.slackWebhookURL) || null,
                orgName: (defaultValues && defaultValues.orgName) || '',
            },
            subjectOptions: [],
            isSubmitting: false,
            isFocused: false,
            sawUnsupportedNotifyQueryWarning: false,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => !this.state.isFocused && event.key === Key.Escape && !this.state.isSubmitting))
                .subscribe(() => this.props.onDidCancel())
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        const { onDidCancel, title, submitLabel } = this.props
        const {
            values: { query, description, notify, notifySlack, userID, orgID },
            isSubmitting,
            error,
        } = this.state

        return (
            <Form className="saved-query-form" onSubmit={this.handleSubmit}>
                <button type="reset" className="btn btn-icon saved-query-form__close" onClick={onDidCancel}>
                    <CloseIcon className="icon-inline" />
                </button>
                <h2 className="saved-query-form__title">{title}</h2>
                <div className="form-group">
                    <label>Search query</label>
                    <input
                        type="text"
                        name="query"
                        className="form-control"
                        placeholder="Query"
                        onChange={this.handleQueryChange}
                        value={query}
                        required={true}
                        autoCorrect="off"
                        spellCheck={false}
                        autoCapitalize="off"
                        autoFocus={!query}
                        onFocus={this.handleInputFocus}
                        onBlur={this.handleInputBlur}
                    />
                </div>
                <div className="form-group">
                    <label>Description of this search</label>
                    <input
                        type="text"
                        name="description"
                        className="form-control"
                        placeholder="Description"
                        onChange={this.handleDescriptionChange}
                        value={description}
                        required={true}
                        autoFocus={!!query && !description}
                        onFocus={this.handleInputFocus}
                        onBlur={this.handleInputBlur}
                    />
                </div>
                <div className="form-group">
                    <label>Save location</label>
                    <div>User</div>
                    <div className="saved-query-form__save-location">
                        {this.props.authenticatedUser && (
                            <label className="saved-query-form__save-location-options">
                                <input
                                    className="saved-query-form__save-location-input"
                                    onChange={this.userOwnerChangeHandler(this.props.authenticatedUser.databaseID)}
                                    type="radio"
                                    value={this.props.authenticatedUser.username}
                                    checked={this.props.authenticatedUser.databaseID === userID}
                                />
                                {this.props.authenticatedUser.username}
                            </label>
                        )}
                    </div>
                    {this.props.authenticatedUser && this.props.authenticatedUser.organizations.nodes.length > 0 && (
                        <div>Organizations</div>
                    )}
                    <div className="saved-query-form__save-location">
                        {this.props.authenticatedUser &&
                            this.props.authenticatedUser.organizations.nodes.map((org, i) => (
                                <label className="saved-query-form__save-location-options" key={i}>
                                    <input
                                        className="saved-query-form__save-location-input"
                                        onChange={this.orgOwnerChangeHandler(org)}
                                        type="radio"
                                        value={org.name}
                                        checked={org.databaseID === orgID}
                                    />
                                    {org.name}
                                </label>
                            ))}
                    </div>
                    <div className="saved-query-form__save-location">
                        <span className="saved-query-form__save-location-options">
                            <label data-tooltip={`Send email notifications to config owner (${this.saveTargetName()})`}>
                                <input
                                    className="saved-query-form__save-location-input"
                                    type="checkbox"
                                    defaultChecked={notify}
                                    onChange={this.handleNotifyChange}
                                />{' '}
                                Email notifications
                            </label>
                        </span>
                        {notifySlack && (
                            <span className="saved-query-form__save-location-options">
                                <label
                                    data-tooltip={`Send Slack notifications to webhook defined in configuration for ${this.saveTargetName()}`}
                                >
                                    <input
                                        className="saved-query-form__save-location-input"
                                        type="checkbox"
                                        defaultChecked={notifySlack}
                                        onChange={this.handleNotifySlackChange}
                                    />{' '}
                                    Slack notifications
                                </label>
                            </span>
                        )}
                    </div>
                </div>
                {this.isUnsupportedNotifyQuery(this.state.values) && (
                    <div className="alert alert-warning mb-2">
                        <strong>Warning:</strong> non-commit searches do not currently support notifications. Consider
                        adding <code>type:diff</code> or <code>type:commit</code> to your query.
                    </div>
                )}
                {notify && !window.context.emailEnabled && !this.isUnsupportedNotifyQuery(this.state.values) && (
                    <div className="alert alert-warning mb-2">
                        <strong>Warning:</strong> Sending emails is not currently configured on this Sourcegraph server.{' '}
                        {this.props.authenticatedUser && this.props.authenticatedUser.siteAdmin
                            ? 'Use the email.smtp site configuration setting to enable sending emails.'
                            : 'Contact your server admin for more information.'}
                    </div>
                )}
                {notifySlack && this.isSavedSearchMissingSlackWebhook() && (
                    <div className="alert alert-warning mb-2">
                        <strong>Required:</strong>{' '}
                        <Link target="_blank" to={this.getConfigureSlackURL()}>
                            Configure a Slack webhook URL
                        </Link>{' '}
                        to receive Slack notifications.
                    </div>
                )}
                {error && !isSubmitting && (
                    <div className="alert alert-danger mb-2">
                        <strong>Error:</strong> {error.message}
                    </div>
                )}
                <div>
                    <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
                        {submitLabel}
                    </button>{' '}
                    <button type="reset" className="btn btn-secondary" disabled={isSubmitting} onClick={onDidCancel}>
                        Cancel
                    </button>
                </div>
            </Form>
        )
    }

    /**
     * Tells if the query is unsupported for sending notifications.
     */
    private isUnsupportedNotifyQuery(v: SavedQueryFields): boolean {
        const notifying = v.notify || v.notifySlack
        return notifying && !v.query.includes('type:diff') && !v.query.includes('type:commit')
    }

    private isSavedSearchMissingSlackWebhook = () =>
        !this.state.values.slackWebhookURL || this.state.values.slackWebhookURL === ''

    private getConfigureSlackURL = () => {
        const chosen = this.state.subjectOptions.find(subjectOption => subjectOption.id === this.state.values.subject)
        if (!chosen) {
            return 'https://docs.sourcegraph.com/user/search/saved_searches#configuring-email-and-slack-notifications'
        }
        if (chosen.__typename === 'Org') {
            return `/organizations/${chosen.name}/settings`
        }
        if (chosen.__typename === 'User') {
            return `/settings`
        }
        return 'https://docs.sourcegraph.com/user/search/saved_searches#configuring-email-and-slack-notifications' // unexpected
    }

    private saveTargetName = () => {
        const chosen: 'user' | 'org' = this.state.values.userID === null ? 'user' : 'org'
        if (chosen === 'user') {
            return (this.props.authenticatedUser && this.props.authenticatedUser.username) || '(user)'
        }
        return this.state.values.orgName
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()

        this.setState({ isSubmitting: true })

        this.subscriptions.add(
            this.props
                .onSubmit(this.state.values)
                .pipe(
                    catchError(error => {
                        console.error(error)
                        this.setState({ error })

                        return []
                    })
                )
                .subscribe(() => this.props.onDidCommit())
        )
    }

    private handleInputFocus = () => {
        this.setState(() => ({ isFocused: true }))
    }

    private handleInputBlur = () => {
        this.setState(() => ({ isFocused: false }))
    }

    private handleQueryChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const newValues = {
            ...this.state.values,
            query: event.currentTarget.value,
        }
        if (!this.state.sawUnsupportedNotifyQueryWarning && this.isUnsupportedNotifyQuery(newValues)) {
            this.setState({ sawUnsupportedNotifyQueryWarning: true })
            eventLogger.log('SavedSearchUnsupportedNotifyQueryWarning')
        }

        const newQuery = event.currentTarget.value
        this.setState(prevState => ({
            values: {
                ...prevState.values,
                query: newQuery,
            },
        }))
    }

    /**
     * Returns an input change handler that updates the SavedQueryFields in the component's state
     *
     * @param key The key of saved query fields that a change of this input should update
     */
    private createInputChangeHandler(key: keyof SavedQueryFields): React.FormEventHandler<HTMLInputElement> {
        return event => {
            const { value, checked, type } = event.currentTarget
            this.setState(state => ({
                values: {
                    ...state.values,
                    [key]: type === 'checkbox' ? checked : value,
                },
            }))
        }
    }

    private orgOwnerChangeHandler(org: GQL.IOrg): React.FormEventHandler<HTMLInputElement> {
        return event => {
            this.setState(state => ({
                values: {
                    ...state.values,
                    orgID: org.databaseID,
                    userID: null,
                    orgName: org.displayName || org.name,
                },
            }))
        }
    }

    private userOwnerChangeHandler(id: number): React.FormEventHandler<HTMLInputElement> {
        return event => {
            this.setState(state => ({
                values: {
                    ...state.values,
                    userID: id,
                    orgID: null,
                },
            }))
        }
    }
}
