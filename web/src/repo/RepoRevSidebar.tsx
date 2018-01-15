import ListIcon from '@sourcegraph/icons/lib/List'
import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { makeRepoURI } from '.'
import { gql, queryGraphQL } from '../backend/graphql'
import { Resizable } from '../components/Resizable'
import { Tree } from '../tree/Tree'
import { TreeHeader } from '../tree/TreeHeader'
import { memoizeObservable } from '../util/memoize'

const fetchTree = memoizeObservable(
    (args: { repoPath: string; commitID: string }): Observable<string[]> =>
        queryGraphQL(
            gql`
                query FileTree($repoPath: String!, $commitID: String!) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            tree(recursive: true) {
                                files {
                                    name
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.tree ||
                    !data.repository.commit.tree.files
                ) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                return data.repository.commit.tree.files.map(file => file.name)
            })
        ),
    makeRepoURI
)

interface Props {
    repoPath: string
    rev: string | undefined
    commitID: string
    filePath: string
    defaultBranch: string
    className: string
    history: H.History
}

interface State {
    loading: boolean
    error?: string

    showSidebar: boolean

    /**
     * All file paths in the repository.
     */
    files?: string[]
}

/**
 * The sidebar for a specific repo revision that shows the list of files and directories.
 */
export class RepoRevSidebar extends React.PureComponent<Props, State> {
    private static STORAGE_KEY = 'repo-rev-sidebar-hidden'

    public state: State = {
        loading: true,
        showSidebar: localStorage.getItem(RepoRevSidebar.STORAGE_KEY) === null,
    }

    private specChanges = new Subject<{ repoPath: string; commitID: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Fetch repository revision.
        this.subscriptions.add(
            this.specChanges
                .pipe(switchMap(({ repoPath, commitID }) => fetchTree({ repoPath, commitID })))
                .subscribe(files => this.setState({ files }), err => this.setState({ error: err.message }))
        )
        this.specChanges.next({ repoPath: this.props.repoPath, commitID: this.props.commitID })
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.repoPath !== this.props.repoPath || props.commitID !== this.props.commitID) {
            this.specChanges.next({ repoPath: props.repoPath, commitID: props.commitID })
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.showSidebar) {
            return (
                <button
                    type="button"
                    className={`btn btn-icon repo-rev-sidebar-toggle ${this.props.className}-toggle`}
                    onClick={this.onSidebarToggle}
                    data-tooltip="Show file tree"
                >
                    <ListIcon />
                </button>
            )
        }

        return (
            <Resizable
                className="repo-rev-container__sidebar-resizable"
                handlePosition="right"
                storageKey="repo-rev-sidebar"
                defaultSize={256 /* px */}
                element={
                    <div
                        id="explorer"
                        className={`repo-rev-sidebar ${this.props.className} ${
                            this.state.showSidebar ? `repo-rev-sidebar--open ${this.props.className}--open` : ''
                        }`}
                    >
                        <TreeHeader title="Files" onDismiss={this.onSidebarToggle} />
                        {this.state.files && (
                            <Tree
                                repoPath={this.props.repoPath}
                                rev={this.props.rev}
                                history={this.props.history}
                                scrollRootSelector="#explorer"
                                selectedPath={this.props.filePath || ''}
                                paths={this.state.files}
                            />
                        )}
                    </div>
                }
            />
        )
    }

    private onSidebarToggle = () => {
        if (this.state.showSidebar) {
            localStorage.setItem(RepoRevSidebar.STORAGE_KEY, 'true')
        } else {
            localStorage.removeItem(RepoRevSidebar.STORAGE_KEY)
        }
        this.setState({ showSidebar: !this.state.showSidebar })
    }
}
