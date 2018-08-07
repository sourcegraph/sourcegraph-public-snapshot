import * as React from 'react'

import storage from '../../extension/storage'
import { setServerUrls } from '../util/context'
import { CodeIntelStatusIndicator } from './CodeIntelStatusIndicator'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className: string
    style: React.CSSProperties
    iconStyle?: React.CSSProperties
}

interface CodeViewToolbarProps {
    repoPath: string
    filePath: string

    baseCommitID: string
    baseRev?: string
    headCommitID?: string
    headRev?: string

    onEnabledChange?: (enabled: boolean) => void

    buttonProps: ButtonProps
}

export class CodeViewToolbar extends React.PureComponent<CodeViewToolbarProps> {
    public componentDidMount(): void {
        storage.onChanged(items => {
            if (items.serverUrls && items.serverUrls.newValue) {
                setServerUrls(items.serverUrls.newValue)
            }
        })
    }

    public render(): JSX.Element | null {
        return (
            <div style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}>
                <CodeIntelStatusIndicator
                    key="code-intel-status"
                    userIsSiteAdmin={false}
                    repoPath={this.props.repoPath}
                    commitID={this.props.baseCommitID}
                    filePath={this.props.filePath}
                    onChange={this.props.onEnabledChange}
                />
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
