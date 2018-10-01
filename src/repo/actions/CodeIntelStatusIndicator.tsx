import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ServerCapabilities } from 'javascript-typescript-langserver/lib/request-type'
import { isEqual, upperFirst } from 'lodash'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import PowerPlugIcon from 'mdi-react/PowerPlugIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { forkJoin, Observable, Subject } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { AbsoluteRepoFile } from '..'
import { ModeSpec } from '../../backend/features'
import { EMODENOTFOUND, fetchServerCapabilities } from '../../backend/lsp'
import { PLAINTEXT_MODE } from '../../util'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { fetchLangServer } from './backend'

interface LangServer {
    displayName?: string
    homepageURL?: string
    issuesURL?: string
    /** defaults to `false` */
    experimental: boolean
    /** `capabilities` is undefined when no language server is connected. */
    capabilities?: ServerCapabilities
}

function hasCrossRepositoryCodeIntelligence(capabilities: ServerCapabilities): boolean {
    return !!capabilities.xdefinitionProvider && !!capabilities.xworkspaceReferencesProvider
}

const CapabilityStatus: React.StatelessComponent<{ label: string; provided: boolean }> = ({ label, provided }) => (
    <li>
        {provided ? (
            <CheckIcon className="icon-inline text-success" />
        ) : (
            <CloseIcon className="icon-inline text-danger" />
        )}{' '}
        {label}
    </li>
)

const propsToStateUpdate = (obs: Observable<CodeIntelStatusIndicatorProps>) =>
    obs.pipe(
        distinctUntilChanged((a, b) => a.mode === b.mode),
        switchMap(({ repoPath, rev, commitID, filePath, mode }) => {
            if (!mode) {
                return [null]
            }
            return forkJoin(
                fetchLangServer(mode),
                fetchServerCapabilities({ repoPath, rev, commitID, filePath, mode }).pipe(
                    catchError(err => {
                        if (err.code === EMODENOTFOUND) {
                            return [undefined]
                        } else {
                            throw err
                        }
                    })
                )
            ).pipe(
                map(
                    ([langServer, capabilities]): LangServer | null =>
                        langServer && {
                            displayName: langServer.displayName,
                            homepageURL: langServer.homepageURL || undefined,
                            issuesURL: langServer.issuesURL || undefined,
                            experimental: langServer.experimental,
                            capabilities,
                        }
                ),
                catchError(err => [asError(err)])
            )
        }),
        map(langServerOrError => ({ langServerOrError }))
    )

interface CodeIntelStatusIndicatorProps extends AbsoluteRepoFile, ModeSpec {
    userIsSiteAdmin: boolean
}
interface CodeIntelStatusIndicatorState {
    /** The language server, error, undefined while loading or null if no language server registered */

    langServerOrError?: LangServer | ErrorLike | null
}
export class CodeIntelStatusIndicator extends React.Component<
    CodeIntelStatusIndicatorProps,
    CodeIntelStatusIndicatorState
