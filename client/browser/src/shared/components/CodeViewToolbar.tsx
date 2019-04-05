import classNames from 'classnames'
import H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ActionNavItemsClassProps, ActionsNavItems } from '../../../../../shared/src/actions/ActionsNavItems'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ISite, IUser } from '../../../../../shared/src/graphql/schema'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { toURIWithPath } from '../../../../../shared/src/util/url'
import { FileInfo } from '../../libs/code_intelligence'
import { fetchCurrentUser, fetchSite } from '../backend/server'
import { OpenDiffOnSourcegraph } from './OpenDiffOnSourcegraph'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className?: string
    style?: React.CSSProperties
    iconStyle?: React.CSSProperties
}

export interface CodeViewToolbarClassProps extends ActionNavItemsClassProps {
    /**
     * Class name for the `<ul>` element wrapping all toolbar items
     */
    className?: string
}

export interface CodeViewToolbarProps
    extends PlatformContextProps<'forceUpdateTooltip'>,
        ExtensionsControllerProps,
        FileInfo,
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
        this.subscriptions.add(fetchSite().subscribe(site => this.setState(() => ({ site }))))
        this.subscriptions.add(fetchCurrentUser().subscribe(currentUser => this.setState(() => ({ currentUser }))))
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
                    scope={{
                        type: 'textEditor',
                        item: {
                            uri: toURIWithPath(this.props),
                            languageId: getModeFromPath(this.props.filePath) || 'could not determine mode',
                        },
                        selections: [],
                    }}
                />{' '}
                {this.props.baseCommitID && this.props.baseHasFileContents && (
                    <li className={classNames('code-view-toolbar__item', this.props.listItemClass)}>
                        <OpenDiffOnSourcegraph
                            label="View file diff"
                            ariaLabel="View file diff on Sourcegraph"
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
                                    baseRev: this.props.baseRev!,
                                    headRev: this.props.rev!,
                                },
                            }}
                        />
                    </li>
                )}{' '}
                {!this.props.baseCommitID && (
                    <li className={classNames('code-view-toolbar__item', this.props.listItemClass)}>
                        <OpenOnSourcegraph
                            label="View file"
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
