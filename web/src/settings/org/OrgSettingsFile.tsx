import * as React from 'react'
import { tap } from 'rxjs/operators/tap'
import { updateOrgSettings } from '../backend'
import { SettingsFile } from '../SettingsFile'

export interface Props {
    orgID: string
    settings: GQL.ISettings | null

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
                <h3>Configuration</h3>
                <SettingsFile
                    settings={this.props.settings}
                    commitError={this.state.commitError}
                    onDidCommit={(lastKnownSettingsID: number | null, contents: string) =>
                        updateOrgSettings(this.props.orgID, lastKnownSettingsID, contents) // tslint:disable-next-line:jsx-no-lambda
                            .pipe(tap(this.props.onDidCommit))
                            .subscribe(
                                () => this.setState({ commitError: undefined }),
                                err => {
                                    this.setState({ commitError: err })
                                    console.error(err)
                                }
                            )
                    }
                />
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/search#scope">
                        Customizing search scopes for org members
                    </a>
                </small>
                <small className="form-text">
                    This configuration applies to all org members and takes effect in Sourcegraph Editor and on the web.
                    You can also run the 'Preferences: Open Organization Settings' command inside of Sourcegraph Editor
                    to change this configuration.
                </small>
            </div>
        )
    }
}
