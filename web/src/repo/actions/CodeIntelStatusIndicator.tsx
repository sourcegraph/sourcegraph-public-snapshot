import Loader from '@sourcegraph/icons/lib/Loader'
import { isEqual, upperFirst } from 'lodash'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import PowerPlugIcon from 'mdi-react/PowerPlugIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { forkJoin, Observable, Subject } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { ServerCapabilities } from 'vscode-languageserver-protocol'
import { AbsoluteRepoFile } from '..'
import { EMODENOTFOUND, fetchServerCapabilities } from '../../backend/lsp'
import { getModeFromPath } from '../../util'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { fetchLangServer } from './backend'

interface LangServer {
    displayName?: string
    homepageURL?: string
    issuesURL?: string
    capabilities: ServerCapabilities
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
        distinctUntilChanged((a, b) => a.language === b.language),
        switchMap(({ repoPath, commitID, filePath, language }) => {
            if (!language) {
                return [null]
            }
            return forkJoin(
                fetchLangServer(language),
                fetchServerCapabilities({ repoPath, commitID, filePath, language })
            ).pipe(
                map(([langServer, capabilities]): LangServer => ({
                    displayName: (langServer && langServer.displayName) || undefined,
                    homepageURL: (langServer && langServer.homepageURL) || undefined,
                    issuesURL: (langServer && langServer.issuesURL) || undefined,
                    capabilities,
                })),
                catchError(err => (err.code === EMODENOTFOUND ? [null] : [asError(err)]))
            )
        }),
        map(langServerOrError => ({ langServerOrError }))
    )

interface CodeIntelStatusIndicatorProps extends AbsoluteRepoFile {
    userIsSiteAdmin: boolean
    language?: string
}
interface CodeIntelStatusIndicatorState {
    /** The language server, error, undefined while loading or null if no langserver registered */

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
            this.props.language !== nextProps.language
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
        if (this.state.langServerOrError === null || isErrorLike(this.state.langServerOrError)) {
            return 'text-danger'
        }
        if (
            !this.state.langServerOrError.capabilities.hoverProvider ||
            !this.state.langServerOrError.capabilities.referencesProvider ||
            !this.state.langServerOrError.capabilities.definitionProvider
        ) {
            return 'text-warning'
        }
        return 'text-success'
    }

    public render(): React.ReactNode {
        const language = getModeFromPath(this.props.filePath)
        return (
            <div className="code-intel-status-indicator">
                <button
                    className={`btn btn-link btn-sm composite-container__header-action ${this.getButtonColorCSSClass()}`}
                >
                    <PowerPlugIcon className="icon-inline" />
                </button>
                <div className="code-intel-status-indicator__popover card">
                    <div className="card-body">
                        {this.state.langServerOrError === undefined ? (
                            <div className="text-center">
                                <Loader className="icon-inline" />
                            </div>
                        ) : isErrorLike(this.state.langServerOrError) ? (
                            <span className="text-danger">{upperFirst(this.state.langServerOrError.message)}</span>
                        ) : this.state.langServerOrError === null ? (
                            <>
                                <h3>No language server connected</h3>
                                Check{' '}
                                <a href="http://langserver.org/" target="_blank">
                                    langserver.org
                                </a>{' '}
                                for {language} language servers
                            </>
                        ) : (
                            <>
                                <h3>
                                    Connected to the <wbr />
                                    <a href={this.state.langServerOrError.homepageURL} target="_blank">
                                        {this.state.langServerOrError.displayName || language} language server
                                    </a>
                                </h3>
                                <ul className="list-unstyled">
                                    <CapabilityStatus
                                        label="Hover tooltips"
                                        provided={!!this.state.langServerOrError.capabilities.hoverProvider}
                                    />
                                    <CapabilityStatus
                                        label="Go to definition"
                                        provided={!!this.state.langServerOrError.capabilities.definitionProvider}
                                    />
                                    <CapabilityStatus
                                        label="Find all references"
                                        provided={!!this.state.langServerOrError.capabilities.referencesProvider}
                                    />
                                    <CapabilityStatus
                                        label="Find implementations"
                                        provided={!!this.state.langServerOrError.capabilities.implementationProvider}
                                    />
                                </ul>
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
