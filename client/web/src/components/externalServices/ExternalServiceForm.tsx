import React, { useCallback } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ErrorAlert, ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { ErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, LoadingSpinner, Alert, Typography } from '@sourcegraph/wildcard'

import { AddExternalServiceInput } from '../../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../../settings/DynamicallyImportedMonacoSettingsEditor'

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
    return (
        <Form className="external-service-form" onSubmit={onSubmit}>
            {error && <ErrorAlert error={error} />}
            {warning && (
                <Alert variant="warning">
                    <Typography.H4>Warning</Typography.H4>
                    <ErrorMessage error={warning} />
                </Alert>
            )}
            {hideDisplayNameField || (
                <div className="form-group">
                    <label className="font-weight-bold" htmlFor="test-external-service-form-display-name">
                        Display name:
                    </label>
                    <input
                        id="test-external-service-form-display-name"
                        type="text"
                        className="form-control"
                        required={true}
                        autoCorrect="off"
                        autoComplete="off"
                        autoFocus={autoFocus}
                        spellCheck={false}
                        value={input.displayName}
                        onChange={onDisplayNameChange}
                        disabled={loading}
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
                    isLightTheme={isLightTheme}
                    onChange={onConfigChange}
                    history={history}
                    actions={editorActions}
                    className="test-external-service-editor"
                    telemetryService={telemetryService}
                />
                <p className="form-text text-muted">
                    <small>Use Ctrl+Space for completion, and hover over JSON properties for documentation.</small>
                </p>
            </div>
            <Button
                type="submit"
                className={classNames(
                    'mb-3',
                    mode === 'create' ? 'test-add-external-service-button' : 'test-update-external-service-button'
                )}
                disabled={loading}
                variant="primary"
            >
                {loading && <LoadingSpinner />}
                {submitName ?? (mode === 'edit' ? 'Update repositories' : 'Add repositories')}
            </Button>
        </Form>
    )
}
