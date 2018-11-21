import { last } from 'lodash'
import * as React from 'react'
import { EMPTY, from, Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Toggle } from '../../../shared/src/components/Toggle'
import { ConfiguredExtension, isExtensionEnabled } from '../../../shared/src/extensions/extension'
import { SettingsCascade, SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { ExtensionsProps, isExtensionAdded, SettingsCascadeProps } from './ExtensionsClientCommonContext'

interface Props extends SettingsCascadeProps, ExtensionsProps {
    /** The extension that this element is for. */
    extension: ConfiguredExtension

    disabled?: boolean

    /** Class name applied to this element. */
    className?: string

    /** Class name applied to this element when it is an "Add" button. */
    addClassName?: string

    /** Called when the component performs an update that requires the parent component to refresh data. */
    onUpdate: () => void
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
                        if (isErrorLike(this.props.settingsCascade.subjects)) {
                            // TODO: Show error.
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
                            !confirmAddExtension(this.props.extension.id, this.props.extension.manifest)
                        ) {
                            return EMPTY
                        }

                        return from(
                            this.props.platformContext.updateSettings(highestPrecedenceSubject.subject.id, {
                                extensionID: this.props.extension.id,
                                enabled,
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
        const subject = isErrorLike(cascade)
            ? undefined
            : last(cascade.subjects.filter(subject => isExtensionAdded(subject.settings, this.props.extension.id)))
        const state = subject && {
            state: subject.settings.extensions ? subject.settings.extensions[this.props.extension.id] : false,
            name: subject.subject.__typename,
        }

        const onToggle = (enabled: boolean) => {
            this.toggles.next(enabled)
        }

        return (
            <Toggle
                value={isExtensionEnabled(this.props.settingsCascade.final, this.props.extension.id)}
                onToggle={onToggle}
                title={state ? `${state.state ? 'Enabled' : 'Disabled'} in ${state.name} settings` : 'Click to enable'}
            />
        )
    }
}

/**
 * Shows a modal confirmation prompt to the user confirming whether to add an extension.
 */
function confirmAddExtension(extensionID: string, extensionManifest?: ConfiguredExtension['manifest']): boolean {
    // Either `"title" (id)` (if there is a title in the manifest) or else just `id`. It is
    // important to show the ID because it indicates who the publisher is and allows
    // disambiguation from other similarly titled extensions.
    let displayName: string
    if (extensionManifest && !isErrorLike(extensionManifest) && extensionManifest.title) {
        displayName = `${JSON.stringify(extensionManifest.title)} (${extensionID})`
    } else {
        displayName = extensionID
    }
    return confirm(
        `Add Sourcegraph extension ${displayName}?\n\nIt can:\n- Read repositories and files you view using Sourcegraph\n- Read and change your Sourcegraph settings`
    )
}

/** Converts a SettingsCascadeOrError to a SettingsCascade, returning the first error it finds. */
function extractErrors(c: SettingsCascadeOrError): SettingsCascade | ErrorLike {
    if (c.subjects === null || isErrorLike(c.subjects)) {
        return new Error('Subjects was ' + c.subjects)
    } else if (c.final === null || isErrorLike(c.final)) {
        return new Error('Merged was ' + c.final)
    } else if (c.subjects.find(isErrorLike)) {
        return new Error('One of the subjects was ' + c.subjects.find(isErrorLike))
    } else {
        return c as SettingsCascade
    }
}
