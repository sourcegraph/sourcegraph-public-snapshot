import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { from, fromEvent, Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, map } from 'rxjs/operators'
import { Key } from 'ts-key-enum'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps, SettingsSubject } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { eventLogger } from '../../tracking/eventLogger'

export interface SavedQueryFields {
    key: string
    description: string
    query: string
    subject: GQL.ID
    showOnHomepage: boolean
    notify: boolean
    notifySlack: boolean
    ownerKind: GQL.SavedQueryOwnerKind
    userID: number | null
    orgID: number | null
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
    slackWebhooks: Map<GQL.ID, string | null> // subject GraphQL ID -> slack webhook
}

export class SavedQueryForm extends React.Component<Props, State> {
    private handleDescriptionChange = this.createInputChangeHandler('description')
    private handleShowOnHomeChange = this.createInputChangeHandler('showOnHomepage')
    private handleNotifyChange = this.createInputChangeHandler('notify')
    private handleNotifySlackChange = this.createInputChangeHandler('notifySlack') // This needs to update the subject to correspond to the org that is selected now.

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const { defaultValues } = props

        this.state = {
            values: {
                key: (defaultValues && defaultValues.key) || '',
                query: (defaultValues && defaultValues.query) || '',
                description: (defaultValues && defaultValues.description) || '',
                subject: (defaultValues && defaultValues.subject) || '',
                showOnHomepage: !!(defaultValues && defaultValues.showOnHomepage),
                notify: !!(defaultValues && defaultValues.notify),
                notifySlack: !!(defaultValues && defaultValues.notifySlack),
                ownerKind: (defaultValues && defaultValues.ownerKind) || ('USER' as GQL.SavedQueryOwnerKind.USER),
                userID: (defaultValues && defaultValues.userID) || null,
                orgID: (defaultValues && defaultValues.orgID) || null,
            },
            subjectOptions: [],
            isSubmitting: false,
            isFocused: false,
            sawUnsupportedNotifyQueryWarning: false,
            slackWebhooks: new Map<GQL.ID, string | null>(),
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            from(this.componentUpdates)
                .pipe(
                    map(({ settingsCascade }) => settingsCascade),
                    filter(isSettingsValid)
                )
                .subscribe(settingsCascade => {
                    const subject = settingsCascade.subjects.find(s => !!s.subject.id)

                    this.setState(state => ({
                        subjectOptions: settingsCascade.subjects.map(({ subject }) => subject),
                        values: {
                            ...state.values,
                            subject: state.values.subject || (subject && subject.subject.id) || '',
                        },
                    }))

                    settingsCascade.subjects.map(subject => {
                        if (subject.settings) {
                            let slackWebhookURL: string | null
                            try {
                                const settings = subject.settings
                                if (settings && settings['notifications.slack']) {
                                    slackWebhookURL = settings['notifications.slack']!.webhookURL
                                }
                            } catch (e) {
                                slackWebhookURL = null
                            }
                            this.setState(state => ({
                                slackWebhooks: state.slackWebhooks.set(subject.subject.id, slackWebhookURL),
                            }))
                        }
                    })
                })
        )

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
            values: { query, description, subject, showOnHomepage, notify, notifySlack, userID, orgID },
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
                    <div>Organizations</div>
                    <div className="saved-query-form__save-location">
                        {this.props.authenticatedUser &&
                            this.props.authenticatedUser.organizations.nodes.map((org, i) => (
                                <label className="saved-query-form__save-location-options" key={i}>
                                    <input
                                        className="saved-query-form__save-location-input"
                                        onChange={this.orgOwnerChangeHandler(org.databaseID)}
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
                            <label>
                                <input
                                    className="saved-query-form__save-location-input"
                                    type="checkbox"
                                    defaultChecked={showOnHomepage}
                                    onChange={this.handleShowOnHomeChange}
                                />{' '}
                                Show on homepage
                            </label>
                        </span>
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
                {notifySlack && this.isSubjectMissingSlackWebhook() && (
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
                    <button type="submit" className="btn btn-primary" disabled={!subject || isSubmitting}>
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

    private isSubjectMissingSlackWebhook = () => {
        const chosen = this.state.subjectOptions.find(subjectOption => subjectOption.id === this.state.values.subject)
        if (!chosen) {
            return false
        }
        return !this.state.slackWebhooks.get(chosen.id)
    }

    private getConfigureSlackURL = () => {
        const chosen = this.state.subjectOptions.find(subjectOption => subjectOption.id === this.state.values.subject)
        if (!chosen) {
            return ''
        }
        if (chosen.__typename === 'Org') {
            return `/organizations/${chosen.name}/settings`
        }
        if (chosen.__typename === 'User') {
            return `/settings`
        }
        return '' // unexpected
    }

    private saveTargetName = () => {
        const chosen = this.state.subjectOptions
            .filter(
                (subjectOption: SettingsSubject): subjectOption is GQL.IOrg | GQL.IUser =>
                    subjectOption.__typename === 'Org' || subjectOption.__typename === 'User'
            )
            .find(subjectOption => subjectOption.id === this.state.values.subject)
        return chosen && settingsSubjectLabel(chosen, true)
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
                .subscribe(this.props.onDidCommit)
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

    private orgOwnerChangeHandler(id: number): React.FormEventHandler<HTMLInputElement> {
        return event => {
            this.setState(state => ({
                values: {
                    ...state.values,
                    ownerKind: 'ORG' as GQL.SavedQueryOwnerKind.ORG,
                    orgID: id,
                    userID: null,
                },
            }))
        }
    }

    private userOwnerChangeHandler(id: number): React.FormEventHandler<HTMLInputElement> {
        return event => {
            this.setState(state => ({
                values: {
                    ...state.values,
                    ownerKind: 'USER' as GQL.SavedQueryOwnerKind.USER,
                    userID: id,
                    orgID: null,
                },
            }))
        }
    }
}

function settingsSubjectLabel(s: GQL.IUser | GQL.IOrg, short?: boolean): string {
    switch (s.__typename) {
        case 'User':
            return short ? s.username : `${s.username} (user settings)`
        case 'Org':
            return short ? s.name : `${s.name} (org settings)`
    }
}
