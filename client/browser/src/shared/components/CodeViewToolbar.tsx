import classNames from 'classnames'
import H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ActionNavItemsClassProps, ActionsNavItems } from '../../../../../shared/src/actions/ActionsNavItems'
import { ContributionScope } from '../../../../../shared/src/api/client/context/context'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ISite, IUser } from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { FileInfoWithContents } from '../../libs/code_intelligence/code_views'
import { fetchCurrentUser, fetchSite } from '../backend/server'
import { OpenDiffOnSourcegraph } from './OpenDiffOnSourcegraph'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className?: string
}

export interface CodeViewToolbarClassProps extends ActionNavItemsClassProps {
    /**
     * Class name for the `<ul>` element wrapping all toolbar items
     */
    className?: string

    /**
     * The scope of this toolbar (e.g., the view component that it is associated with).
     */
    scope?: ContributionScope
}

export interface CodeViewToolbarProps
    extends PlatformContextProps<'forceUpdateTooltip' | 'requestGraphQL'>,
        ExtensionsControllerProps,
        FileInfoWithContents,
        TelemetryProps,
        CodeViewToolbarClassProps {
    onEnabledChange?: (enabled: boolean) => void

    buttonProps?: ButtonProps
    location: H.Location
}

interface CodeViewToolbarState {
    site?: ISite
    currentUser?: IUser
}

export class CodeViewToolbar extends React.Component<CodeViewToolbarProps, CodeViewToolbarState> {
    public state: CodeViewToolbarState = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            fetchSite(this.props.platformContext.requestGraphQL).subscribe(site => this.setState(() => ({ site })))
        )
        this.subscriptions.add(
            fetchCurrentUser(this.props.platformContext.requestGraphQL).subscribe(currentUser =>
                this.setState(() => ({ currentUser }))
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <ul className={classNames('code-view-toolbar', this.props.className)}>
                <ActionsNavItems
                    {...this.props}
                    listItemClass={classNames('code-view-toolbar__item', this.props.listItemClass)}
                    menu={ContributableMenu.EditorTitle}
                    extensionsController={this.props.extensionsController}
                    platformContext={this.props.platformContext}
                    location={this.props.location}
                    scope={this.props.scope}
                />{' '}
                {this.props.baseCommitID && this.props.baseHasFileContents && (
                    <li className={classNames('code-view-toolbar__item', this.props.listItemClass)}>
                        <OpenDiffOnSourcegraph
                            ariaLabel="View file diff on Sourcegraph"
                            platformContext={this.props.platformContext}
                            className={this.props.actionItemClass}
                            iconClassName={this.props.actionItemIconClass}
                            openProps={{
                                repoName: this.props.baseRepoName || this.props.repoName,
                                filePath: this.props.baseFilePath || this.props.filePath,
                                rev: this.props.baseRev || this.props.baseCommitID,
                                query: {
                                    diff: {
                                        rev: this.props.baseCommitID,
                                    },
                                },
                                commit: {
                                    baseRev: this.props.baseRev || this.props.baseCommitID,
                                    headRev: this.props.rev || this.props.commitID,
                                },
                            }}
                        />
                    </li>
                )}{' '}
                {// Only show the "View file" button if we were able to fetch the file contents
                // from the Sourcegraph instance
                !this.props.baseCommitID && (this.props.content !== undefined || this.props.baseContent !== undefined) && (
                    <li className={classNames('code-view-toolbar__item', this.props.listItemClass)}>
                        <OpenOnSourcegraph
                            ariaLabel="View file on Sourcegraph"
                            className={this.props.actionItemClass}
                            iconClassName={this.props.actionItemIconClass}
                            openProps={{
                                repoName: this.props.repoName,
                                filePath: this.props.filePath,
                                rev: this.props.rev || this.props.commitID,
                                query: this.props.commitID
                                    ? {
                                          diff: {
                                              rev: this.props.commitID,
                                          },
                                      }
                                    : undefined,
                            }}
                        />
                    </li>
                )}
            </ul>
        )
    }
}
