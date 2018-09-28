import * as React from 'react'
import { ExtensionsProps } from '../context'
import { isErrorLike } from '../errors'
import { ConfigurationCascadeProps, ConfigurationSubject, Settings } from '../settings'
import { ConfiguredExtension, confirmAddExtension, isExtensionAdded } from './extension'
import { ExtensionAddButton } from './ExtensionAddButton'
import { ExtensionConfigureButtonDropdown } from './ExtensionConfigureButtonDropdown'

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
 * Displays the primary action for an extension.
 *
 * - "Add" if the extension is not yet added and can be added.
 * - "Configure (icon)" dropdown menu in all other cases.
 */
export class ExtensionPrimaryActionButton<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<Props<S, C>> {
    public render(): JSX.Element | null {
        if (this.props.configurationCascade.subjects === null) {
            return null
        }
        if (isErrorLike(this.props.configurationCascade.subjects)) {
            // TODO: Show error.
            return null
        }

        // Only operate on the highest precedence settings, for simplicity.
        const subjects = this.props.configurationCascade.subjects
        if (subjects.length === 0) {
            return null
        }
        const highestPrecedenceSubject = subjects[subjects.length - 1]
        if (!highestPrecedenceSubject || !highestPrecedenceSubject.subject.viewerCanAdminister) {
            return null
        }

        if (
            this.props.configurationCascade.subjects.every(s => !isExtensionAdded(s.settings, this.props.extension.id))
        ) {
            return (
                <ExtensionAddButton
                    extension={this.props.extension}
                    subject={highestPrecedenceSubject}
                    confirm={this.confirm}
                    onUpdate={this.props.onUpdate}
                    className={`btn ${this.props.className || ''} ${this.props.addClassName || ''}`}
                    extensions={this.props.extensions}
                >
                    <this.props.extensions.context.icons.Add className="icon-inline" /> Add
                </ExtensionAddButton>
            )
        }
        return (
            <div className="btn-group" role="group" aria-label="Extension configuration actions">
                <ExtensionConfigureButtonDropdown
                    extension={this.props.extension}
                    subject={highestPrecedenceSubject}
                    configurationCascade={this.props.configurationCascade}
                    confirm={this.confirm}
                    onUpdate={this.props.onUpdate}
                    buttonClassName={`btn-outline-link ${this.props.className || ''}`}
                    extensions={this.props.extensions}
                >
                    <this.props.extensions.context.icons.Settings className="icon-inline" />
                </ExtensionConfigureButtonDropdown>
            </div>
        )
    }

    public confirm = () => confirmAddExtension(this.props.extension.id, this.props.extension.manifest)
}
