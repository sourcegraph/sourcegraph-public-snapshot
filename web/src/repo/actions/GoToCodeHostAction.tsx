import * as H from 'history'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { parseBrowserRepoURL, ParsedRepoURI } from '..'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepoListConfig } from '../backend'

interface RepoListConfig {
    blobURL: string | null
    commitURL: string | null
    treeURL: string | null
    viewURL: string | null
}

interface Props {
    repo: string
    location: H.Location
}

interface State {
    repoListConfig?: RepoListConfig | null
}

/**
 * A repository header action that goes to the corresponding URL on an external code host.
 */
export class GoToCodeHostAction extends React.PureComponent<Props, State> {
    public state: State = {}

    private repoChanges = new Subject<string>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.repoChanges
                .pipe(
                    tap(() => {
                        if (this.state.repoListConfig) {
                            this.setState({ repoListConfig: null })
                        }
                    }),
                    switchMap(repo =>
                        fetchRepoListConfig({ repoPath: repo }).pipe(
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    config => {
                        if (config) {
                            this.setState({
                                repoListConfig: {
                                    commitURL: config.commitURL,
                                    viewURL: config.viewURL,
                                    treeURL: config.treeURL,
                                    blobURL: config.blobURL,
                                },
                            })
                        }
                    },
                    err => console.error(err)
                )
        )

        this.repoChanges.next(this.props.repo)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.repo !== this.props.repo) {
            this.repoChanges.next(props.repo)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.repoListConfig) {
            return null
        }

        const { repoPath, filePath, rev } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
        const isDirectory = location.pathname.includes('/-/tree') // TODO(sqs): hacky

        const url = urlToCodeHost(this.state.repoListConfig, isDirectory, { repoPath, filePath, rev })
        if (url === null) {
            return null
        }

        return (
            <a
                className="btn btn-link btn-sm composite-container__header-action"
                onClick={onClick}
                href={url}
                data-tooltip="View on code host"
            >
                <span className="composite-container__header-action-text">View on code host</span>
            </a>
        )
    }
}

function onClick(): void {
    eventLogger.log('OpenInCodeHostClicked')
}

function urlToCodeHost(
    config: RepoListConfig,
    isDirectory: boolean,
    { repoPath, filePath, rev }: ParsedRepoURI
): string | null {
    if (!filePath && repoPath && config.viewURL) {
        return config.viewURL
    }
    if (filePath && isDirectory && config.treeURL) {
        return config.treeURL.replace('{rev}', rev || '').replace('{path}', filePath)
    }
    if (filePath && !isDirectory && config.blobURL) {
        return config.blobURL.replace('{rev}', rev || '').replace('{path}', filePath)
    }
    return null
}
