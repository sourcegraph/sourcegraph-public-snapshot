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
import { CodeHostCard } from '../components/CodeHostCard'
import { CodeHostVariant, getCodeHost } from './externalServices'
import { SiteAdminCodeHostForm } from './SiteAdminCodeHostForm'

interface Props extends ThemeProps {
    history: H.History
    kind: GQL.CodeHostKind
    variant?: CodeHostVariant
    eventLogger: {
        logViewEvent: (event: 'AddCodeHost') => void
        log: (event: 'AddCodeHostFailed' | 'AddCodeHostSucceeded', eventProperties?: any) => void
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
    externalService?: GQL.ICodeHost
}

/**
 * Page for adding a single external service
 */
export class SiteAdminAddCodeHostPage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        const serviceKindMetadata = getCodeHost(this.props.kind, this.props.variant)
        this.state = {
            loading: false,
            displayName: serviceKindMetadata.defaultDisplayName,
            config: serviceKindMetadata.defaultConfig,
        }
    }

    private submits = new Subject<GQL.IAddCodeHostInput>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.props.eventLogger.logViewEvent('AddCodeHost')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(input =>
                        addCodeHost(input, this.props.eventLogger).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ error, loading: false })
                                return []
                            })
                        )
                    )
                )
                .subscribe(externalService => {
                    if (externalService.warning) {
                        this.setState({ externalService, error: undefined, loading: false })
                    } else {
                        // Refresh site flags so that global site alerts
                        // reflect the latest configuration.
                        // tslint:disable-next-line: rxjs-no-nested-subscribe
                        refreshSiteFlags().subscribe({ error: err => console.error(err) })
                        this.setState({ loading: false })
                        this.props.history.push('/site-admin/external-services')
                    }
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const kindMetadata = getCodeHost(this.props.kind, this.props.variant)
        const createdCodeHost = this.state.externalService
        return (
            <div className="add-external-service-page mt-3">
                <PageTitle title="Add external service" />
                <h1>Add external service</h1>
                {createdCodeHost?.warning ? (
                    <div>
                        <div className="mb-3">
                            <CodeHostCard
                                {...kindMetadata}
                                kind={this.props.kind}
                                title={createdCodeHost.displayName}
                                shortDescription="Update this external service configuration to manage repository mirroring."
                                to={`/site-admin/external-services/${createdCodeHost.id}`}
                            />
                        </div>
                        <div className="alert alert-warning">
                            <h4>Warning</h4>
                            <Markdown dangerousInnerHTML={renderMarkdown(createdCodeHost.warning)} />
                        </div>
                    </div>
                ) : (
                    <div>
                        <div className="mb-3">
                            <CodeHostCard {...kindMetadata} kind={this.props.kind} />
                        </div>
                        <div className="mb-4">{kindMetadata.longDescription}</div>
                        <SiteAdminCodeHostForm
                            {...this.props}
                            error={this.state.error}
                            input={this.getCodeHostInput()}
                            editorActions={kindMetadata.editorActions}
                            jsonSchema={kindMetadata.jsonSchema}
                            mode="create"
                            onSubmit={this.onSubmit}
                            onChange={this.onChange}
                            loading={this.state.loading}
                        />
                    </div>
                )}
            </div>
        )
    }

    private getCodeHostInput(): GQL.IAddCodeHostInput {
        return {
            displayName: this.state.displayName,
            config: this.state.config,
            kind: this.props.kind,
        }
    }

    private onChange = (input: GQL.IAddCodeHostInput): void => {
        this.setState({
            displayName: input.displayName,
            config: input.config,
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        this.submits.next(this.getCodeHostInput())
    }
}

function addCodeHost(
    input: GQL.IAddCodeHostInput,
    eventLogger: Pick<Props['eventLogger'], 'log'>
): Observable<GQL.ICodeHost> {
    return mutateGraphQL(
        gql`
            mutation addCodeHost($input: AddCodeHostInput!) {
                addCodeHost(input: $input) {
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
            if (!data || !data.addCodeHost || (errors && errors.length > 0)) {
                eventLogger.log('AddCodeHostFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('AddCodeHostSucceeded', {
                externalService: {
                    kind: data.addCodeHost.kind,
                },
            })
            return data.addCodeHost
        })
    )
}
