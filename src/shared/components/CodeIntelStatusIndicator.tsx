import { ISite } from '@sourcegraph/extensions-client-common/lib/schema/graphqlschema'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ServerCapabilities } from 'javascript-typescript-langserver/lib/request-type'
import { isEqual, upperFirst } from 'lodash'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import PowerPlugIcon from 'mdi-react/PowerPlugIcon'
import React from 'react'
import { Button } from 'reactstrap'
import { forkJoin, Observable, Subject } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { asError, ErrorLike, isErrorLike } from '../backend/errors'
import { EMODENOTFOUND, SimpleProviderFns } from '../backend/lsp'
import { isPhabricator } from '../context'
import { AbsoluteRepoFile } from '../repo'
import { fetchLangServer } from '../repo/backend'
import { getModeFromPath, sourcegraphUrl } from '../util/context'

interface LangServer {
    displayName?: string
    homepageURL?: string
    issuesURL?: string
    /** defaults to `false` */
    experimental?: boolean
    capabilities?: ServerCapabilities
}

export const isCodeIntelligenceEnabled = (filePath: string) => {
    const disabledFiles = localStorage.getItem('disabledCodeIntelligenceFiles') || '{}'
    return !JSON.parse(disabledFiles)[`${window.location.pathname}:${filePath}`]
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
        map(({ filePath, ...rest }) => ({ ...rest, filePath, language: getModeFromPath(filePath) })),
        distinctUntilChanged((a, b) => a.language === a.language),
        switchMap(({ repoPath, commitID, filePath, language, simpleProviderFns }) => {
            if (!language) {
                return [null]
            }
            return forkJoin(
                fetchLangServer(language),
                simpleProviderFns.fetchServerCapabilities({ repoPath, commitID, filePath, language })
            ).pipe(
                map(
                    ([langServer, capabilities]): LangServer => ({
                        displayName: (langServer && langServer.displayName) || undefined,
                        homepageURL: (langServer && langServer.homepageURL) || undefined,
                        issuesURL: (langServer && langServer.issuesURL) || undefined,
                        experimental: (langServer && langServer.experimental) || undefined,
                        capabilities,
                    })
                ),
                catchError(err => (err.code === EMODENOTFOUND ? [null] : [asError(err)]))
            )
        }),
        map(langServerOrError => ({ langServerOrError }))
    )

interface CodeIntelStatusIndicatorProps extends AbsoluteRepoFile {
    userIsSiteAdmin: boolean
    /**
     * Called whenever the enabled status changed.
     * @param enabled is whether code intelligence is enabled or not.
     */
    onChange?: (enabled: boolean) => void
    simpleProviderFns: SimpleProviderFns
    site?: ISite
}
interface CodeIntelStatusIndicatorState {
    /** The language server, error, undefined while loading or null if no langserver registered. */
    langServerOrError?: LangServer | ErrorLike | null
    /** Whether code intelligence is toggled on or off. */
    enabled: boolean
}
export class CodeIntelStatusIndicator extends React.Component<
    CodeIntelStatusIndicatorProps,
    CodeIntelStatusIndicatorState
