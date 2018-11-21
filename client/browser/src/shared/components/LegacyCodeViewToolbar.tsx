/**
 * This component is deprecated and will be removed once the GitHub inject method is no longer in use and
 * code intelligence on GitHub is provided by the `code_intelligence` library.
 *
 * This file contains the code for the CodeViewToolbar before https://github.com/sourcegraph/browser-extensions/pull/32.
 * That PR contains changes that made it easier for abstract support that fits each code host (mainly just making Phabricator possible).
 * It wasn't worth making the changes backward compatible or changing the existing inject method to work with the new one.
 */

import H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ActionsNavItems } from '../../../../../shared/src/actions/ActionsNavItems'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { ControllerProps } from '../../../../../shared/src/extensions/controller'
import { ISite, IUser } from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { SimpleProviderFns } from '../backend/lsp'
import { fetchCurrentUser, fetchSite } from '../backend/server'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className: string
    style: React.CSSProperties
    iconStyle?: React.CSSProperties
}

interface CodeViewToolbarProps extends Partial<PlatformContextProps>, Partial<ControllerProps> {
    repoPath: string
    filePath: string

    baseCommitID: string
    baseRev?: string
    headCommitID?: string
    headRev?: string

    onEnabledChange?: (enabled: boolean) => void

    buttonProps: ButtonProps
    actionsNavItemClassProps?: {
        listClass?: string
        actionItemClass?: string
    }
    simpleProviderFns: SimpleProviderFns
    location: H.Location
}

interface CodeViewToolbarState extends SettingsCascadeProps {
    site?: ISite
    currentUser?: IUser
}

export class CodeViewToolbar extends React.Component<CodeViewToolbarProps, CodeViewToolbarState> {
    public state: CodeViewToolbarState = {
        settingsCascade: { subjects: [], final: {} },
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        if (this.props.platformContext) {
            this.subscriptions.add(
                this.props.platformContext.settingsCascade.subscribe(
                    settingsCascade => this.setState({ settingsCascade }),
                    err => console.error(err)
                )
            )
        }
        this.subscriptions.add(fetchSite().subscribe(site => this.setState(() => ({ site }))))
        this.subscriptions.add(fetchCurrentUser().subscribe(currentUser => this.setState(() => ({ currentUser }))))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}>
                <ul className={`nav ${this.props.platformContext ? 'pr-1' : ''}`}>
                    {this.props.extensionsController &&
                        this.props.platformContext && (
                            <div className="BtnGroup">
                                <ActionsNavItems
                                    menu={ContributableMenu.EditorTitle}
                                    extensionsController={this.props.extensionsController}
                                    platformContext={this.props.platformContext}
                                    listClass="BtnGroup"
                                    actionItemClass="btn btn-sm tooltipped tooltipped-n BtnGroup-item"
                                    location={this.props.location}
                                />
                            </div>
                        )}
                </ul>
                <OpenOnSourcegraph
                    label={`View File${this.props.headCommitID ? ' (base)' : ''}`}
                    ariaLabel="View file on Sourcegraph"
                    openProps={{
                        repoPath: this.props.repoPath,
                        filePath: this.props.filePath,
                        rev: this.props.baseRev || this.props.baseCommitID,
                        query: this.props.headCommitID
                            ? {
                                  diff: {
                                      rev: this.props.baseCommitID,
                                  },
                              }
                            : undefined,
                    }}
                    className={this.props.buttonProps.className}
                    style={this.props.buttonProps.style}
                    iconStyle={this.props.buttonProps.iconStyle}
                />
                {this.props.headCommitID && (
                    <OpenOnSourcegraph
                        label={'View File (head)'}
                        ariaLabel="View file on Sourcegraph"
                        openProps={{
                            repoPath: this.props.repoPath,
                            filePath: this.props.filePath,
                            rev: this.props.headRev || this.props.headCommitID,
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
            </div>
        )
    }
}
