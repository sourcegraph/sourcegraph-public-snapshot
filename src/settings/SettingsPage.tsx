import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, mergeMap } from 'rxjs/operators'
import { overwriteSettings } from '../configuration/backend'
import { refreshConfiguration } from '../user/settings/backend'
import { SettingsAreaPageProps } from './SettingsArea'
import { SettingsFile } from './SettingsFile'

interface Props extends SettingsAreaPageProps, Pick<RouteComponentProps<{}>, 'history' | 'location'> {
    isLightTheme: boolean

    /** Optional description to render above the editor. */
    description?: JSX.Element
}

interface State {
    commitError?: Error
}

/**
 * Displays a page where the settings for a subject can be edited.
 */
export class SettingsPage extends React.PureComponent<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <SettingsFile
                settings={this.props.data.subjects[this.props.data.subjects.length - 1].latestSettings}
                jsonSchema={this.props.data.settingsJSONSchema}
                commitError={this.state.commitError}
                onDidCommit={this.onDidCommit}
                onDidDiscard={this.onDidDiscard}
                history={this.props.history}
                isLightTheme={this.props.isLightTheme}
            />
        )
    }

    private onDidCommit = (lastID: number | null, contents: string) => {
        this.setState({ commitError: undefined })
        overwriteSettings(this.props.subject.id, lastID, contents)
            .pipe(mergeMap(() => refreshConfiguration().pipe(concat([null]))))
            .subscribe(
                () => {
                    this.setState({ commitError: undefined })
                    this.props.onUpdate()
                },
                err => {
                    this.setState({ commitError: err })
                    console.error(err)
                }
            )
    }

    private onDidDiscard = (): void => {
        this.setState({ commitError: undefined })
    }
}
