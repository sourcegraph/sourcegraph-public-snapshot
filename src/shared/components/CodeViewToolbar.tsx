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
import { FileInfo } from '../../libs/code_intelligence'
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
        Partial<ControllerProps<ConfigurationSubject, Settings>>,
        FileInfo {
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
            <div
                className="code-view-toolbar"
                style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}
            >
                <ul className={`nav ${this.props.extensions ? 'pr-1' : ''}`}>
                    {this.props.extensionsController &&
                        this.props.extensions && (
                            <ActionsNavItems
                                menu={ContributableMenu.EditorTitle}
                                extensionsController={this.props.extensionsController}
                                extensions={this.props.extensions}
                                listClass="BtnGroup"
                                actionItemClass="btn btn-sm tooltipped tooltipped-n BtnGroup-item"
                                location={this.props.location}
                            />
                        )}
                </ul>
                {!this.props.extensionsController && (
                    <CodeIntelStatusIndicator
                        key="code-intel-status"
                        userIsSiteAdmin={currentUser ? currentUser.siteAdmin : false}
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        filePath={this.props.filePath}
                        onChange={this.props.onEnabledChange}
                        simpleProviderFns={this.props.simpleProviderFns}
                        site={site}
                    />
                )}
                {this.props.baseCommitID &&
                    this.props.baseHasFileContents && (
                        <OpenOnSourcegraph
                            label={'View File (base)'}
                            ariaLabel="View file on Sourcegraph"
                            openProps={{
                                repoPath: this.props.baseRepoPath || this.props.repoPath,
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
                            repoPath: this.props.repoPath,
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
