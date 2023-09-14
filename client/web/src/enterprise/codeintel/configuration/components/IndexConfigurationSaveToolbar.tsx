import React from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { SaveToolbar, type SaveToolbarProps } from '../../../../components/SaveToolbar'

import { ConfigurationInferButton } from './ConfigurationInferButton'

export interface IndexConfigurationSaveToolbarProps {
    loading: boolean
    inferEnabled: boolean
    onInfer?: () => void
}

export const IndexConfigurationSaveToolbar: React.FunctionComponent<
    SaveToolbarProps & IndexConfigurationSaveToolbarProps
> = ({ dirty, loading, saving, error, onSave, onDiscard, inferEnabled, onInfer, saveDiscardDisabled }) => (
    <SaveToolbar
        dirty={dirty}
        saving={saving}
        onSave={onSave}
        error={error}
        saveDiscardDisabled={saveDiscardDisabled}
        onDiscard={onDiscard}
    >
        {loading ? (
            <LoadingSpinner className="mt-2 ml-2" />
        ) : (
            inferEnabled && <ConfigurationInferButton onClick={onInfer} />
        )}
    </SaveToolbar>
)
