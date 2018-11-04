import { ConfiguredExtension, isExtensionEnabled } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascade,
    ConfigurationCascadeOrError,
    ConfigurationSubject,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { Toggle } from '@sourcegraph/extensions-client-common/lib/ui/generic/Toggle'
import { last } from 'lodash'
import * as React from 'react'
import { EMPTY, from, Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Settings } from '../schema/settings.schema'
import { ErrorLike, isErrorLike } from '../util/errors'
import { ConfigurationCascadeProps, ExtensionsProps, isExtensionAdded } from './ExtensionsClientCommonContext'

interface Props<S extends ConfigurationSubject, C extends Settings> extends ConfigurationCascadeProps, ExtensionsProps {
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
export class ExtensionToggle<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>
> {
    private toggles = new Subject<boolean>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.toggles
                .pipe(
                    switchMap(enabled => {
                        if (this.props.configurationCascade.subjects === null) {
                            return EMPTY
                        }
                        if (isErrorLike(this.props.configurationCascade.subjects)) {
                            // TODO: Show error.
                            return EMPTY
                        }

                        // Only operate on the highest precedence settings, for simplicity.
                        const subjects = this.props.configurationCascade.subjects
                        if (subjects.length === 0) {
                            return EMPTY
                        }
                        const highestPrecedenceSubject = subjects[subjects.length - 1]
                        if (!highestPrecedenceSubject || !highestPrecedenceSubject.subject.viewerCanAdminister) {
                            return EMPTY
                        }

                        if (
                            !isExtensionAdded(this.props.configurationCascade.merged, this.props.extension.id) &&
                            !confirmAddExtension(this.props.extension.id, this.props.extension.manifest)
                        ) {
                            return EMPTY
                        }

                        return from(
                            this.props.extensions.context.updateExtensionSettings(highestPrecedenceSubject.subject.id, {
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
        const cascade = extractErrors(this.props.configurationCascade)
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
                value={isExtensionEnabled(this.props.configurationCascade.merged, this.props.extension.id)}
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

/** Converts a ConfigurationCascadeOrError to a ConfigurationCascade, returning the first error it finds. */
function extractErrors(
    c: ConfigurationCascadeOrError<ConfigurationSubject, Settings>
): ConfigurationCascade | ErrorLike {
    if (c.subjects === null || isErrorLike(c.subjects)) {
        return new Error('Subjects was ' + c.subjects)
    } else if (c.merged === null || isErrorLike(c.merged)) {
        return new Error('Merged was ' + c.merged)
    } else if (c.subjects.find(isErrorLike)) {
        return new Error('One of the subjects was ' + c.subjects.find(isErrorLike))
    } else {
        return c as ConfigurationCascade
    }
}
