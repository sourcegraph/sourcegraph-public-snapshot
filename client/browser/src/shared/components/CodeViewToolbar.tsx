import H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ActionsNavItems } from '../../../../../shared/src/actions/ActionsNavItems'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ISite, IUser } from '../../../../../shared/src/graphql/schema'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { toURIWithPath } from '../../../../../shared/src/util/url'
import { FileInfo } from '../../libs/code_intelligence'
import { fetchCurrentUser, fetchSite } from '../backend/server'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className: string
    style: React.CSSProperties
    iconStyle?: React.CSSProperties
}

interface CodeViewToolbarProps extends PlatformContextProps, ExtensionsControllerProps, FileInfo {
    onEnabledChange?: (enabled: boolean) => void

    buttonProps: ButtonProps
    actionsNavItemClassProps?: {
        listItemClass?: string
        actionItemClass?: string
    }
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
            <div
                className="code-view-toolbar"
                style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}
            >
                <ul className={`nav ${this.props.platformContext ? 'pr-1' : ''}`}>
                    <ActionsNavItems
                        menu={ContributableMenu.EditorTitle}
                        extensionsController={this.props.extensionsController}
                        platformContext={this.props.platformContext}
                        listItemClass="BtnGroup"
                        actionItemClass="btn btn-sm tooltipped tooltipped-n BtnGroup-item"
                        location={this.props.location}
                        scope={{
                            type: 'textEditor',
                            item: {
                                uri: toURIWithPath(this.props),
                                languageId: getModeFromPath(this.props.filePath) || 'could not determine mode',
                            },
                            selections: [],
                        }}
                    />
                </ul>
                {this.props.baseCommitID &&
                    this.props.baseHasFileContents && (
                        <OpenOnSourcegraph
                            label={'View File (base)'}
                            ariaLabel="View file on Sourcegraph"
                            openProps={{
                                repoName: this.props.baseRepoName || this.props.repoName,
                                filePath: this.props.baseFilePath || this.props.filePath,
                                rev: this.props.baseRev || this.props.baseCommitID,
                                query: {
                                    diff: {
                                        rev: this.props.baseCommitID,
                                    },
                                },
                            }}
                            className={this.props.buttonProps.className}
                            style={this.props.buttonProps.style}
                            iconStyle={this.props.buttonProps.iconStyle}
                        />
                    )}

                {/*
                  Use a ternary here because prettier insists on changing parens resulting in this button only being rendered
                  if the condition after the || is satisfied.
                 */}
                {!this.props.baseCommitID || (this.props.baseCommitID && this.props.headHasFileContents) ? (
                    <OpenOnSourcegraph
                        label={`View File${this.props.baseCommitID ? ' (head)' : ''}`}
                        ariaLabel="View file on Sourcegraph"
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
                        className={this.props.buttonProps.className}
                        style={this.props.buttonProps.style}
                        iconStyle={this.props.buttonProps.iconStyle}
                    />
                ) : null}
            </div>
        )
    }
}
