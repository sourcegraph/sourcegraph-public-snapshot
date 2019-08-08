import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import * as GQL from '../../../../../../shared/src/graphql/schema'

interface Props {
    /**
     * The raw definition as JSONC.
     */
    value: GQL.IJSONC['raw']

    /**
     * Called when the value changes.
     */
    onChange: (value: GQL.IJSONC['raw']) => void
}

/**
 * A form control for specifying a rule's definition.
 */
export const RuleDefinitionFormControl: React.FunctionComponent<Props> = ({ value, onChange }) => {
    const onRawChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => onChange(e.currentTarget.value),
        [onChange]
    )
    return (
        <div className="form-group">
            <label htmlFor="rule-definition-form-control__raw">Definition</label>

            <TextareaAutosize
                id="rule-definition-form-control__raw"
                className="form-control"
                placeholder="Definition"
                required={true}
                minRows={4}
                value={value}
                onChange={onRawChange}
            />
        </div>
    )
}