> {
    private componentUpdates = new Subject<CodeIntelStatusIndicatorProps>()
    private subscription = this.componentUpdates
        .pipe(propsToStateUpdate)
        .subscribe(stateUpdate => this.setState(stateUpdate))

    constructor(props: CodeIntelStatusIndicatorProps) {
        super(props)

        this.state = {
            enabled: isCodeIntelligenceEnabled(props.filePath),
        }
    }

    public shouldComponentUpdate(
        nextProps: CodeIntelStatusIndicatorProps,
        nextState: CodeIntelStatusIndicatorState
    ): boolean {
        return (
            !isEqual(this.state, nextState) ||
            this.props.userIsSiteAdmin !== nextProps.userIsSiteAdmin ||
            getModeFromPath(this.props.filePath) !== getModeFromPath(nextProps.filePath)
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(
        prevProps: CodeIntelStatusIndicatorProps,
        prevState: CodeIntelStatusIndicatorState
    ): void {
        this.componentUpdates.next(this.props)

        if (prevState.enabled !== this.state.enabled) {
            const disabledCodeIntelStorage = localStorage.getItem('disabledCodeIntelligenceFiles') || '{}'
            const disabledFiles = JSON.parse(disabledCodeIntelStorage)

            if (this.state.enabled) {
                delete disabledFiles[`${window.location.pathname}:${this.props.filePath}`]
            } else {
                disabledFiles[`${window.location.pathname}:${this.props.filePath}`] = true
            }
            localStorage.setItem('disabledCodeIntelligenceFiles', JSON.stringify(disabledFiles))

            if (this.props.onChange) {
                this.props.onChange(this.state.enabled)
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscription.unsubscribe()
    }

    private getButtonColorCSSClass(): string {
        if (!this.state.enabled) {
            return 'text-dark'
        }

        if (this.state.langServerOrError === undefined) {
            return ''
        }
        if (this.state.langServerOrError === null || isErrorLike(this.state.langServerOrError)) {
            return 'text-danger'
        }
        if (
            !this.state.langServerOrError.capabilities ||
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

    private toggleCodeIntelligence = () => {
        this.setState(({ enabled }) => ({ enabled: !enabled }))
    }

    private handleNoCodeIntelligence = (language?: string) => {
        const { site, userIsSiteAdmin } = this.props
        if (site) {
            const { hasCodeIntelligence } = site
            if (!hasCodeIntelligence) {
                if (!userIsSiteAdmin) {
                    return (
                        <>
                            <h3>Code intelligence disabled</h3>
                            Code intelligence is disabled. Contact your site admin to enable language servers. Code
                            intelligence is available for open source repositories.
                        </>
                    )
                }

                return (
                    <>
                        <h3>Code intelligence disabled</h3>
                        Enable code intelligence for jump to definition, hover tooltips, and find references.
                        <div>
                            <Button
                                color="primary"
                                size="sm"
                                className="mt-1"
                                href={`${sourcegraphUrl}/site-admin/code-intelligence`}
                                target="_blank"
                            >
                                Enable
                            </Button>
                        </div>
                    </>
                )
            }
        }
        return (
            <>
                <h3>No language server connected</h3>
                Check{' '}
                <a href="http://langserver.org/" target="_blank">
                    langserver.org
                </a>{' '}
                for {language} language servers
            </>
        )
    }

    public render(): JSX.Element {
        const language = getModeFromPath(this.props.filePath)
        const buttonClass = isPhabricator
            ? 'button grey button-grey has-icon has-text phui-button-default  msl'
            : 'btn btn-sm mr-1 aui-button'
        return (
            <div className="code-intel-status-indicator">
                <button
                    onClick={this.toggleCodeIntelligence}
                    className={`${buttonClass} composite-container__header-action ${this.getButtonColorCSSClass()}`}
                >
                    <PowerPlugIcon className="composite-container__icon icon-inline" />
                </button>
                <div className="code-intel-status-indicator__popover card">
                    <div className="card-body">
                        {this.state.langServerOrError === undefined ? (
                            <div className="text-center">
                                <LoadingSpinner className="icon-inline" />
                            </div>
                        ) : isErrorLike(this.state.langServerOrError) ? (
                            <span className="text-danger">{upperFirst(this.state.langServerOrError.message)}</span>
                        ) : this.state.langServerOrError === null ? (
                            this.handleNoCodeIntelligence(language)
                        ) : (
                            <>
                                <h3>
                                    Connected to the <wbr />
                                    <a href={this.state.langServerOrError.homepageURL} target="_blank">
                                        {this.state.langServerOrError.displayName || language} language server
                                    </a>
                                </h3>
                                <h4 className="mt-2 mb-0">Provides:</h4>
                                <ul className="code-intel-status-indicator__unstyled-list list-unstyled">
                                    <CapabilityStatus
                                        label="Hovers"
                                        provided={
                                            !!this.state.langServerOrError.capabilities &&
                                            !!this.state.langServerOrError.capabilities.hoverProvider
                                        }
                                    />
                                    <CapabilityStatus
                                        label="Definitions"
                                        provided={
                                            !!this.state.langServerOrError.capabilities &&
                                            !!this.state.langServerOrError.capabilities.definitionProvider
                                        }
                                    />
                                    <CapabilityStatus
                                        label="References"
                                        provided={
                                            !!this.state.langServerOrError.capabilities &&
                                            !!this.state.langServerOrError.capabilities.referencesProvider
                                        }
                                    />
                                    <CapabilityStatus
                                        label="Implementations"
                                        provided={
                                            !!this.state.langServerOrError.capabilities &&
                                            !!(this.state.langServerOrError.capabilities as any).implementationProvider
                                        }
                                    />
                                </ul>
                                <h4 className="mt-2 mb-0">Scope:</h4>
                                <ul className="code-intel-status-indicator__unstyled-list list-unstyled">
                                    <CapabilityStatus label="Local" provided={true} />
                                    <CapabilityStatus
                                        label="Cross-repository"
                                        provided={
                                            !!this.state.langServerOrError.capabilities &&
                                            hasCrossRepositoryCodeIntelligence(
                                                this.state.langServerOrError.capabilities
                                            )
                                        }
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
                                        <a href={`${sourcegraphUrl}/site-admin/code-intelligence`}>Manage</a>
                                    </p>
                                )}
                                {this.state.langServerOrError.issuesURL && (
                                    <p className="mt-2 mb-0">
                                        <a href={this.state.langServerOrError.issuesURL} target="_blank">
                                            Report issue
                                        </a>
                                    </p>
                                )}
                                <p className="mt-2 mb-0">
                                    <a
                                        onClick={this.toggleCodeIntelligence}
                                        style={{ cursor: 'pointer' }}
                                        target="_blank"
                                    >
                                        {this.state.enabled ? 'Disable' : 'Enable'} code intelligence
                                    </a>
                                </p>
                            </>
                        )}
                    </div>
                </div>
            </div>
        )
    }
}
