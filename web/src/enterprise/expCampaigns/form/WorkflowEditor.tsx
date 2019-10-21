import React, { useCallback, useMemo } from 'react'
import workflowSchemaJSON from '../../../../../shared/src/schema/workflow.schema.json'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'
import H from 'history'
import { ThemeProps } from '../../../theme'
import { CampaignFormControl } from './CampaignForm'
import { JSONSchema7 } from 'json-schema'
import { isDefined } from '../../../../../shared/src/util/types'

interface Props extends CampaignFormControl, ThemeProps {
    history: H.History
}

export const WorkflowEditor: React.FunctionComponent<Props> = ({
    value,
    workflowJSONSchema,
    onChange: parentOnChange,
    ...props
}) => {
    const onChange = useCallback(
        (value: string): void => {
            parentOnChange({ workflowAsJSONCString: value })
        },
        [parentOnChange]
    )
    const schema = useMemo<JSONSchema7>(
        () => ({ allOf: [workflowSchemaJSON as JSONSchema7, workflowJSONSchema].filter(isDefined) }),
        []
    )
    return (
        <DynamicallyImportedMonacoSettingsEditor
            {...props}
            value={value.workflowAsJSONCString}
            jsonSchema={schema}
            canEdit={false}
            height={300}
            onChange={onChange}
        />
    )
}
