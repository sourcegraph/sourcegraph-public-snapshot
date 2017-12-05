import * as React from 'react'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { refreshConfiguration, updateOrgSettings } from '../backend'
import { SettingsFile } from '../SettingsFile'

export interface Props {
    orgID: string
    settings: GQL.ISettings | null
    orgInEditorBeta: boolean
    isLightTheme: boolean

    /**
     * Called when the user saves changes to the settings file's contents.
     */
    onDidCommit: () => void
}

interface State {
    commitError?: Error
}

export class OrgSettingsFile extends React.PureComponent<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <div className="settings-file-container">
                <SettingsFile
                    settings={this.props.settings}
                    commitError={this.state.commitError}
                    onDidCommit={this.onDidCommit}
                    isLightTheme={this.props.isLightTheme}
                />
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/search#scope">
                        Customizing search scopes for org members
                    </a>
                </small>
                {this.props.orgInEditorBeta && (
                    <small className="form-text">
                        This configuration applies to all org members and takes effect in Sourcegraph Editor and on the
                        web. You can also run the 'Preferences: Open Organization Settings' command inside of
                        Sourcegraph Editor to change this configuration.
                    </small>
                )}
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string) =>
        updateOrgSettings(this.props.orgID, lastKnownSettingsID, contents)
            .pipe(tap(this.props.onDidCommit), switchMap(refreshConfiguration))
            .subscribe(
                () => this.setState({ commitError: undefined }),
                err => {
                    this.setState({ commitError: err })
                    console.error(err)
                }
            )
}
