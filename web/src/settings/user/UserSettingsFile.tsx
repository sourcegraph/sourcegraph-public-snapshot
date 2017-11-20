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
                <h3>Configuration</h3>
                <SettingsFile settings={this.props.settings} onDidCommit={this.onDidCommit} />
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/search#scope">
                        Customizing search scopes
                    </a>
                </small>
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
