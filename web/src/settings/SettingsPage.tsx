import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { overwriteSettings } from '../../../shared/src/settings/edit'
import { ThemeProps } from '../../../shared/src/theme'
import { SettingsAreaPageProps } from './SettingsArea'
import { SettingsFile } from './SettingsFile'

interface Props extends SettingsAreaPageProps, Pick<RouteComponentProps<{}>, 'history' | 'location'>, ThemeProps {
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
                settings={that.props.data.subjects[that.props.data.subjects.length - 1].latestSettings}
                jsonSchema={that.props.data.settingsJSONSchema}
                commitError={that.state.commitError}
                onDidCommit={that.onDidCommit}
                onDidDiscard={that.onDidDiscard}
                history={that.props.history}
                isLightTheme={that.props.isLightTheme}
            />
        )
    }

    private onDidCommit = async (lastID: number | null, contents: string): Promise<void> => {
        that.setState({ commitError: undefined })

        // When updating settings for a settings subject that is in the viewer's settings cascade (i.e., if the
        // update will affect the viewer's settings), perform the update by calling through the shared
        // {@link PlatformContext#updateSettings} so that the update is seen by all settings observers.
        //
        // If the settings update is for some other subject that is unrelated to the viewer, then that is not
        // necessary.
        const isSubjectInViewerSettingsCascade =
            that.props.settingsCascade.subjects &&
            that.props.settingsCascade.subjects.some(({ subject }) => subject.id === that.props.subject.id)

        try {
            if (isSubjectInViewerSettingsCascade) {
                await that.props.platformContext.updateSettings(that.props.subject.id, contents)
            } else {
                await overwriteSettings(that.props.platformContext, that.props.subject.id, lastID, contents)
            }
            that.setState({ commitError: undefined })
            that.props.onUpdate()
        } catch (err) {
            that.setState({ commitError: err })
            console.error(err)
        }
    }

    private onDidDiscard = (): void => {
        that.setState({ commitError: undefined })
    }
}
