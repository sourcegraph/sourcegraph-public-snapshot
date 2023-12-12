import * as React from 'react'

import { logger } from '@sourcegraph/common'
import { overwriteSettings } from '@sourcegraph/shared/src/settings/edit'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container } from '@sourcegraph/wildcard'

import type { SettingsAreaPageProps } from './SettingsArea'
import { SettingsFile } from './SettingsFile'

interface Props extends SettingsAreaPageProps, TelemetryProps {
    /** Optional description to render above the editor. */
    description?: JSX.Element
    isLightTheme: boolean
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
            <Container className="mb-3">
                <SettingsFile
                    settings={this.props.data.subjects[this.props.data.subjects.length - 1].latestSettings}
                    commitError={this.state.commitError}
                    onDidCommit={this.onDidCommit}
                    onDidDiscard={this.onDidDiscard}
                    isLightTheme={this.props.isLightTheme}
                    telemetryService={this.props.telemetryService}
                    telemetryRecorder={this.props.telemetryRecorder}
                />
            </Container>
        )
    }

    private onDidCommit = async (lastID: number | null, contents: string): Promise<void> => {
        this.setState({ commitError: undefined })

        // When updating settings for a settings subject that is in the viewer's settings cascade (i.e., if the
        // update will affect the viewer's settings), perform the update by calling through the shared
        // {@link PlatformContext#updateSettings} so that the update is seen by all settings observers.
        //
        // If the settings update is for some other subject that is unrelated to the viewer, then this is not
        // necessary.
        const isSubjectInViewerSettingsCascade = this.props.settingsCascade.subjects?.some(
            ({ subject }) => subject.id === this.props.subject.id
        )

        try {
            if (isSubjectInViewerSettingsCascade) {
                await this.props.platformContext.updateSettings(this.props.subject.id, contents)
            } else {
                await overwriteSettings(this.props.platformContext, this.props.subject.id, lastID, contents)
            }

            this.setState({ commitError: undefined })
            this.props.onUpdate()
        } catch (commitError) {
            this.setState({ commitError })
            logger.error(commitError)
        }
    }

    private onDidDiscard = (): void => {
        this.setState({ commitError: undefined })
    }
}
