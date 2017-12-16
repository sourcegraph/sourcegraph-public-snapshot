import * as React from 'react'
import { concat } from 'rxjs/operators/concat'
import { switchMap } from 'rxjs/operators/switchMap'
import { fetchCurrentUser, updateUserSettings } from '../../auth'
import { SettingsFile } from '../SettingsFile'

interface Props {
    userInEditorBeta: boolean
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
                <SettingsFile settings={this.props.settings} onDidCommit={this.onDidCommit} />
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/search#scope">
                        Customizing search scopes
                    </a>
                </small>
                {this.props.userInEditorBeta && (
                    <small className="form-text">
                        Editor beta users: This configuration does not yet take effect in Sourcegraph Editor, unlike org
                        config (which does). It can only be used to configure the Sourcegraph web app.
                    </small>
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
