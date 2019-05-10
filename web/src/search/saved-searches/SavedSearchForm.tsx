import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subscription } from 'rxjs'
import { catchError } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Form } from '../../components/Form'
export interface SavedQueryFields {
    id: GQL.ID
    description: string
    query: string
    notify: boolean
    notifySlack: boolean
    userID: GQL.ID | null
    orgID: GQL.ID | null
    slackWebhookURL: string | null
}

interface Props extends RouteComponentProps<{}> {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    defaultValues?: Partial<SavedQueryFields>
    title?: string
    submitLabel: string
    emailNotificationLabel: string
    onSubmit: (fields: SavedQueryFields) => Observable<void>
}

interface State {
    values: SavedQueryFields
    isSubmitting: boolean
    error?: any
}

export class SavedSearchForm extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            values: {
                id: (props.defaultValues && props.defaultValues.id) || '',
                description: (props.defaultValues && props.defaultValues.description) || '',
                query: (props.defaultValues && props.defaultValues.query) || '',
                notify: (props.defaultValues && props.defaultValues.notify) || false,
                notifySlack: (props.defaultValues && props.defaultValues.notifySlack) || false,
                userID: (props.defaultValues && props.defaultValues.userID) || null,
                orgID: (props.defaultValues && props.defaultValues.orgID) || null,
                slackWebhookURL: (props.defaultValues && props.defaultValues.slackWebhookURL) || null,
            },
            isSubmitting: false,
        }
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
                    ...this.state.values,
                    [key]: type === 'checkbox' ? checked : value,
                },
            }))
        }
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
                        this.setState({ error, isSubmitting: false })

                        return []
                    })
                )
                .subscribe()
        )
    }

    public render(): JSX.Element | null {
        const {
            values: { query, description, notify, notifySlack, slackWebhookURL },
        } = this.state

        return (
            <div className="saved-search-form">
                <div className="saved-search-form__header">
                    <h2>{this.props.title}</h2>
                    <div>Get notifications when there are new results for specific search queries</div>
                </div>
                <Form onSubmit={this.handleSubmit}>
                    <div className="saved-search-form__input">
                        <label className="saved-search-form__label">Description:</label>
                        <input
                            type="text"
                            name="description"
                            className="form-control"
                            placeholder="Description"
                            required={true}
                            value={description}
                            onChange={this.createInputChangeHandler('description')}
                        />
                    </div>
                    <div className="saved-search-form__input">
                        <label className="saved-search-form__label">Query:</label>
                        <input
                            type="text"
                            name="query"
                            className="form-control"
                            placeholder="Query"
                            required={true}
                            value={query}
                            onChange={this.createInputChangeHandler('query')}
                        />
                    </div>
                    <div className="saved-search-form__input">
                        <label className="saved-search-form__label">Email notifications:</label>
                        <div>
                            <input
                                type="checkbox"
                                name="Notify owner"
                                className="saved-search-form__checkbox"
                                defaultChecked={notify}
                                onChange={this.createInputChangeHandler('notify')}
                            />{' '}
                            <span>{this.props.emailNotificationLabel}</span>
                        </div>
                    </div>
                    {notifySlack && slackWebhookURL && (
                        <div className="saved-search-form__input">
                            <label className="saved-search-form__label">Slack notifications:</label>
                            <input
                                type="text"
                                name="Slack webhook URL"
                                className="form-control"
                                value={slackWebhookURL}
                                disabled={true}
                                onChange={this.createInputChangeHandler('slackWebhookURL')}
                            />
                            <label className="small">
                                Slack webhooks are deprecated and will be removed in a future Sourcegraph version.
                            </label>
                        </div>
                    )}
                    {this.isUnsupportedNotifyQuery(this.state.values) && (
                        <div className="alert alert-warning mb-3">
                            <strong>Warning:</strong> non-commit searches do not currently support notifications.
                            Consider adding <code>type:diff</code> or <code>type:commit</code> to your query.
                        </div>
                    )}
                    {notify && !window.context.emailEnabled && !this.isUnsupportedNotifyQuery(this.state.values) && (
                        <div className="alert alert-warning mb-3">
                            <strong>Warning:</strong> Sending emails is not currently configured on this Sourcegraph
                            server.{' '}
                            {this.props.authenticatedUser && this.props.authenticatedUser.siteAdmin
                                ? 'Use the email.smtp site configuration setting to enable sending emails.'
                                : 'Contact your server admin for more information.'}
                        </div>
                    )}
                    <button type="submit" className="btn btn-primary saved-search-form__submit-button">
                        {this.props.submitLabel}
                    </button>
                    {this.state.error && !this.state.isSubmitting && (
                        <div className="alert alert-danger mb-3">
                            <strong>Error:</strong> {this.state.error.message}
                        </div>
                    )}
                </Form>
            </div>
        )
    }
    /**
     * Tells if the query is unsupported for sending notifications.
     */
    private isUnsupportedNotifyQuery(v: SavedQueryFields): boolean {
        const notifying = v.notify || v.notifySlack
        return notifying && !v.query.includes('type:diff') && !v.query.includes('type:commit')
    }
}
