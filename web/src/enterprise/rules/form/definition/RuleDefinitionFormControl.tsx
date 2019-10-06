import { applyEdits, format } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { parseJSON } from '../../../../settings/configuration'
import { defaultFormattingOptions } from '../../../../site-admin/configHelpers'
import { useLocalStorage } from '../../../../util/useLocalStorage'
import { parseDiagnosticQuery } from '../../../diagnostics/diagnosticQuery'
import { RuleDefinition } from '../../types'
import { useCommandRegistryCommands } from '../useCommandRegistryCommands'

interface Props extends ExtensionsControllerProps {
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
export const RuleDefinitionFormControl: React.FunctionComponent<Props> = ({
    value: raw,
    onChange,
    extensionsController,
}) => {
    const parsed: RuleDefinition = raw ? parseJSON(raw) : {}

    const onPropertyChange = useCallback(
        <P extends keyof RuleDefinition>(property: P, value: RuleDefinition[P]) => {
            onChange(applyEdits(raw, setProperty(raw, [property], value, defaultFormattingOptions)))
        },
        [onChange, raw]
    )

    const onQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => onPropertyChange('query', parseDiagnosticQuery(e.currentTarget.value)),
        [onPropertyChange]
    )

    const onActionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
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

    const registeredCommands = useCommandRegistryCommands(extensionsController)

    return (
        <>
            <div className="form-group">
                <label htmlFor="rule-definition-form-control__query">Query</label>
                <input
                    type="text"
                    id="rule-definition-form-control__query"
                    className="form-control"
                    placeholder="Search query"
                    value={parsed.query ? parsed.query.input : ''}
                    onChange={onQueryChange}
                />
            </div>
            <div className="form-group">
                <label htmlFor="rule-definition-form-control__action">Action</label>
                <input
                    id="rule-definition-form-control__action"
                    type="text"
                    className="form-control"
                    onChange={onActionChange}
                    value={parsed.action || ''}
                    list="rule-definition-form-control__action-datalist"
                />
                <datalist id="rule-definition-form-control__action-datalist">
                    {registeredCommands.map(({ command }) => (
                        <option key={command} value={command} />
                    ))}
                </datalist>
            </div>
            {isRawVisible ? (
                <div className="form-group">
                    <label htmlFor="rule-definition-form-control__raw">Raw JSON</label>
                    <TextareaAutosize
                        id="rule-definition-form-control__raw"
                        className="form-control text-monospace small"
                        required={true}
                        minRows={4}
                        value={raw ? applyEdits(raw, format(raw, undefined as any, defaultFormattingOptions)) : '{}'}
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
