import GitHubIcon from '@sourcegraph/icons/lib/GitHub'
import PhabricatorIcon from '@sourcegraph/icons/lib/Phabricator'
import ShareIcon from '@sourcegraph/icons/lib/Share'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Position, Range } from 'vscode-languageserver-types'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchFileMetadata, FileMetadata } from '../backend'

interface Props {
    repo?: GQL.IRepository | null
    rev: string
    filePath?: string
    position?: Position
    range?: Range
}

interface State {
    file?: FileMetadata | undefined
}

/**
 * A repository header action that goes to the corresponding URL on an external code host.
 */
export class GoToCodeHostAction extends React.PureComponent<Props, State> {
    public state: State = {}

    private fileChanges = new Subject<string>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.fileChanges
                .pipe(
                    tap(() => {
                        if (this.state.file) {
                            this.setState({ file: undefined })
                        }
                    }),
                    switchMap(filePath => {
                        if (!this.props.repo || !filePath) {
                            return []
                        }
                        return fetchFileMetadata({
                            repoPath: this.props.repo.uri,
                            rev: this.props.rev,
                            filePath,
                        }).pipe(
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    })
                )
                .subscribe(
                    file => {
                        if (file) {
                            this.setState({
                                file,
                            })
                        }
                    },
                    err => console.error(err)
                )
        )

        this.fileChanges.next(this.props.filePath)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.filePath !== this.props.filePath) {
            this.fileChanges.next(props.filePath)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.props.repo) {
            return null
        }

        const rawURL = urlToCodeHost(this.props.repo, this.state.file)
        if (rawURL === null) {
            return null
        }
        const url = new URL(rawURL)

        let tooltip: string
        let label = ''
        let icon: JSX.Element | null = null

        switch (this.props.repo.hostType) {
            case 'GitHub':
            case 'GitHub Enterprise':
                tooltip = 'View on GitHub'
                icon = <GitHubIcon className="icon-inline" />
                if (this.props.range) {
                    url.hash = `#L${this.props.range.start.line}-L${this.props.range.end.line}`
                } else if (this.props.position) {
                    url.hash = '#L' + this.props.position.line
                }
                break
            case 'Phabricator':
                tooltip = 'View on Phabricator'
                icon = <PhabricatorIcon className="icon-inline" />
                break
            case 'gitlab':
                tooltip = 'View on GitLab'
                icon = <ShareIcon className="icon-inline" />
                break
            default:
                label = 'View on code host'
                tooltip = label
        }

        return (
            <a
                className="btn btn-link btn-sm composite-container__header-action"
                onClick={onClick}
                href={url.href}
                data-tooltip={tooltip}
            >
                {icon}
                <span className="composite-container__header-action-text">{label}</span>
            </a>
        )
    }
}

function onClick(): void {
    eventLogger.log('OpenInCodeHostClicked')
}

function urlToCodeHost(repo: GQL.IRepository, file?: FileMetadata): string | null {
    if (file && file.url) {
        return file.url
    }
    if (!file && repo.url) {
        return repo.url
    }
    return null
}
