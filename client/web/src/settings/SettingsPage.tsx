import * as React from 'react'

import { RouteComponentProps } from 'react-router'

import { SettingsSubject } from '@sourcegraph/shared/src/schema'
import { overwriteSettings } from '@sourcegraph/shared/src/settings/edit'
import { SettingsCascade } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isSourcegraphAuthoredExtension } from '@sourcegraph/shared/src/util/extensions'
import { allowOnlySourcegraphAuthoredExtensionsFromSettings } from '@sourcegraph/shared/src/util/settings'

import { logCodeInsightsChanges } from '../insights/analytics'

import { SettingsAreaPageProps } from './SettingsArea'
import { SettingsFile } from './SettingsFile'

interface Props
    extends SettingsAreaPageProps,
        Pick<RouteComponentProps<{}>, 'history' | 'location'>,
        ThemeProps,
        TelemetryProps {
    /** Optional description to render above the editor. */
    description?: JSX.Element
}

interface State {
    commitError?: Error
}

const getSubjectByTypename = (
    settingsCascade: Props['settingsCascade'],
    typename: SettingsSubject['__typename']
): SettingsCascade['subjects'][number] | undefined =>
    (settingsCascade as SettingsCascade).subjects?.find(({ subject }) => subject.__typename === typename)

const reportExtensionsSettingsChange = (
    originalSettingsCascade: Props['settingsCascade'],
    updatedSettingsCascade: Props['settingsCascade'],
    currentSubject: Props['subject']
): void => {
    const {
        value: allowOnlySourcegraphAuthoredExtensions,
        subject: subjectDefinedSetting,
    } = allowOnlySourcegraphAuthoredExtensionsFromSettings(updatedSettingsCascade)

    const originalSubject = getSubjectByTypename(originalSettingsCascade, currentSubject.__typename)
    const updatedSubject = getSubjectByTypename(updatedSettingsCascade, currentSubject.__typename)
    const subjectDefinedAllowOnlySgExtensionSetting =
        subjectDefinedSetting && getSubjectByTypename(updatedSettingsCascade, subjectDefinedSetting)

    if (!originalSubject?.settings || !updatedSubject?.settings) {
        return
    }

    const currentSubjectHasLowerPriority =
        subjectDefinedAllowOnlySgExtensionSetting &&
        (updatedSettingsCascade as SettingsCascade).subjects.indexOf(subjectDefinedAllowOnlySgExtensionSetting) <
            (updatedSettingsCascade as SettingsCascade).subjects.indexOf(updatedSubject)

    const toggledNonSourcegraphExtensionsInstallation =
        originalSubject.settings['extensions.allowOnlySourcegraphAuthored'] !==
        updatedSubject.settings['extensions.allowOnlySourcegraphAuthored']

    const nonSourcegraphExtensionsEnabled = Object.entries(updatedSubject.settings.extensions || {}).some(
        ([extensionId, isEnabled]) =>
            !isSourcegraphAuthoredExtension(extensionId) &&
            isEnabled &&
            !originalSubject.settings!.extensions?.[extensionId]
    )

    if (currentSubjectHasLowerPriority) {
        if (
            allowOnlySourcegraphAuthoredExtensions &&
            ((toggledNonSourcegraphExtensionsInstallation &&
                !updatedSubject.settings['extensions.allowOnlySourcegraphAuthored']) ||
                nonSourcegraphExtensionsEnabled)
        ) {
            window.alert(
                `Installing non-SG extensions is disabled on ${subjectDefinedSetting} level. Any non-SG extension enabled in your settings will be ignored.`
            )
        }
    } else if (
        updatedSubject.settings['extensions.allowOnlySourcegraphAuthored'] &&
        (toggledNonSourcegraphExtensionsInstallation || nonSourcegraphExtensionsEnabled)
    ) {
        window.alert(
            'Installing non-SG extensions is disabled. Any non-SG extension enabled in your settings will be ignored.'
        )
    }
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
                telemetryService={this.props.telemetryService}
            />
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
            const originalSettings = this.props.settingsCascade

            if (isSubjectInViewerSettingsCascade) {
                await this.props.platformContext.updateSettings(this.props.subject.id, contents)
            } else {
                await overwriteSettings(this.props.platformContext, this.props.subject.id, lastID, contents)
            }

            const newSettings = this.props.settingsCascade
            if (originalSettings.final && newSettings.final) {
                reportExtensionsSettingsChange(originalSettings, newSettings, this.props.subject)

                logCodeInsightsChanges(originalSettings.final, newSettings.final, this.props.telemetryService)
            }

            this.setState({ commitError: undefined })
            this.props.onUpdate()
        } catch (commitError) {
            this.setState({ commitError })
            console.error(commitError)
        }
    }

    private onDidDiscard = (): void => {
        this.setState({ commitError: undefined })
    }
}
