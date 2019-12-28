import * as H from 'history'
import * as React from 'react'
import { Link } from '../../../shared/src/components/Link'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { ExternalServiceKindMetadata } from '../site-admin/externalServices'
import { PageTitle } from '../components/PageTitle'
import { SiteAdminExternalServiceForm } from '../site-admin/SiteAdminExternalServiceForm'
import * as GQL from '../../../shared/src/graphql/schema'
import { Subject, Subscription } from 'rxjs'
import { switchMap, catchError, tap } from 'rxjs/operators'
import { addExternalService } from '../site-admin/SiteAdminAddExternalServicePage'
import { refreshSiteFlags } from '../site/backend'
import { ThemeProps } from '../../../shared/src/theme'
import { Markdown } from '../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../shared/src/util/markdown'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'

interface Props extends ThemeProps, ActivationProps {
    history: H.History
    externalService: ExternalServiceKindMetadata
    eventLogger: {
        logViewEvent: (event: 'AddExternalService') => void
        log: (event: 'AddExternalServiceFailed' | 'AddExternalServiceSucceeded', eventProperties?: any) => void
    }
}
interface State {
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

export class WelcomeAddExternalServicePage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            loading: false,
            config: props.externalService.defaultConfig,
        }
    }

    private submits = new Subject<GQL.IAddExternalServiceInput>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.props.eventLogger.logViewEvent('AddExternalService')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(input =>
                        addExternalService(input, this.props.eventLogger).pipe(
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
                        if (this.props.activation) {
                            this.props.activation.update({ ConnectedCodeHost: true })
                        }
                        this.props.history.push('/onboard/guide')
                    }
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const createdExternalService = this.state.externalService
        return (
            <div className="welcome-page-left">
                <div className="welcome-page-left__buttons">
                    <Link className="btn btn-secondary welcome-page-left__back-button" to="/onboard/choose-code-host">
                        &lt; Back
                    </Link>
                    &nbsp;
                    <Link className="btn btn-secondary welcome-page-left__back-button" to="/onboard/guide">
                        Skip &gt;
                    </Link>
                </div>
                <div className="welcome-page-left__add-code-host-content mb-5">
                    <PageTitle title="Onboarding" />
                    <div>
                        <div className="mb-3">
                            <ExternalServiceCard {...this.props.externalService} />
                        </div>
                        {createdExternalService?.warning && (
                            <div className="alert alert-warning">
                                <h4>Warning</h4>
                                <Markdown dangerousInnerHTML={renderMarkdown(createdExternalService.warning)} />
                            </div>
                        )}
                        <h3>Follow these instructions to finish adding code to Sourcegraph:</h3>
                        <div className="mb-4">{this.props.externalService.longDescription}</div>
                        <SiteAdminExternalServiceForm
                            {...this.props}
                            error={this.state.error}
                            input={this.getExternalServiceInput()}
                            editorActions={this.props.externalService.editorActions}
                            jsonSchema={this.props.externalService.jsonSchema}
                            mode="create"
                            submitName="Next"
                            onSubmit={this.onSubmit}
                            onChange={this.onChange}
                            loading={this.state.loading}
                            hideDisplayNameField={true}
                        />
                    </div>
                </div>
            </div>
        )
    }

    private getExternalServiceInput(): GQL.IAddExternalServiceInput {
        return {
            displayName: this.props.externalService.defaultDisplayName,
            config: this.state.config,
            kind: this.props.externalService.kind,
            requireValidation: true,
        }
    }

    private onChange = (input: GQL.IAddExternalServiceInput): void => {
        this.setState({
            config: input.config,
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        this.submits.next(this.getExternalServiceInput())
    }
}
