import React from 'react'

import { Button, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { SaveToolbar, SaveToolbarProps } from '../../../../components/SaveToolbar'

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
            inferEnabled && (
                <Tooltip content="Infer index configuration from HEAD">
                    <Button type="button" variant="secondary" outline={true} className="ml-2" onClick={onInfer}>
                        Infer configuration
                    </Button>
                </Tooltip>
            )
        )}
    </SaveToolbar>
)
