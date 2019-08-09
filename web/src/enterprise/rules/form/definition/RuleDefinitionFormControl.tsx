import { applyEdits } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Select } from '../../../../components/Select'
import { parseJSON } from '../../../../settings/configuration'
import { defaultFormattingOptions } from '../../../../site-admin/configHelpers'
import { useLocalStorage } from '../../../../util/useLocalStorage'

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

export interface RuleDefinition {
    conditions?: string
    action?: string
}

/**
 * A form control for specifying a rule's definition.
 */
export const RuleDefinitionFormControl: React.FunctionComponent<Props> = ({ value: raw, onChange }) => {
    const parsed: RuleDefinition = raw ? parseJSON(raw) : {}

    const onPropertyChange = useCallback(
        (property: keyof RuleDefinition, value: string) => {
            onChange(applyEdits(raw, setProperty(raw, [property], value, defaultFormattingOptions)))
        },
        [onChange, raw]
    )

    const onConditionsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => onPropertyChange('conditions', e.currentTarget.value),
        [onPropertyChange]
    )

    const onActionChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        e => onPropertyChange('action', e.currentTarget.value),
        [onPropertyChange]
    )

    const [isRawVisible, setIsRawVisible] = useLocalStorage('RuleDefinitionFormControl.isRawVisible', false)
    const onShowRawClick = useCallback(() => setIsRawVisible(true), [setIsRawVisible])
    const onHideRawClick = useCallback(() => setIsRawVisible(false), [setIsRawVisible])
    const onRawChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => onChange(e.currentTarget.value),
        [onChange]
    )

    return (
        <>
            <div className="form-group">
                <label htmlFor="rule-definition-form-control__conditions">Watch for objects matching</label>
                <input
                    type="text"
                    id="rule-definition-form-control__conditions"
                    className="form-control"
                    placeholder="Search query"
                    value={parsed.conditions || ''}
                    onChange={onConditionsChange}
                />
            </div>
            <div className="form-group">
                <label htmlFor="rule-definition-form-control__subject">Action</label>
                <Select className="form-control w-auto" onChange={onActionChange} value={parsed.action}>
                    <option value={undefined} disabled={true}>
                        Select...
                    </option>
                    <option value="changeset-fix">Open or update changesets with action: Fix</option>
                    <option value="add-diagnostics">Add diagnostics to thread</option>
                </Select>
            </div>
            {isRawVisible ? (
                <div className="form-group">
                    <label htmlFor="rule-definition-form-control__raw">Raw JSON</label>
                    <TextareaAutosize
                        id="rule-definition-form-control__raw"
                        className="form-control text-monospace small"
                        required={true}
                        minRows={4}
                        value={raw || '{}'}
                        onChange={onRawChange}
                        readOnly={true}
                    />
                    <button type="button" className="btn btn-sm btn-link px-0 pt-0" onClick={onHideRawClick}>
                        Hide raw JSON
                    </button>
                </div>
            ) : (
                <button type="button" className="btn btn-sm btn-link px-0" onClick={onShowRawClick}>
                    Show raw JSON
                </button>
            )}
        </>
    )
}
