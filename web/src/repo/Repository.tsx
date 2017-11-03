import ChevronLeftIcon from '@sourcegraph/icons/lib/ChevronLeft'
import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ListIcon from '@sourcegraph/icons/lib/List'
import * as H from 'history'
import * as React from 'react'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Position } from 'vscode-languageserver-types'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { ChromeExtensionToast, FirefoxExtensionToast } from '../marketing/BrowserExtensionToast'
import { SurveyToast } from '../marketing/SurveyToast'
import { IS_CHROME, IS_FIREFOX } from '../marketing/util'
import { ReferencesWidget } from '../references/ReferencesWidget'
import { viewEvents } from '../tracking/events'
import { Tree } from '../tree/Tree'
import { TreeHeader } from '../tree/TreeHeader'
import { parseHash } from '../util/url'
import { fetchHighlightedFile, listAllFiles } from './backend'
import { Blob } from './Blob'
import { DirectoryPage } from './DirectoryPage'
import { RepoNav } from './RepoNav'

export interface Props {
    repoPath: string
    rev?: string
    commitID: string
    filePath?: string
    location: H.Location
    history: H.History
    onToggleFullWidth: () => void
    isFullWidth: boolean
    phabricatorCallsign?: string
    isDirectory: boolean
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
        isDirectory: false,
    }
    private componentUpdates = new Subject<Props>()
    private showAnywayButtonClicks = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const parsedHash = parseHash(this.props.location.hash)
        this.state.isDirectory = props.isDirectory || !props.filePath
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

        // Transitions to routes with file should update file contents
        this.subscriptions.add(
            Observable.merge(
                this.componentUpdates.map(props => ({ ...props, showHighlightingAnyway: false })),
                this.showAnywayButtonClicks.map(() => ({ ...this.props, showHighlightingAnyway: true }))
            )
                .filter(props => !props.isDirectory && Boolean(props.filePath))
                .switchMap(props =>
                    fetchHighlightedFile({
                        repoPath: props.repoPath,
                        commitID: props.commitID,
                        filePath: props.filePath!,
                        disableTimeout: props.showHighlightingAnyway,
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
                    highlightingError: undefined,
                }),
                err => console.error(err)
                )
        )
        this.subscriptions.add(
            this.componentUpdates.subscribe(
                props => this.setState({
                    isDirectory: props.isDirectory || !props.filePath,
                }),
                err => console.error(err)
            )
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)

        const hash = parseHash(this.props.location.hash)
        viewEvents.Blob.log({ fileShown: Boolean(this.props.filePath), referencesShown: hash.modal === 'references' })
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)

        const thisHash = parseHash(nextProps.location.hash)
        const nextHash = parseHash(nextProps.location.hash)
        const showRefs = nextHash.modal === 'references'
        const position = nextHash.line ? { line: nextHash.line, character: nextHash.character || 0 } : undefined
        if (
            this.props.location.pathname !== nextProps.location.pathname ||
            this.props.location.search !== nextProps.location.search ||
            thisHash.modal !== nextHash.modal
        ) {
            viewEvents.Blob.log({ fileShown: Boolean(nextProps.filePath), referencesShown: showRefs })
        }
        this.setState({ showRefs, position })
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='repository'>
                <PageTitle title={this.getPageTitle()} />
                <RepoNav {...this.props} />
                {IS_CHROME &&
                    <ChromeExtensionToast />}
                {IS_FIREFOX &&
                    <FirefoxExtensionToast />}
                <SurveyToast />
                <div className='repository__content'>
                    <div id='explorer' className={'repository__sidebar' + (this.state.showTree ? ' repository__sidebar--open' : '')}>
                        <button type='button' className='btn btn-icon repository__sidebar-toggle' onClick={this.onTreeToggle}><ListIcon /></button>
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
                        <button type='button' className='btn btn-icon repository__full-width-toggle' onClick={this.props.onToggleFullWidth} title='toggle full width'>
                            {this.props.isFullWidth ? <ChevronLeftIcon /> : <ChevronRightIcon />}
                        </button>
                        {
                            this.state.isDirectory &&
                            <DirectoryPage repoPath={this.props.repoPath} commitID={this.props.commitID} rev={this.props.rev} filePath={this.props.filePath || ''} />
                        }
                        {
                            this.state.highlightingError &&
                            <HeroPage icon={ErrorIcon} title='' subtitle={'Error: ' + this.state.highlightingError.message} />
                        }
                        {
                            this.state.highlightedFile && this.state.highlightedFile.aborted &&
                            <p className='repository__blob-alert'>
                                <ErrorIcon className='icon-inline' />
                                Syntax highlighting for this file has been disabled because it took too long.
                                    (<a href='' onClick={this.handleShowAnywayButtonClick}>show anyway</a>)
                                    {/* NOTE: The above parentheses are so that the text renders literally as "(show anyway)" */}
                            </p>
                        }
                        {!this.state.isDirectory &&
                            (this.state.highlightedFile ?
                                <Blob {...this.props} filePath={this.props.filePath!} html={this.state.highlightedFile.html} /> :
                                /* render placeholder for layout before content is fetched */
                                <div className='repository__blob-placeholder'></div>)
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

    private handleShowAnywayButtonClick = (e: React.MouseEvent<HTMLElement>) => {
        e.preventDefault()
        this.showAnywayButtonClicks.next()
    }
}
