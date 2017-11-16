import ArrowUpParentIcon from '@sourcegraph/icons/lib/ArrowUpParent'
import FolderIcon from '@sourcegraph/icons/lib/Folder'
import formatDistance from 'date-fns/formatDistance'
import isEqual from 'lodash/isEqual'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { UserAvatar } from '../settings/user/UserAvatar'
import { parseCommitDateString } from '../util/time'
import { toPrettyRepoURL } from '../util/url'
import { toTreeURL } from '../util/url'
import { fetchDirTree } from './backend'
import { fetchFileCommitInfo } from './backend'
import { DirectoryPageEntry } from './DirectoryPageEntry'

interface Props {
    repoPath: string
    // filePath is a directory path in DirectoryPage. We call it filePath for consistency elsewhere.
    filePath: string
    commitID: string
    rev?: string
}

interface State {
    dirTree?: GQL.ITree
    dirCommitInfo?: GQL.ICommitInfo
}

export class DirectoryPage extends React.PureComponent<Props, State> {
    public state: State = {}
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    constructor(props: Props) {
        super(props)
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(isEqual),
                    tap(() => {
                        if (this.state.dirTree) {
                            this.setState({ dirTree: undefined })
                        }
                    }),
                    switchMap(props =>
                        fetchDirTree(props).pipe(
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    )
                )
                .subscribe(dirTree => this.setState({ dirTree }), err => console.error(err))
        )
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(isEqual),
                    tap(() => {
                        if (this.state.dirCommitInfo) {
                            this.setState({ dirCommitInfo: undefined })
                        }
                    }),
                    switchMap(props =>
                        fetchFileCommitInfo(props).pipe(
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    )
                )
                .subscribe(dirCommitInfo => this.setState({ dirCommitInfo }), err => console.error(err))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { dirTree } = this.state
        const lastCommit = this.state.dirCommitInfo
        const person = lastCommit && lastCommit.committer && lastCommit.committer.person && lastCommit.committer.person
        const date =
            lastCommit &&
            lastCommit.committer &&
            formatDistance(parseCommitDateString(lastCommit.committer.date), new Date(), { addSuffix: true })

        if (!dirTree) {
            return null
        }

        return (
            <div className="dir-page">
                <h1 className="dir-page__head">
                    <FolderIcon className="dir-page__icon icon-inline" />
                    <span>{this.getLastPathPart()}</span>
                </h1>
                <table className="dir-page__table table table-hover">
                    <thead>
                        <tr className="dir-page__head-row">
                            {/* empty tds set the structure for the rest of the table to follow */}
                            <td className="dir-page__head-commit-spacer-cell dir-page__empty-cell" />
                            <td className="dir-page__name-cell dir-page__empty-cell" />
                            <td className="dir-page__commit-message-cell dir-page__empty-cell" />
                            <td className="dir-page__committer-cell dir-page__empty-cell" />
                            <td className="dir-page__date-cell dir-page__empty-cell" />
                            <td className="dir-page__commit-hash-cell dir-page__empty-cell" />
                        </tr>
                        <tr>
                            <td className="dir-page__head-commit-spacer-cell" />
                            <td
                                className="dir-page__commit-message-cell"
                                colSpan={2}
                                title={lastCommit && lastCommit.message}
                            >
                                {lastCommit && lastCommit.message}
                            </td>
                            <td className="dir-page__committer-cell">
                                {person && <UserAvatar user={person} />}
                                {person && person.name}
                            </td>
                            <td>{date}</td>
                            <td
                                title={lastCommit && lastCommit.rev.substring(0, 7)}
                                className="dir-page__commit-hash-cell"
                            >
                                {lastCommit && (
                                    <Link
                                        to={toTreeURL({
                                            repoPath: this.props.repoPath,
                                            filePath: this.props.filePath,
                                            rev: lastCommit && lastCommit.rev,
                                        })}
                                        className="dir-page__commit-hash-link"
                                    >
                                        {lastCommit.rev.substring(0, 7)}
                                    </Link>
                                )}
                            </td>
                        </tr>
                    </thead>
                    <tbody>
                        {this.props.filePath ? (
                            <tr>
                                <td className="dir-page__return-arrow-cell" colSpan={window.innerWidth <= 1200 ? 4 : 6}>
                                    <span>
                                        <Link
                                            to={
                                                this.getParentPath()
                                                    ? toTreeURL({
                                                          repoPath: this.props.repoPath,
                                                          filePath: this.getParentPath(),
                                                          rev: this.props.rev,
                                                      })
                                                    : toPrettyRepoURL({
                                                          repoPath: this.props.repoPath,
                                                          rev: this.props.rev,
                                                      })
                                            }
                                        >
                                            <ArrowUpParentIcon className="icon-inline dir-page__return-arrow-link" /> ..
                                        </Link>
                                    </span>
                                </td>
                            </tr>
                        ) : null}
                        {dirTree.directories.map((dir, i) => (
                            <DirectoryPageEntry
                                isDirectory={true}
                                key={i}
                                repoPath={this.props.repoPath}
                                filePath={[this.props.filePath, dir.name].filter(s => !!s).join('/')}
                                commitID={this.props.commitID}
                                rev={this.props.rev}
                            />
                        ))}
                        {dirTree.files.map((file, i) => (
                            <DirectoryPageEntry
                                isDirectory={false}
                                key={i}
                                repoPath={this.props.repoPath}
                                filePath={[this.props.filePath, file.name].filter(s => !!s).join('/')}
                                commitID={this.props.commitID}
                                rev={this.props.rev}
                            />
                        ))}
                    </tbody>
                </table>
            </div>
        )
    }

    private getLastPathPart(): string {
        if (!this.props.filePath) {
            return this.props.repoPath.substr(this.props.repoPath.lastIndexOf('/') + 1)
        }
        return this.props.filePath.substr(this.props.filePath.lastIndexOf('/') + 1)
    }

    private getParentPath(): string {
        if (!this.props.filePath) {
            return ''
        }
        if (this.props.filePath.lastIndexOf('/') > -1) {
            return this.props.filePath.substr(0, this.props.filePath.lastIndexOf('/'))
        }
        return ''
    }
}