> {
    public state: CodeIntelStatusIndicatorState = {}
    private componentUpdates = new Subject<CodeIntelStatusIndicatorProps>()
    private subscription = this.componentUpdates
        .pipe(propsToStateUpdate)
        .subscribe(stateUpdate => this.setState(stateUpdate))

    public shouldComponentUpdate(
        nextProps: CodeIntelStatusIndicatorProps,
        nextState: CodeIntelStatusIndicatorState
    ): boolean {
        return (
            !isEqual(this.state, nextState) ||
            this.props.userIsSiteAdmin !== nextProps.userIsSiteAdmin ||
            this.props.mode !== nextProps.mode
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(oldProps: CodeIntelStatusIndicatorProps): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscription.unsubscribe()
    }

    private getButtonColorCSSClass(): string {
        if (this.state.langServerOrError === undefined) {
            return ''
        }
        if (
            this.state.langServerOrError === null ||
            isErrorLike(this.state.langServerOrError) ||
            !this.state.langServerOrError.capabilities
        ) {
            return 'text-danger'
        }
        if (
            !this.state.langServerOrError.capabilities.hoverProvider ||
            !this.state.langServerOrError.capabilities.referencesProvider ||
            !this.state.langServerOrError.capabilities.definitionProvider ||
            this.state.langServerOrError.experimental ||
            !hasCrossRepositoryCodeIntelligence(this.state.langServerOrError.capabilities)
        ) {
            return 'text-warning'
        }
        return 'text-success'
    }

    public render(): React.ReactNode {
        return (
            <div className="code-intel-status-indicator">
                <a
                    className={`code-intel-status-indicator__icon nav-link ${this.getButtonColorCSSClass()}`}
                    tabIndex={0}
                >
                    <PowerPlugIcon className="icon-inline" />
                </a>
                <div className="code-intel-status-indicator__popover card" tabIndex={-1}>
                    <div className="card-body">
                        {this.state.langServerOrError === undefined ? (
                            <div className="text-center">
                                <LoadingSpinner className="icon-inline" />
                            </div>
                        ) : isErrorLike(this.state.langServerOrError) ? (
                            <span className="text-danger">{upperFirst(this.state.langServerOrError.message)}</span>
                        ) : this.state.langServerOrError === null ? (
                            this.props.mode === PLAINTEXT_MODE ? (
                                'No code intelligence available on plain text files.'
                            ) : (
                                <>
                                    <h3>No language server connected</h3>
                                    Check{' '}
                                    <a href="http://langserver.org/" target="_blank">
                                        langserver.org
                                    </a>{' '}
                                    for {this.props.mode} language servers.
                                </>
                            )
                        ) : !this.state.langServerOrError.capabilities ? (
                            <>
                                <h3>
                                    The {this.state.langServerOrError.displayName || this.props.mode} language server is
                                    disabled
                                </h3>
                                {this.props.userIsSiteAdmin ? (
                                    <>
                                        You can enable the {this.state.langServerOrError.displayName || this.props.mode}{' '}
                                        language server on the{' '}
                                        <Link to="/site-admin/code-intelligence">
                                            code intelligence administration page
                                        </Link>.
                                    </>
                                ) : (
                                    <>
                                        Ask your site admin to enable the{' '}
                                        {this.state.langServerOrError.displayName || this.props.mode} language server.
                                    </>
                                )}
                            </>
                        ) : (
                            <>
                                <h3>
                                    Connected to the <wbr />
                                    <a href={this.state.langServerOrError.homepageURL} target="_blank">
                                        {this.state.langServerOrError.displayName || this.props.mode} language server
                                    </a>
                                </h3>
                                <h4 className="mt-2 mb-0">Provides:</h4>
                                <ul className="list-unstyled">
                                    <CapabilityStatus
                                        label="Hovers"
                                        provided={!!this.state.langServerOrError.capabilities.hoverProvider}
                                    />
                                    <CapabilityStatus
                                        label="Definitions"
                                        provided={!!this.state.langServerOrError.capabilities.definitionProvider}
                                    />
                                    <CapabilityStatus
                                        label="References"
                                        provided={!!this.state.langServerOrError.capabilities.referencesProvider}
                                    />
                                    <CapabilityStatus
                                        label="Implementations"
                                        provided={!!this.state.langServerOrError.capabilities.implementationProvider}
                                    />
                                </ul>
                                <h4 className="mt-2 mb-0">Scope:</h4>
                                <ul className="list-unstyled">
                                    <CapabilityStatus label="Local" provided={true} />
                                    <CapabilityStatus
                                        label="Cross-repository"
                                        provided={hasCrossRepositoryCodeIntelligence(
                                            this.state.langServerOrError.capabilities
                                        )}
                                    />
                                </ul>
                                {this.state.langServerOrError.experimental && (
                                    <p className="mt-2 mb-0 text-warning font-weight-light">
                                        <em>
                                            This language server is experimental - some code intelligence actions might
                                            not work correctly.
                                        </em>
                                    </p>
                                    // TODO - Add docs link about experimental code intelligence when written
                                )}
                                {this.props.userIsSiteAdmin && (
                                    <p className="mt-2 mb-0">
                                        <Link to="/site-admin/code-intelligence">Manage</Link>
                                    </p>
                                )}
                                {this.state.langServerOrError.issuesURL && (
                                    <p className="mt-2 mb-0">
                                        <a href={this.state.langServerOrError.issuesURL} target="_blank">
                                            Report issue
                                        </a>
                                    </p>
                                )}
                            </>
                        )}
                    </div>
                </div>
            </div>
        )
    }
}
