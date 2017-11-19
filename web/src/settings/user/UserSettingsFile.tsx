import * as React from 'react'
import { concat } from 'rxjs/operators/concat'
import { switchMap } from 'rxjs/operators/switchMap'
import { fetchCurrentUser, updateUserSettings } from '../../auth'
import { SettingsFile } from '../SettingsFile'

interface Props {
    settings: GQL.ISettings | null
}

interface State {
    commitError?: Error
}

export class UserSettingsFile extends React.PureComponent<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <div className="settings-file-container">
                <h3>Current user configuration</h3>
                {this.props.settings && this.props.settings.configuration.highlighted ? (
                    <SettingsFile settings={this.props.settings} onDidCommit={this.onDidCommit} />
                ) : (
                    <p className="form-text">No user configuration settings exist yet.</p>
                )}
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string): void => {
        updateUserSettings(lastKnownSettingsID, contents)
            .pipe(switchMap(fetchCurrentUser), concat([null]))
            .subscribe(
                () => this.setState({ commitError: undefined }),
                err => {
                    this.setState({ commitError: err })
                    console.error(err)
                }
            )
    }
}
