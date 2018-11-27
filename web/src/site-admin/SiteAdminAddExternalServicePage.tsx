import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mergeMap, tap } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { mutateGraphQL } from '../backend/graphql'
import { Form } from '../components/Form'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    history: H.History
}

interface State extends GQL.IAddExternalServiceInput {
    /**
     * Holds any error returned by the remote GraphQL endpoint on failed requests.
     */
    error?: Error

    /**
     * True if the form is currently being submitted
     */
    loading: boolean
}

export class SiteAdminAddExternalServicePage extends React.Component<Props, State> {
    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            loading: false,
            kind: GQL.ExternalServiceKind.GITHUB,
            displayName: '',
            config: '{}',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('AddExternalService')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(event => {
                        event.preventDefault()
                        eventLogger.log('AddExternalServiceClicked')
                    }),
                    filter(event => event.currentTarget.checkValidity()),
                    mergeMap(() =>
                        this.addExternalService().pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    service => {
                        this.props.history.push(`/site-admin/external-services/${service.id}`)
                    },
                    error => {
                        console.error(error)
                    }
                )
        )
    }

    private addExternalService(): Observable<GQL.IExternalService> {
        return mutateGraphQL(
            gql`
                mutation addExternalService($input: AddExternalServiceInput!) {
                    addExternalService(input: $input) {
                        id
                    }
                }
            `,
            { input: this.state }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.addExternalService || (errors && errors.length > 0)) {
                    eventLogger.log('AddExternalServiceFailed')
                    throw createAggregateError(errors)
                }
                eventLogger.log('AddExternalServiceSucceeded', {
                    externalService: {
                        kind: data.addExternalService.kind,
                    },
                })
                return data.addExternalService
            })
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="add-external-service-page">
                <PageTitle title="Add external service" />
                <Form className="settings-form" onSubmit={this.onSubmit}>
                    <h1>Add a new external service</h1>
                    <p>Sourcegraph can synchronize data (e.g. code) from external services.</p>
                    {this.state.error && <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>}
                    <div className="form-group">
                        <label htmlFor="add-external-service-page__form-display-name">Display name</label>
                        <input
                            id="add-external-service-page__form-display-name"
                            type="text"
                            className="form-control"
                            placeholder="ACME private code"
                            required={true}
                            autoCorrect="off"
                            autoComplete="off"
                            autoFocus={true}
                            value={this.state.displayName}
                            onChange={this.onDisplayNameChange}
                            disabled={this.state.loading}
                            // aria-describedby="add-external-service-page__form-display-name-help"
                        />
                        {/* <small id="add-external-service-page__form-display-name-help" className="form-text text-muted">
                            A descriptive name of this external service.
                        </small> */}
                    </div>

                    <div className="form-group">
                        <label htmlFor="add-external-service-page__form-kind">Kind</label>
                        <input
                            id="add-external-service-page__form-kind"
                            type="text"
                            className="form-control"
                            placeholder="GITHUB"
                            required={true}
                            autoCorrect="off"
                            value={this.state.kind}
                            onChange={this.onKindChange}
                            disabled={this.state.loading}
                        />
                    </div>

                    <div className="form-group">
                        <label htmlFor="add-external-service-page__form-config">Config</label>
                        <input
                            id="add-external-service-page__form-config"
                            type="text"
                            className="form-control"
                            placeholder="{}"
                            required={true}
                            autoCorrect="off"
                            value={this.state.config}
                            onChange={this.onConfigChange}
                            disabled={this.state.loading}
                        />
                    </div>
                    <button type="submit" className="btn btn-primary" disabled={this.state.loading}>
                        Add external service
                    </button>
                    {this.state.loading && <LoadingSpinner className="icon-inline" />}
                </Form>
            </div>
        )
    }

    private onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.setState({ displayName: event.currentTarget.value })
    }

    private onKindChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        // TODO: validation of config
        this.setState({ kind: event.currentTarget.value as GQL.ExternalServiceKind })
    }

    private onConfigChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.setState({ config: event.currentTarget.value })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        this.submits.next(event)
    }
}
