/**
 * This component is deprecated and will be removed once the GitHub inject method is no longer in use and
 * code intelligence on GitHub is provided by the `code_intelligence` library.
 *
 * This file contains the code for the CodeViewToolbar before https://github.com/sourcegraph/browser-extensions/pull/32.
 * That PR contains changes that made it easier for abstract support that fits each code host (mainly just making Phabricator possible).
 * It wasn't worth making the changes backward compatible or changing the existing inject method to work with the new one.
 */

import { ActionsNavItems } from '@sourcegraph/extensions-client-common/lib/app/actions/ActionsNavItems'
import { ControllerProps } from '@sourcegraph/extensions-client-common/lib/client/controller'
import { ExtensionsProps } from '@sourcegraph/extensions-client-common/lib/context'
import { ISite, IUser } from '@sourcegraph/extensions-client-common/lib/schema/graphqlschema'
import {
    ConfigurationCascadeProps,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ContributableMenu } from 'sourcegraph/module/protocol'
import { SimpleProviderFns } from '../backend/lsp'
import { fetchCurrentUser, fetchSite } from '../backend/server'
import { CodeIntelStatusIndicator } from './CodeIntelStatusIndicator'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className: string
    style: React.CSSProperties
    iconStyle?: React.CSSProperties
}

interface CodeViewToolbarProps
    extends Partial<ExtensionsProps<ConfigurationSubject, Settings>>,
        Partial<ControllerProps<ConfigurationSubject, Settings>> {
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

interface CodeViewToolbarState extends ConfigurationCascadeProps<ConfigurationSubject, Settings> {
    site?: ISite
    currentUser?: IUser
}

export class CodeViewToolbar extends React.Component<CodeViewToolbarProps, CodeViewToolbarState> {
    public state: CodeViewToolbarState = {
        configurationCascade: { subjects: [], merged: {} },
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        if (this.props.extensions) {
            this.subscriptions.add(
                this.props.extensions.context.configurationCascade.subscribe(
                    configurationCascade => this.setState({ configurationCascade }),
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
        const { site, currentUser } = this.state
        return (
            <div style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}>
                <ul className={`nav ${this.props.extensions ? 'pr-1' : ''}`}>
                    {this.props.extensionsController &&
                        this.props.extensions && (
                            <div className="BtnGroup">
                                <ActionsNavItems
                                    menu={ContributableMenu.EditorTitle}
                                    extensionsController={this.props.extensionsController}
                                    extensions={this.props.extensions}
                                    listClass="BtnGroup"
                                    actionItemClass="btn btn-sm tooltipped tooltipped-n BtnGroup-item"
                                    location={this.props.location}
                                />
                            </div>
                        )}
                </ul>
                {!this.props.extensionsController && (
                    <CodeIntelStatusIndicator
                        key="code-intel-status"
                        userIsSiteAdmin={currentUser ? currentUser.siteAdmin : false}
                        repoPath={this.props.repoPath}
                        commitID={this.props.baseCommitID}
                        filePath={this.props.filePath}
                        onChange={this.props.onEnabledChange}
                        simpleProviderFns={this.props.simpleProviderFns}
                        site={site}
                    />
                )}
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
