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
                <h3>Editor settings</h3>
                {this.props.settings &&
                    this.props.settings.configuration.highlighted && [
                        <SettingsFile
                            key={0}
                            settings={this.props.settings}
                            commitError={this.state.commitError}
                            // tslint:disable-next-line:jsx-no-lambda
                            onDidCommit={(lastKnownSettingsID: number | null, contents: string) =>
                                updateOrgSettings(this.props.orgID, lastKnownSettingsID, contents)
                                    .pipe(tap(this.props.onDidCommit))
                                    .subscribe(
                                        () => this.setState({ commitError: undefined }),
                                        err => {
                                            this.setState({ commitError: err })
                                            console.error(err)
                                        }
                                    )
                            }
                        />,
                        <small key={1} className="form-text">
                            Run the 'Preferences: Open Organization Settings' command inside of Sourcegraph Editor to
                            change this configuration.
                        </small>,
                    ]}
                {!this.props.settings && (
                    <p className="form-text">
                        This organization hasn't created a configuration file yet. Run the 'Preferences: Open
                        Organization Settings' command inside of Sourcegraph Editor to create one.
                    </p>
                )}
            </div>
        )
    }
}
