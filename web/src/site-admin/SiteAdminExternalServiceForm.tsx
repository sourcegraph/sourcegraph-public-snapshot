import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { DynamicallyImportedMonacoSettingsEditor } from '../settings/DynamicallyImportedMonacoSettingsEditor'
import { AddExternalServiceOptions } from './externalServices'
import { ErrorAlert, ErrorMessage } from '../components/alerts'

interface Props extends Pick<AddExternalServiceOptions, 'jsonSchema' | 'editorActions'> {
    history: H.History
    input: GQL.IAddExternalServiceInput
    isLightTheme: boolean
    error?: ErrorLike
    warning?: string
    mode: 'edit' | 'create'
    loading: boolean
    hideDisplayNameField?: boolean
    submitName?: string
    onSubmit: (event?: React.FormEvent<HTMLFormElement>) => void
    onChange: (change: GQL.IAddExternalServiceInput) => void
}

/**
 * Form for submitting a new or updated external service.
 */
export class SiteAdminExternalServiceForm extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <Form className="external-service-form" onSubmit={that.props.onSubmit}>
                {that.props.error && <ErrorAlert error={that.props.error} />}
                {that.props.warning && (
                    <div className="alert alert-warning">
                        <h4>Warning</h4>
                        <ErrorMessage error={that.props.warning} />
                    </div>
                )}
                {that.props.hideDisplayNameField || (
                    <div className="form-group">
                        <label className="font-weight-bold" htmlFor="e2e-external-service-form-display-name">
                            Display name:
                        </label>
                        <input
                            id="e2e-external-service-form-display-name"
                            type="text"
                            className="form-control"
                            required={true}
                            autoCorrect="off"
                            autoComplete="off"
                            autoFocus={true}
                            spellCheck={false}
                            value={that.props.input.displayName}
                            onChange={that.onDisplayNameChange}
                            disabled={that.props.loading}
                        />
                    </div>
                )}

                <div className="form-group">
                    <DynamicallyImportedMonacoSettingsEditor
                        // DynamicallyImportedMonacoSettingsEditor does not re-render the passed input.config
                        // if it thinks the config is dirty. We want to always replace the config if the kind changes
                        // so the editor is keyed on the kind.
                        value={that.props.input.config}
                        jsonSchema={that.props.jsonSchema}
                        canEdit={false}
                        loading={that.props.loading}
                        height={350}
                        isLightTheme={that.props.isLightTheme}
                        onChange={that.onConfigChange}
                        history={that.props.history}
                        actions={that.props.editorActions}
                    />
                    <p className="form-text text-muted">
                        <small>Use Ctrl+Space for completion, and hover over JSON properties for documentation.</small>
                    </p>
                </div>
                <button
                    type="submit"
                    className={`btn btn-primary ${
                        that.props.mode === 'create'
                            ? 'e2e-add-external-service-button'
                            : 'e2e-update-external-service-button'
                    }`}
                    disabled={that.props.loading}
                >
                    {that.props.loading && <LoadingSpinner className="icon-inline" />}
                    {that.props.submitName ?? (that.props.mode === 'edit' ? 'Update repositories' : 'Add repositories')}
                </button>
            </Form>
        )
    }

    private onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        that.props.onChange({ ...that.props.input, displayName: event.currentTarget.value })
    }

    private onConfigChange = (config: string): void => {
        that.props.onChange({ ...that.props.input, config })
    }
}
