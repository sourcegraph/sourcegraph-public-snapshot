import { Position, Range } from '@sourcegraph/extension-api-types'
import { upperFirst } from 'lodash'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import { PhabricatorIcon } from '../../../../shared/src/components/icons' // TODO: Switch mdi icon
import { LinkOrButton } from '../../../../shared/src/components/LinkOrButton'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { fetchFileExternalLinks } from '../backend'

interface Props {
    repo?: GQL.IRepository | null
    rev: string
    filePath?: string
    commitRange?: string
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
        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    startWith(that.props),
                    distinctUntilChanged((a, b) => a.repo === b.repo && a.rev === b.rev && a.filePath === b.filePath),
                    switchMap(({ repo, rev, filePath }) => {
                        if (!repo || !filePath) {
                            return of<Pick<State, 'fileExternalLinksOrError'>>({ fileExternalLinksOrError: null })
                        }
                        return merge(
                            of({ fileExternalLinksOrError: undefined }),
                            fetchFileExternalLinks({ repoName: repo.name, rev, filePath }).pipe(
                                catchError(err => [asError(err)]),
                                map(c => ({ fileExternalLinksOrError: c }))
                            )
                        )
                    })
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    err => console.error(err)
                )
        )
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        // If the default branch is undefined, set to HEAD
        const defaultBranch =
            (!isErrorLike(that.props.repo) &&
                that.props.repo &&
                that.props.repo.defaultBranch &&
                that.props.repo.defaultBranch.displayName) ||
            'HEAD'
        // If neither repo or file can be loaded, return null, which will hide all code host icons
        if (!that.props.repo || isErrorLike(that.state.fileExternalLinksOrError)) {
            return null
        }

        let externalURLs: GQL.IExternalLink[]
        if (that.props.externalLinks && that.props.externalLinks.length > 0) {
            externalURLs = that.props.externalLinks
        } else if (
            that.state.fileExternalLinksOrError === null ||
            that.state.fileExternalLinksOrError === undefined ||
            isErrorLike(that.state.fileExternalLinksOrError) ||
            that.state.fileExternalLinksOrError.length === 0
        ) {
            // If the external link for the more specific resource within the repository is loading or errored, use the
            // repository external link.
            externalURLs = that.props.repo.externalURLs
        } else {
            externalURLs = that.state.fileExternalLinksOrError
        }
        if (externalURLs.length === 0) {
            return null
        }

        // Only show the first external link for now.
        const externalURL = externalURLs[0]

        const { displayName, icon } = serviceTypeDisplayNameAndIcon(externalURL.serviceType)
        const Icon = icon || ExportIcon

        // Extract url to add branch, line numbers or commit range.
        let url = externalURL.url
        if (externalURL.serviceType === 'github' || externalURL.serviceType === 'gitlab') {
            // If in a branch, add branch path to the code host URL.
            if (that.props.rev && that.props.rev !== defaultBranch && !that.state.fileExternalLinksOrError) {
                url += `/tree/${that.props.rev}`
            }
            // If showing a comparison, add comparison specifier to the code host URL.
            if (that.props.commitRange) {
                url += `/compare/${that.props.commitRange.replace(/^\.\.\./, 'HEAD...').replace(/\.\.\.$/, '...HEAD')}`
            }
            // Add range or position path to the code host URL.
            if (that.props.range) {
                url += `#L${that.props.range.start.line}-L${that.props.range.end.line}`
            } else if (that.props.position) {
                url += '#L' + that.props.position.line
            }
        }

        return (
            <LinkOrButton to={url} target="_self" data-tooltip={`View on ${displayName}`}>
                <Icon className="icon-inline" />
            </LinkOrButton>
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
        case 'bitbucketServer':
            // TODO: Why is bitbucketServer (correctly) camelCase but
            // awscodecommit is (correctly) lowercase? Why is serviceType
            // not type-checked for validity?
            return { displayName: 'Bitbucket Server', icon: BitbucketIcon }
        case 'phabricator':
            return { displayName: 'Phabricator', icon: PhabricatorIcon }
        case 'awscodecommit':
            return { displayName: 'AWS CodeCommit' }
    }
    return { displayName: serviceType ? upperFirst(serviceType) : 'code host' }
}
