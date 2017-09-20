import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ListIcon from '@sourcegraph/icons/lib/List'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import * as React from 'react'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/partition'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Position } from 'vscode-languageserver-types'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { ReferencesWidget } from '../references/ReferencesWidget'
import { Tree } from '../tree/Tree'
import { TreeHeader } from '../tree/TreeHeader'
import { parseHash } from '../util/url'
import { fetchHighlightedFile, listAllFiles } from './backend'
import { Blob } from './Blob'
import { RepoNav } from './RepoNav'
import { RevSwitcher } from './RevSwitcher'

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
     * show the rev switcher
     */
    showRevSwitcher: boolean
    /**
     * an array of file paths in the repository
     */
    files?: string[]
    /**
     * the highlighted file
     */
    highlightedFile?: GQL.IHighlightedFile
    /**
     * the current path is a directory
     */
    isDirectory: boolean
    /**
     * error preventing fetching file contents
     */
    highlightingError?: Error
    /**
     * the current position selected
     */
    position?: Position
}

export class Repository extends React.Component<Props, State> {
    public state: State = {
        showTree: true,
        showRefs: false,
        showRevSwitcher: false,
        isDirectory: false
    }
    private componentUpdates = new Subject<Props>()
    private showAnywayButtonClicks = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const parsedHash = parseHash(this.props.location.hash)
        this.state.showRefs = parsedHash.modal === 'references'
        this.state.position = parsedHash.line ? { line: parsedHash.line!, character: parsedHash.character || 0 } : undefined
        this.subscriptions.add(
            this.componentUpdates
                .switchMap(props =>
                    listAllFiles({ repoPath: props.repoPath, commitID: props.commitID })
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
                    fetchHighlightedFile({
                        repoPath: props.repoPath,
                        commitID: props.commitID,
                        filePath: props.filePath!,
                        disableTimeout: props.showHighlightingAnyway
                    })
                        .catch(err => {
                            this.setState({ highlightedFile: undefined, isDirectory: false, highlightingError: err })
                            console.error(err)
                            return []
                        })
                )
                .subscribe(
                    result => this.setState({
                        isDirectory: result.isDirectory,
                        // file contents for a directory is a textual representation of the directory tree;
                        // we prefer not to display that
                        highlightedFile: !result.isDirectory ? result.highlightedFile : undefined,
                        highlightingError: undefined
                    }),
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

        const parsedHash = parseHash(nextProps.location.hash)
        const showRefs = parsedHash.modal === 'references'
        const position = parsedHash.line ? { line: parsedHash.line, character: parsedHash.character || 0 } : undefined
        this.setState({ showRefs, position })
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='repository'>
                <PageTitle title={this.getPageTitle()} />
                <RepoNav {...this.props} onClickRevision={this.toggleRevSwitcher} />
                {this.state.showRevSwitcher && <RevSwitcher {...this.props} onClose={this.toggleRevSwitcher} />}
                <div className='repository__content'>
                    <div id='explorer' className={'repository__sidebar' + (this.state.showTree ? ' repository__sidebar--open' : '')}>
                        <button type='button' className='repository__sidebar-toggle' onClick={this.onTreeToggle}><ListIcon /></button>
                        <TreeHeader title='File Explorer' onDismiss={this.onTreeToggle} />
                        <Tree
                            repoPath={this.props.repoPath}
                            rev={this.props.rev}
                            history={this.props.history}
                            scrollRootSelector='#explorer'
                            selectedPath={this.props.filePath || ''}
                            paths={this.state.files || []}
                        />
                    </div>
                    <div className='repository__viewer'>
                        {
                            (!this.props.filePath || this.state.isDirectory) &&
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
                            this.state.showRefs && this.state.position &&
                                <ReferencesWidget {...{ ...this.props, filePath: this.props.filePath!, position: this.state.position! }} />
                        }
                    </div>
                </div>
            </div>
        )
    }

    private getPageTitle(): string {
        const repoPathSplit = this.props.repoPath.split('/')
        const repoStr = repoPathSplit.length > 2 ? repoPathSplit.slice(1).join('/') : this.props.repoPath
        if (this.props.filePath) {
            const fileOrDir = this.props.filePath.split('/').pop()
            return `${fileOrDir} - ${repoStr}`
        }
        return `${repoStr}`
    }

    private onTreeToggle = () => this.setState({ showTree: !this.state.showTree })

    /**
     * toggles display of the rev switcher
     */
    private toggleRevSwitcher = () => {
        this.setState({ showRevSwitcher: !this.state.showRevSwitcher })
    }

    private handleShowAnywayButtonClick = e => {
        e.preventDefault()
        this.showAnywayButtonClicks.next()
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
