import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import * as React from 'react'
import 'rxjs/add/observable/fromPromise'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from 'sourcegraph/components/HeroPage'
import { ReferencesWidget } from 'sourcegraph/references/ReferencesWidget'
import { fetchHighlightedFile, listAllFiles } from 'sourcegraph/repo/backend'
import { hideTooltip } from 'sourcegraph/repo/tooltips'
import { Tree, TreeHeader } from 'sourcegraph/tree/Tree'
import { getCodeCellsForAnnotation, highlightAndScrollToLine } from 'sourcegraph/util'
import * as url from 'sourcegraph/util/url'
import { Blob } from './Blob'
import { RepoNav } from './RepoNav'

export interface Props {
    repoPath: string
    rev?: string
    commitID: string
    filePath?: string
    location: H.Location
    history: H.History
}

interface State {
    /**
     * show the references panel
     */
    showRefs: boolean
    /**
     * show the file tree explorer
     */
    showTree: boolean
    /**
     * an array of file paths in the repository
     */
    files?: string[]
    /**
     * the highlighted file
     */
    highlightedFile?: GQL.IHighlightedFile
    /**
     * error preventing fetching file contents
     */
    highlightingError?: Error
}

export class Repository extends React.Component<Props, State> {
    public state: State = {
        showTree: true,
        showRefs: false
    }
    private componentUpdates = new Subject<Props>()
    private showAnywayButtonClicks = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const u = url.parseBlob()
        this.state.showRefs = Boolean(u.path && u.modal && u.modal === 'references')
        this.subscriptions.add(
            this.componentUpdates
                .switchMap(props =>
                    Observable.fromPromise(listAllFiles({ repoPath: props.repoPath, commitID: props.commitID }))
                        .catch(err => {
                            console.error(err)
                            return []
                        })
                )
                .subscribe(
                    (files: string[]) => this.setState({ files }),
                    err => console.error(err)
                )
        )

        const [contentUpdatesWithFile, contentUpdatesWithoutFile] = Observable.merge(
            this.componentUpdates.map(props => ({ ...props, showHighlightingAnyway: false })),
            this.showAnywayButtonClicks.map(() => ({ ...this.props, showHighlightingAnyway: true }))
        ).partition(props => Boolean(props.filePath))

        // Transitions to routes with file should update file contents
        this.subscriptions.add(
            contentUpdatesWithFile
                .switchMap(props =>
                    Observable.fromPromise(fetchHighlightedFile({
                        repoPath: props.repoPath,
                        commitID: props.commitID,
                        filePath: props.filePath!,
                        disableTimeout: props.showHighlightingAnyway
                    })).catch(err => {
                        this.setState({ highlightedFile: undefined, highlightingError: err })
                        console.error(err)
                        return []
                    })
                )
                .subscribe(
                    result => this.setState({ highlightedFile: result, highlightingError: undefined }),
                    err => console.error(err)
                )
        )
        // Transitions to routes without file should unset file contents
        this.subscriptions.add(
            contentUpdatesWithoutFile
                .subscribe(() => {
                    this.setState({ highlightedFile: undefined, highlightingError: undefined })
                })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
        const hash = url.parseHash(nextProps.location.hash)
        const showRefs = Boolean(nextProps.filePath && hash.modal && hash.modal === 'references')
        if (showRefs !== this.state.showRefs) {
            this.setState({ showRefs })
        }
        if (this.props.location.hash !== nextProps.location.hash && nextProps.history.action === 'POP') {
            // handle 'back' and 'forward'
            this.scrollToLine(nextProps)
        } else if (this.props.location.pathname !== nextProps.location.pathname) {
            hideTooltip() // clear tooltip when transitioning between files
            this.scrollToLine(nextProps)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='repository'>
                <RepoNav {...this.props} onClickNavigation={() => this.setState({ showTree: !this.state.showTree })} />
                <div className='repository__content'>
                    {
                        this.state.showTree &&
                            <div id='explorer' className='repository__sidebar'>
                                <TreeHeader title='Files' onDismiss={() => this.setState({ showTree: false })} />
                                <Tree
                                    scrollRootSelector='#explorer'
                                    selectedPath={this.props.filePath || ''}
                                    onSelectPath={this.selectTreePath}
                                    paths={this.state.files || []}
                                />
                            </div>
                    }
                    <div className='repository__viewer'>
                        {
                            !this.props.filePath &&
                                <HeroPage icon={RepoIcon} title={this.props.repoPath.split('/').slice(1).join('/')} subtitle='Select a file to begin browsing.' />
                        }
                        {
                            this.state.highlightingError &&
                                <HeroPage icon={ErrorIcon} title='' subtitle={'Error: ' + this.state.highlightingError.message} />
                        }
                        {
                            this.state.highlightedFile && this.state.highlightedFile.aborted &&
                                <p className='repository__blob-alert'>
                                        <ErrorIcon />
                                    Syntax highlighting for this file has been disabled because it took too long.
                                    (<a href='#' onClick={this.handleShowAnywayButtonClick}>show anyway</a>)
                                    {/* NOTE: The above parentheses are so that the text renders literally as "(show anyway)" */}
                                </p>
                        }
                        {
                            this.state.highlightedFile ?
                                <Blob {...this.props} filePath={this.props.filePath!} html={this.state.highlightedFile.html} /> :
                                /* render placeholder for layout before content is fetched */
                                <div className='repository__blob-placeholder'></div>
                        }
                        {
                            this.state.showRefs &&
                                <ReferencesWidget onDismiss={() => {
                                    const currURL = url.parseBlob()
                                    this.props.history.push(url.toBlob({ ...currURL, modal: undefined, modalMode: undefined }))
                                }} />
                        }
                    </div>
                </div>
            </div>
        )
    }

    private handleShowAnywayButtonClick = e => {
        e.preventDefault()
        this.showAnywayButtonClicks.next()
    }

    private selectTreePath = (path: string, isDir: boolean) => {
        if (!isDir) {
            this.props.history.push(url.toBlob({ uri: this.props.repoPath, rev: this.props.rev, path }))
        }
    }

    private scrollToLine = (props: Props) => {
        const line = url.parseHash(props.location.hash).line
        if (line) {
            highlightAndScrollToLine(props.history, props.repoPath,
                props.commitID, props.filePath!, line, getCodeCellsForAnnotation())
        }
    }
}

export class RepositoryCloneInProgress extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='repository'>
                <RepoNav {...this.props} />
                <HeroPage icon={RepoIcon} title={this.props.repoPath.split('/').slice(1).join('/')} subtitle='Cloning in progress' />
            </div>
        )
    }
}

export class RepositoryNotFound extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='repository'>
                <RepoNav {...this.props} />
                <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, the requested URL was not found.' />
            </div>
        )
    }
}
