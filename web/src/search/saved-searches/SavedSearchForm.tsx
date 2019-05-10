import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subscription } from 'rxjs'
import { catchError } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Form } from '../../components/Form'
import { OrgAreaPageProps } from '../../org/area/OrgArea'

export interface SavedQueryFields {
    description: string
    query: string
    notify: boolean
    notifySlack: boolean
    userID: number | null
    orgID: number | null
    slackWebhookURL: string | null
}

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: GQL.IUser | null
    defaultValues?: Partial<SavedQueryFields>
    title?: string
    submitLabel: string
    onSubmit: (fields: SavedQueryFields) => Observable<void>
    onDidCommit: () => void
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
                description: (props.defaultValues && props.defaultValues.description) || '',
                query: (props.defaultValues && props.defaultValues.query) || '',
                notify: (props.defaultValues && props.defaultValues.notify) || false,
                notifySlack: (props.defaultValues && props.defaultValues.notifySlack) || false,
                userID: (props.defaultValues && props.defaultValues.userID) || null,
                orgID: (props.defaultValues && props.defaultValues.orgID) || null,
                slackWebhookURL: null,
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
                        this.setState({ error })

                        return []
                    })
                )
                .subscribe(() => this.props.onDidCommit())
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="saved-search-form">
                <div className="saved-search-form__header">
                    <h2>{this.props.title}</h2>
                    <div>Get notifications when there are new results for specific search queries</div>
                </div>
                <Form onSubmit={this.handleSubmit}>
                    <div className="saved-search-form__input">
                        <label>Description</label>
                        <input
                            type="text"
                            name="description"
                            className="form-control"
                            placeholder="Description"
                            required={true}
                            onChange={this.createInputChangeHandler('description')}
                        />
                    </div>
                    <div className="saved-search-form__input">
                        <label>Query</label>
                        <input
                            type="text"
                            name="query"
                            className="form-control"
                            placeholder="Query"
                            required={true}
                            onChange={this.createInputChangeHandler('query')}
                        />
                    </div>
                    <div className="saved-search-form__input">
                        <span>
                            <input
                                type="checkbox"
                                name="Notify owner"
                                className="saved-search-form__checkbox"
                                required={true}
                                onChange={this.createInputChangeHandler('notify')}
                            />{' '}
                            Send email notifications to all members of this organization
                        </span>
                    </div>
                    <button type="submit" className="btn btn-primary">
                        {this.props.submitLabel}
                    </button>
                    {this.state.error && !this.state.isSubmitting && (
                        <div className="alert alert-danger mb-2">
                            <strong>Error:</strong> {this.state.error.message}
                        </div>
                    )}
                </Form>
            </div>
        )
    }
}
