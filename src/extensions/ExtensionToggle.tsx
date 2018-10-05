import { last } from 'lodash-es'
import * as React from 'react'
import { EMPTY, from, Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ExtensionsProps } from '../context'
import { isErrorLike } from '../errors'
import { ConfigurationCascadeProps, ConfigurationSubject, extractErrors, Settings } from '../settings'
import { Toggle } from '../ui/generic/Toggle'
import { ConfiguredExtension, confirmAddExtension, isExtensionAdded, isExtensionEnabled } from './extension'

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ConfigurationCascadeProps<S, C>,
        ExtensionsProps<S, C> {
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
