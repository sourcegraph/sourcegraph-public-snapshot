import React from 'react'
import Form, { FormProps } from 'react-jsonschema-form'
import {
    DynamicallyImportedMonacoSettingsEditor,
    DynamicallyImportedMonacoSettingsEditorProps,
} from '../../settings/DynamicallyImportedMonacoSettingsEditor'

interface Props extends DynamicallyImportedMonacoSettingsEditorProps {
    form: FormProps<object>
}

/**
 * An editor for an instance (document) of a JSON Schema. It exposes both a UI and a raw JSON
 * editor.
 */
export const JSONSchemaInstanceEditor: React.FunctionComponent<Props> = props => (
    <div className="json-schema-instance-editor">
        <Form {...props.form} />
        <hr className="my-5" />
        {false && <DynamicallyImportedMonacoSettingsEditor {...props} />}
    </div>
)
