import { ConfiguredExtension, isExtensionAdded } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ADDED_AND_CAN_ADMINISTER,
    ALL_CAN_ADMINISTER,
    ExtensionConfigureButton,
    ExtensionConfiguredSubjectItemForAdd,
    ExtensionConfiguredSubjectItemForConfigure,
    ExtensionConfiguredSubjectItemForRemove,
} from '@sourcegraph/extensions-client-common/lib/extensions/ExtensionConfigureButton'
import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { isErrorLike } from '../../util/errors'
import { ConfigurationCascadeProps, ExtensionsProps } from '../ExtensionsClientCommonContext'

interface Props extends ConfigurationCascadeProps, ExtensionsProps {
    /** The extension that this element is for. */
    extension: ConfiguredExtension

    /** Class name applied to this element. */
    className?: string

    /** Class name applied to this element when it is a <button>. */
    buttonClassName?: string

    /** Class name applied to this element when it is NOT a <button>. */
    nonButtonClassName?: string

    /** Called when the component performs an update that requires the parent component to refresh data. */
    onUpdate: () => void
}

/**
 * Displays the primary action for a registry extension when it is the primary subject of the page (i.e., it is not
 * one list item of many).
 *
 * - "Add" if the extension is not yet added and can be added.
 * - "Added" if the extension is added and enabled.
 * - "Disabled" (no action) if the extension is added and disabled.
 */
export const RegistryExtensionDetailActionButton: React.SFC<Props> = props => {
    if (props.configurationCascade.subjects === null) {
        return null
    }
    if (isErrorLike(props.configurationCascade.subjects)) {
        // TODO: Show error.
        return null
    }
    if (props.configurationCascade.subjects.every(s => !isExtensionAdded(s.settings, props.extension.id))) {
        return (
            <ExtensionConfigureButton
                extension={props.extension}
                onUpdate={props.onUpdate}
                header="Add extension for..."
                itemFilter={ALL_CAN_ADMINISTER}
                itemComponent={ExtensionConfiguredSubjectItemForAdd}
                buttonClassName={`btn-primary ${props.className || ''} ${props.buttonClassName || ''}`}
                configurationCascade={props.configurationCascade}
                extensions={props.extensions}
            >
                <AddIcon className="icon-inline" /> Add extension
            </ExtensionConfigureButton>
        )
    }
    return (
        <div className="btn-group" role="group" aria-label="Extension configuration actions">
            <ExtensionConfigureButton
                extension={props.extension}
                onUpdate={props.onUpdate}
                header="Edit settings for..."
                itemFilter={ADDED_AND_CAN_ADMINISTER}
                itemComponent={ExtensionConfiguredSubjectItemForConfigure}
                buttonClassName={`btn-success ${props.className || ''} ${props.buttonClassName || ''}`}
                configurationCascade={props.configurationCascade}
                extensions={props.extensions}
            >
                Configure
            </ExtensionConfigureButton>
            <ExtensionConfigureButton
                extension={props.extension}
                onUpdate={props.onUpdate}
                header="Add extension for..."
                itemFilter={ALL_CAN_ADMINISTER}
                itemComponent={ExtensionConfiguredSubjectItemForAdd}
                buttonClassName={`btn-outline-link ${props.className || ''} ${props.buttonClassName || ''}`}
                configurationCascade={props.configurationCascade}
                extensions={props.extensions}
            >
                Add
            </ExtensionConfigureButton>
            <ExtensionConfigureButton
                extension={props.extension}
                onUpdate={props.onUpdate}
                header="Remove extension for..."
                itemFilter={ADDED_AND_CAN_ADMINISTER}
                itemComponent={ExtensionConfiguredSubjectItemForRemove}
                buttonClassName={`btn-outline-link ${props.className || ''} ${props.buttonClassName || ''}`}
                configurationCascade={props.configurationCascade}
                extensions={props.extensions}
            >
                Remove
            </ExtensionConfigureButton>
        </div>
    )
}
