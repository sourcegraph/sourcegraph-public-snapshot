import { upperFirst } from 'lodash'
import ExportIcon from 'mdi-react/ExportIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import { Position, Range } from 'vscode-languageserver-types'
import * as GQL from '../../backend/graphqlschema'
import { ActionItem } from '../../components/ActionItem'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { PhabricatorIcon } from '../../util/icons' // TODO: Switch mdi icon
import { fetchFileExternalLinks } from '../backend'

interface Props {
    repo?: GQL.IRepository | null
    rev: string
    filePath?: string
    position?: Position
    range?: Range

    externalLinks?: GQL.IExternalLink[]
}

interface State {
    /**
     * The external links for the current file/dir, or undefined while loading, null while not
     * needed (because not viewing a file/dir), or an error.
     */
    fileExternalLinksOrError?: GQL.IExternalLink[] | null | ErrorLike
}

/**
 * A repository header action that goes to the corresponding URL on an external code host.
 */
export class GoToCodeHostAction extends React.PureComponent<Props, State> {
    public state: State = { fileExternalLinksOrError: null }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    distinctUntilChanged((a, b) => a.repo === b.repo && a.rev === b.rev && a.filePath === b.filePath),
                    switchMap(({ repo, rev, filePath }) => {
                        if (!repo || !filePath) {
                            return of<Pick<State, 'fileExternalLinksOrError'>>({ fileExternalLinksOrError: null })
                        }
                        return merge(
                            of({ fileExternalLinksOrError: undefined }),
                            fetchFileExternalLinks({ repoPath: repo.name, rev, filePath }).pipe(
                                catchError(err => [asError(err)]),
                                map(c => ({ fileExternalLinksOrError: c }))
                            )
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )
    }

    public componentWillReceiveProps(props: Props): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.props.repo) {
            return null
        }

        let externalURLs: GQL.IExternalLink[]
        if (this.props.externalLinks && this.props.externalLinks.length > 0) {
            externalURLs = this.props.externalLinks
        } else if (
            this.state.fileExternalLinksOrError === null ||
            this.state.fileExternalLinksOrError === undefined ||
            isErrorLike(this.state.fileExternalLinksOrError) ||
            this.state.fileExternalLinksOrError.length === 0
        ) {
            // If the external link for the more specific resource within the repository is loading or errored, use the
            // repository external link.
            externalURLs = this.props.repo.externalURLs
        } else {
            externalURLs = this.state.fileExternalLinksOrError
        }
        if (externalURLs.length === 0) {
            return null
        }

        // Only show the first external link for now.
        const externalURL = externalURLs[0]

        const { displayName, icon } = serviceTypeDisplayNameAndIcon(externalURL.serviceType)
        const Icon = icon || ExportIcon

        // Special-case for GitHub: add line numbers to URL.
        let url = externalURL.url
        if (externalURL.serviceType === 'github') {
            if (this.props.range) {
                url += `#L${this.props.range.start.line}-L${this.props.range.end.line}`
            } else if (this.props.position) {
                url += '#L' + this.props.position.line
            }
        }

        return (
            <ActionItem to={url} target="_self" data-tooltip={`View on ${displayName}`}>
                <Icon className="icon-inline" />
            </ActionItem>
        )
    }
}

function serviceTypeDisplayNameAndIcon(
    serviceType: string | null
): { displayName: string; icon?: React.ComponentType<{ className?: string }> } {
    switch (serviceType) {
        case 'github':
            return { displayName: 'GitHub', icon: GithubCircleIcon }
        case 'gitlab':
            return { displayName: 'GitLab' }
        case 'bitbucketserver':
            return { displayName: 'Bitbucket Server' }
        case 'phabricator':
            return { displayName: 'Phabricator', icon: PhabricatorIcon }
        case 'awscodecommit':
            return { displayName: 'AWS CodeCommit' }
    }
    return { displayName: serviceType ? upperFirst(serviceType) : 'code host' }
}
