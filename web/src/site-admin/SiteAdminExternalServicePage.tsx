import { parse as parseJSONC } from '@sqs/jsonc-parser'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, mergeMap, startWith, switchMap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'
import { ErrorAlert } from '../components/alerts'
import { defaultExternalServices, codeHostExternalServices } from './externalServices'
import { hasProperty } from '../../../shared/src/util/types'
import * as H from 'history'
import { CopyableText } from '../components/CopyableText'

type ExternalService = Pick<GQL.IExternalService, 'id' | 'kind' | 'displayName' | 'config' | 'warning' | 'webhookURL'>

interface Props extends RouteComponentProps<{ id: GQL.ID }> {
    isLightTheme: boolean
    history: H.History
}

const LOADING = 'loading' as const

interface State {
    externalServiceOrError: typeof LOADING | ExternalService | ErrorLike

    /**
     * The result of updating the external service: null when complete or not started yet,
     * loading, or an error.
     */
    updatedOrError: null | true | typeof LOADING | ErrorLike

    warning?: string
}

export class SiteAdminExternalServicePage extends React.Component<Props, State> {
    public state: State = {
        externalServiceOrError: LOADING,
        updatedOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private submits = new Subject<GQL.IUpdateExternalServiceInput>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminExternalService')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => props.match.params.id),
                    distinctUntilChanged(),
                    switchMap(id =>
                        fetchExternalService(id).pipe(
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    ),
                    map(result => ({ externalServiceOrError: result }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(input =>
                        concat(
                            [{ updatedOrError: LOADING, warning: null }],
                            updateExternalService(input).pipe(
                                mergeMap(service =>
                                    service.warning
                                        ? of({
                                              warning: service.warning,
                                              externalServiceOrError: service,
                                              updatedOrError: null,
                                          })
                                        : concat(
                                              // Flash "updated" text
                                              of({ updatedOrError: true, externalServiceOrError: service }),
                                              // Hide "updated" text again after 1s
                                              of({ updatedOrError: null }).pipe(delay(1000))
                                          )
                                ),
                                catchError((error: Error) => [{ updatedOrError: asError(error) }])
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public render(): JSX.Element | null {
        let error: ErrorLike | undefined
        if (isErrorLike(this.state.updatedOrError)) {
            error = this.state.updatedOrError
        }

        const externalService =
            (!isErrorLike(this.state.externalServiceOrError) &&
                this.state.externalServiceOrError !== LOADING &&
                this.state.externalServiceOrError) ||
            undefined

        let externalServiceCategory = externalService && defaultExternalServices[externalService.kind]
        if (
            externalService &&
            [GQL.ExternalServiceKind.GITHUB, GQL.ExternalServiceKind.GITLAB].includes(externalService.kind)
        ) {
            const parsedConfig: unknown = parseJSONC(externalService.config)
            const url =
                typeof parsedConfig === 'object' &&
                parsedConfig !== null &&
                hasProperty('url')(parsedConfig) &&
                typeof parsedConfig.url === 'string'
                    ? new URL(parsedConfig.url)
                    : undefined
            // We have no way of finding out whether a externalservice of kind GITHUB is GitHub.com or GitHub enterprise, so we need to guess based on the URL.
            if (externalService.kind === GQL.ExternalServiceKind.GITHUB && url?.hostname !== 'github.com') {
                externalServiceCategory = codeHostExternalServices.ghe
            }
            // We have no way of finding out whether a externalservice of kind GITLAB is Gitlab.com or Gitlab self-hosted, so we need to guess based on the URL.
            if (externalService.kind === GQL.ExternalServiceKind.GITLAB && url?.hostname !== 'gitlab.com') {
                externalServiceCategory = codeHostExternalServices.gitlab
            }
        }

        return (
            <div className="site-admin-configuration-page">
                {externalService ? (
                    <PageTitle title={`External service - ${externalService.displayName}`} />
                ) : (
                    <PageTitle title="External service" />
                )}
                <h2>Update synced repositories</h2>
                {this.state.externalServiceOrError === LOADING && <LoadingSpinner className="icon-inline" />}
                {isErrorLike(this.state.externalServiceOrError) && (
                    <ErrorAlert
                        className="mb-3"
                        error={this.state.externalServiceOrError}
                        history={this.props.history}
                    />
                )}
                {externalServiceCategory && (
                    <div className="mb-3">
                        <ExternalServiceCard {...externalServiceCategory} />
                    </div>
                )}
                {externalService && externalServiceCategory && (
                    <SiteAdminExternalServiceForm
                        input={externalService}
                        editorActions={externalServiceCategory.editorActions}
                        jsonSchema={externalServiceCategory.jsonSchema}
                        error={error}
                        warning={this.state.warning}
                        mode="edit"
                        loading={this.state.updatedOrError === LOADING}
                        onSubmit={this.onSubmit}
                        onChange={this.onChange}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                )}
                {this.state.updatedOrError === true && (
                    <p className="alert alert-success user-settings-profile-page__alert">Updated!</p>
                )}
                {externalService?.webhookURL && (
                    <div className="alert alert-info">
                        <h3>Campaign webhooks</h3>
                        {externalService.kind === GQL.ExternalServiceKind.BITBUCKETSERVER ? (
                            <p>
                                <a
                                    href="https://docs.sourcegraph.com/admin/external_service/bitbucket_server#webhooks"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    Webhooks
                                </a>{' '}
                                will be created automatically on the configured Bitbucket Server instance. In case you
                                don't provide an admin token,{' '}
                                <a
                                    href="https://docs.sourcegraph.com/admin/external_service/bitbucket_server#manual-configuration"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    follow the docs on how to set up webhooks manually
                                </a>
                                .
                                <br />
                                To set up another webhook manually, use the following URL:
                            </p>
                        ) : (
                            <p>
                                Point{' '}
                                <a
                                    href="https://docs.sourcegraph.com/admin/external_service/github#webhooks"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    webhooks
                                </a>{' '}
                                for this code host connection at the following URL:
                            </p>
                        )}
                        <CopyableText
                            className="mb-2"
                            text={externalService.webhookURL}
                            size={externalService.webhookURL.length}
                        />
                        <p className="mb-0">
                            Note that only{' '}
                            <a
                                href="https://docs.sourcegraph.com/user/campaigns"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                Campaigns
                            </a>{' '}
                            make use of this webhook. To enable webhooks to trigger repository updates on Sourcegraph,{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/repo/webhooks"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                see the docs on how to use them
                            </a>
                            .
                        </p>
                    </div>
                )}
            </div>
        )
    }

    private onChange = (input: GQL.IAddExternalServiceInput): void => {
        this.setState(state => {
            if (isExternalService(state.externalServiceOrError)) {
                return { ...state, externalServiceOrError: { ...state.externalServiceOrError, ...input } }
            }
            return state
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        if (isExternalService(this.state.externalServiceOrError)) {
            this.submits.next(this.state.externalServiceOrError)
        }
    }
}

function isExternalService(
    externalServiceOrError: typeof LOADING | ExternalService | ErrorLike
): externalServiceOrError is GQL.IExternalService {
    return externalServiceOrError !== LOADING && !isErrorLike(externalServiceOrError)
}

const externalServiceFragment = gql`
    fragment externalServiceFields on ExternalService {
        id
        kind
        displayName
        config
        warning
        webhookURL
    }
`

function updateExternalService(input: GQL.IUpdateExternalServiceInput): Observable<ExternalService> {
    return mutateGraphQL(
        gql`
            mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
                updateExternalService(input: $input) {
                    ...externalServiceFields
                }
            }

            ${externalServiceFragment}
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateExternalService)
    )
}

function fetchExternalService(id: GQL.ID): Observable<ExternalService> {
    return queryGraphQL(
        gql`
            query ExternalService($id: ID!) {
                node(id: $id) {
                    __typename
                    ...externalServiceFields
                }
            }

            ${externalServiceFragment}
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('External service not found')
            }
            if (node.__typename !== 'ExternalService') {
                throw new Error(`Node is a ${node.__typename}, not a ExternalService`)
            }
            return node
        })
    )
}
