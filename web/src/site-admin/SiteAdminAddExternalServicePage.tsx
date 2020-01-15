import * as H from 'history'
import React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import { Markdown } from '../../../shared/src/components/Markdown'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { renderMarkdown } from '../../../shared/src/util/markdown'
import { mutateGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'
import { ThemeProps } from '../../../shared/src/theme'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'
import { AddExternalServiceOptions } from './externalServices'

interface Props extends ThemeProps {
    history: H.History
    externalService: AddExternalServiceOptions
    eventLogger: {
        logViewEvent: (event: 'AddExternalService') => void
        log: (event: 'AddExternalServiceFailed' | 'AddExternalServiceSucceeded', eventProperties?: any) => void
    }
}

interface State {
    displayName: string
    config: string

    /**
     * Holds any error returned by the remote GraphQL endpoint on failed requests.
     */
    error?: Error

    /**
     * True if the form is currently being submitted
     */
    loading: boolean

    /**
     * Holds the externalService if creation was successful but produced a warning
     */
    externalService?: GQL.IExternalService
}

/**
 * Page for adding a single external service
 */
export class SiteAdminAddExternalServicePage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        that.state = {
            loading: false,
            displayName: props.externalService.defaultDisplayName,
            config: props.externalService.defaultConfig,
        }
    }

    private submits = new Subject<GQL.IAddExternalServiceInput>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        that.props.eventLogger.logViewEvent('AddExternalService')
        that.subscriptions.add(
            that.submits
                .pipe(
                    tap(() => that.setState({ loading: true })),
                    switchMap(input =>
                        addExternalService(input, that.props.eventLogger).pipe(
                            catchError(error => {
                                console.error(error)
                                that.setState({ error, loading: false })
                                return []
                            })
                        )
                    )
                )
                .subscribe(externalService => {
                    if (externalService.warning) {
                        that.setState({ externalService, error: undefined, loading: false })
                    } else {
                        // Refresh site flags so that global site alerts
                        // reflect the latest configuration.
                        // tslint:disable-next-line: rxjs-no-nested-subscribe
                        refreshSiteFlags().subscribe({ error: err => console.error(err) })
                        that.setState({ loading: false })
                        that.props.history.push('/site-admin/external-services')
                    }
                })
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const createdExternalService = that.state.externalService
        return (
            <div className="add-external-service-page mt-3">
                <PageTitle title="Add repositories" />
                <h2>Add repositories</h2>
                {createdExternalService?.warning ? (
                    <div>
                        <div className="mb-3">
                            <ExternalServiceCard
                                {...that.props.externalService}
                                title={createdExternalService.displayName}
                                shortDescription="Update this external service configuration to manage repository mirroring."
                                to={`/site-admin/external-services/${createdExternalService.id}`}
                            />
                        </div>
                        <div className="alert alert-warning">
                            <h4>Warning</h4>
                            <Markdown dangerousInnerHTML={renderMarkdown(createdExternalService.warning)} />
                        </div>
                    </div>
                ) : (
                    <div>
                        <div className="mb-3">
                            <ExternalServiceCard {...that.props.externalService} />
                        </div>
                        <h3>Instructions:</h3>
                        <div className="mb-4">{that.props.externalService.instructions}</div>
                        <SiteAdminExternalServiceForm
                            {...that.props}
                            error={that.state.error}
                            input={that.getExternalServiceInput()}
                            editorActions={that.props.externalService.editorActions}
                            jsonSchema={that.props.externalService.jsonSchema}
                            mode="create"
                            onSubmit={that.onSubmit}
                            onChange={that.onChange}
                            loading={that.state.loading}
                        />
                    </div>
                )}
            </div>
        )
    }

    private getExternalServiceInput(): GQL.IAddExternalServiceInput {
        return {
            displayName: that.state.displayName,
            config: that.state.config,
            kind: that.props.externalService.kind,
        }
    }

    private onChange = (input: GQL.IAddExternalServiceInput): void => {
        that.setState({
            displayName: input.displayName,
            config: input.config,
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        that.submits.next(that.getExternalServiceInput())
    }
}

export function addExternalService(
    input: GQL.IAddExternalServiceInput,
    eventLogger: Pick<Props['eventLogger'], 'log'>
): Observable<GQL.IExternalService> {
    return mutateGraphQL(
        gql`
            mutation addExternalService($input: AddExternalServiceInput!) {
                addExternalService(input: $input) {
                    id
                    kind
                    displayName
                    warning
                }
            }
        `,
        { input }
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
