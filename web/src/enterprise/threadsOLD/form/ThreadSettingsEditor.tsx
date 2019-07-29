import * as H from 'history'
import * as React from 'react'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

interface Props {
    value: string
    onChange: (value: string) => void
    loading: boolean
    isLightTheme: boolean
    history: H.History
}

/**
 * Form for editing a thread's settings JSON.
 */
export const ThreadSettingsEditor: React.FunctionComponent<Props> = ({
    value,
    onChange,
    loading,
    isLightTheme,
    history,
}) => (
    <div className="form-group">
        <DynamicallyImportedMonacoSettingsEditor
            value={value}
            jsonSchema={undefined} // TODO!(sqs)
            canEdit={false}
            loading={loading}
            saving={loading}
            height={350}
            isLightTheme={isLightTheme}
            onChange={onChange}
            history={history}
        />
        <p className="form-text text-muted">
            <small>Use Ctrl+Space for completion, and hover over JSON properties for documentation.</small>
        </p>
    </div>
)
