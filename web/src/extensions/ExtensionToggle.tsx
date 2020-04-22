import { last } from 'lodash'
import * as React from 'react'
import { EMPTY, from, Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Toggle } from '../../../shared/src/components/Toggle'
import { ConfiguredRegistryExtension, isExtensionEnabled } from '../../../shared/src/extensions/extension'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascade, SettingsCascadeOrError, SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { eventLogger } from '../tracking/eventLogger'
import { isExtensionAdded } from './extension/extension'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    /** The extension that this element is for. */
    extension: Pick<ConfiguredRegistryExtension, 'id'>

    className?: string
}

/**
 * Displays a toggle button for an extension.
 */
export class ExtensionToggle extends React.PureComponent<Props> {
    private toggles = new Subject<boolean>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.toggles
                .pipe(
                    switchMap(enabled => {
                        if (this.props.settingsCascade.subjects === null) {
                            return EMPTY
                        }

                        // Only operate on the highest precedence settings, for simplicity.
                        const subjects = this.props.settingsCascade.subjects
                        if (subjects.length === 0) {
                            return EMPTY
                        }
                        const highestPrecedenceSubject = subjects[subjects.length - 1]
                        if (!highestPrecedenceSubject || !highestPrecedenceSubject.subject.viewerCanAdminister) {
                            return EMPTY
                        }

                        if (
                            !isExtensionAdded(this.props.settingsCascade.final, this.props.extension.id) &&
                            !confirmAddExtension(this.props.extension.id)
                        ) {
                            return EMPTY
                        }

                        eventLogger.log('ExtensionToggled', { extension_id: this.props.extension.id })

                        return from(
                            this.props.platformContext.updateSettings(highestPrecedenceSubject.subject.id, {
                                path: ['extensions', this.props.extension.id],
                                value: enabled,
                            })
                        )
                    })
                )
                .subscribe()
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const cascade = extractErrors(this.props.settingsCascade)
        const highestPrecedenceSubjectWithExtensionAdded = isErrorLike(cascade)
            ? undefined
            : last(cascade.subjects.filter(subject => isExtensionAdded(subject.settings, this.props.extension.id)))

        let title: string
        if (highestPrecedenceSubjectWithExtensionAdded) {
            // Describe highest-precedence subject where this extension is enabled.
            title = `${
                isExtensionEnabled(highestPrecedenceSubjectWithExtensionAdded.settings, this.props.extension.id)
                    ? 'Enabled'
                    : 'Disabled'
            } in ${highestPrecedenceSubjectWithExtensionAdded.subject.__typename.toLowerCase()} settings`
        } else {
            title = 'Click to enable'
        }

        return (
            <Toggle
                value={isExtensionEnabled(this.props.settingsCascade.final, this.props.extension.id)}
                onToggle={this.onToggle}
                title={title}
                className={this.props.className}
            />
        )
    }

    private onToggle = (enabled: boolean): void => {
        this.toggles.next(enabled)
    }
}

/**
 * Shows a modal confirmation prompt to the user confirming whether to add an extension.
 */
function confirmAddExtension(extensionID: string): boolean {
    return confirm(
        `Add Sourcegraph extension ${extensionID}?\n\nIt can:\n- Read repositories and files you view using Sourcegraph\n- Read and change your Sourcegraph settings`
    )
}

/** Converts a SettingsCascadeOrError to a SettingsCascade, returning the first error it finds. */
function extractErrors(c: SettingsCascadeOrError): SettingsCascade | ErrorLike {
    if (c.subjects === null) {
        return new Error('Subjects was ' + c.subjects)
    }
    if (c.final === null || isErrorLike(c.final)) {
        return new Error('Merged was ' + c.final)
    }
    if (c.subjects.find(isErrorLike)) {
        return new Error('One of the subjects was ' + c.subjects.find(isErrorLike))
    }
    return c as SettingsCascade
}
