import React, { useCallback } from 'react'

import * as H from 'history'

import { ErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, LoadingSpinner, Alert, H4, Text, Input, ErrorAlert, ErrorMessage, Form } from '@sourcegraph/wildcard'

import { AddExternalServiceInput } from '../../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../../settings/DynamicallyImportedMonacoSettingsEditor'

import { ExternalServiceEditingDisabledAlert } from './ExternalServiceEditingDisabledAlert'
import { ExternalServiceEditingTemporaryAlert } from './ExternalServiceEditingTemporaryAlert'
import { AddExternalServiceOptions } from './externalServices'

interface Props extends Pick<AddExternalServiceOptions, 'jsonSchema' | 'editorActions'>, ThemeProps, TelemetryProps {
    history: H.History
    input: AddExternalServiceInput
    error?: ErrorLike
    warning?: string | null
    mode: 'edit' | 'create'
    loading: boolean
    hideDisplayNameField?: boolean
    submitName?: string
    onSubmit: (event?: React.FormEvent<HTMLFormElement>) => void
    onChange: (change: AddExternalServiceInput) => void
    autoFocus?: boolean
    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean
}

/**
 * Form for submitting a new or updated external service.
 */
export const ExternalServiceForm: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    isLightTheme,
    telemetryService,
    jsonSchema,
    editorActions,
    input,
    error,
    warning,
    mode,
    loading,
    hideDisplayNameField,
    submitName,
    onSubmit,
    onChange,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
    autoFocus = true,
}) => {
    const onDisplayNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            onChange({ ...input, displayName: event.currentTarget.value })
        },
        [input, onChange]
    )

    const onConfigChange = useCallback(
        (config: string): void => {
            onChange({ ...input, config })
        },
        [input, onChange]
    )
    const disabled = externalServicesFromFile && !allowEditExternalServicesWithFile

    return (
        <Form className="external-service-form" onSubmit={onSubmit}>
            {error && <ErrorAlert error={error} />}
            {warning && (
                <Alert variant="warning">
                    <H4>Warning</H4>
                    <ErrorMessage error={warning} />
                </Alert>
            )}

            {disabled && <ExternalServiceEditingDisabledAlert />}
            {externalServicesFromFile && allowEditExternalServicesWithFile && <ExternalServiceEditingTemporaryAlert />}

            {hideDisplayNameField || (
                <div className="form-group">
                    <Input
                        id="test-external-service-form-display-name"
                        required={true}
                        autoCorrect="off"
                        autoComplete="off"
                        autoFocus={autoFocus}
                        spellCheck={false}
                        value={input.displayName}
                        onChange={onDisplayNameChange}
                        disabled={loading || disabled}
                        label="Display name:"
                        className="mb-0"
                    />
                </div>
            )}
            <div className="form-group">
                <DynamicallyImportedMonacoSettingsEditor
                    // DynamicallyImportedMonacoSettingsEditor does not re-render the passed input.config
                    // if it thinks the config is dirty. We want to always replace the config if the kind changes
                    // so the editor is keyed on the kind.
                    value={input.config}
                    jsonSchema={jsonSchema}
                    canEdit={false}
                    loading={loading}
                    height={350}
                    readOnly={disabled}
                    isLightTheme={isLightTheme}
                    onChange={onConfigChange}
                    history={history}
                    actions={editorActions}
                    className="test-external-service-editor"
                    telemetryService={telemetryService}
                    explanation={
                        <Text className="form-text text-muted">
                            <small>
                                Use Ctrl+Space for completion, and hover over JSON properties for documentation.
                            </small>
                        </Text>
                    }
                />
            </div>
            <Button
                type="submit"
                className={
                    mode === 'create' ? 'test-add-external-service-button' : 'test-update-external-service-button'
                }
                disabled={loading || disabled}
                variant="primary"
            >
                {loading && <LoadingSpinner />}
                {submitName ?? (mode === 'edit' ? 'Update configuration' : 'Add repositories')}
            </Button>
        </Form>
    )
}
