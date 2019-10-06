import React, { useCallback } from 'react'
import workflowSchemaJSON from '../../../../../shared/src/schema/workflow.schema.json'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'
import H from 'history'
import { ThemeProps } from '../../../theme'
import { CampaignFormControl } from './CampaignForm'

interface Props extends CampaignFormControl, ThemeProps {
    history: H.History
}

export const WorkflowEditor: React.FunctionComponent<Props> = ({ value, onChange: parentOnChange, ...props }) => {
    const onChange = useCallback(
        (value: string): void => {
            parentOnChange({ workflowAsJSONCString: value })
        },
        [parentOnChange]
    )
    return (
        <DynamicallyImportedMonacoSettingsEditor
            {...props}
            value={value.workflowAsJSONCString}
            jsonSchema={workflowSchemaJSON}
            canEdit={false}
            height={300}
            onChange={onChange}
        />
    )
}
