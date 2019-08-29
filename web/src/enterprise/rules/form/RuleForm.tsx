import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { RuleDefinitionFormControl } from './definition/RuleDefinitionFormControl'
import { RuleActionsFormControl } from './definition/actions/RuleActionsFormControl'

export interface RuleFormData extends Pick<GQL.IRule, 'name' | 'description'> {
    definition: string
}

interface Props extends ExtensionsControllerProps {
    header: React.ReactFragment
    initialValue?: RuleFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss: () => void

    /** Called when the form is submitted. */
    onSubmit: (rule: RuleFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
    history: H.History
}

/**
 * A form to create or edit a rule.
 */
export const RuleForm: React.FunctionComponent<Props> = ({
    header,
    initialValue = { name: '', description: null, definition: '' },
    onDismiss,
    onSubmit: onSubmitRule,
    buttonText,
    isLoading,
    className = '',
    history,
    extensionsController,
}) => {
    const [name, setName] = useState(initialValue.name)
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setName(e.currentTarget.value),
        []
    )
    useEffect(() => setName(initialValue.name), [initialValue.name])

    const [description, setDescription] = useState(initialValue.description || null)
    const onDescriptionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setDescription(e.currentTarget.value),
        []
    )
    useEffect(() => setDescription(initialValue.description || null), [initialValue.description])
    const onAddDescriptionClick = useCallback(() => setDescription(''), [])

    const [definition, setDefinition] = useState(initialValue.definition)
    useEffect(() => setDefinition(initialValue.definition), [initialValue.definition])

    // TODO!(sqs): hackDidSave works around a problem where the history.block prompt shows up after
    // creating/editing successfully.
    const [hackDidSave, setHackDidSave] = useState(false)

    const onSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            setHackDidSave(true)
            onSubmitRule({ name, description, definition })
        },
        [onSubmitRule, name, definition, description]
    )

    // Warn when navigating away from page when that would result in loss of user input.
    useEffect(() => {
        const isDirty =
            name !== initialValue.name ||
            (description || '') !== (initialValue.description || '') ||
            definition !== initialValue.definition
        if (isDirty && !hackDidSave) {
            return history.block('Discard unsaved rule?')
        }
        return undefined
    }, [
        definition,
        description,
        hackDidSave,
        history,
        initialValue.definition,
        initialValue.description,
        initialValue.name,
        name,
    ])

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            {header}
            <div className="form-group">
                <label htmlFor="rule-form__name">Name</label>
                <input
                    type="text"
                    id="rule-form__name"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="Rule name"
                    value={name}
                    onChange={onNameChange}
                    autoFocus={true}
                />
            </div>
            {description !== null ? (
                <div className="form-group">
                    <label htmlFor="rule-form__description">Description</label>
                    <input
                        type="text"
                        id="rule-form__description"
                        className="form-control w-100"
                        placeholder="Optional description"
                        value={description || ''}
                        onChange={onDescriptionChange}
                    />
                </div>
            ) : (
                <button type="button" className="btn btn-sm btn-link px-0 pt-0 mb-2" onClick={onAddDescriptionClick}>
                    Add description
                </button>
            )}
            <RuleActionsFormControl extensionsController={extensionsController} />
            {false && (
                <RuleDefinitionFormControl
                    value={definition}
                    onChange={setDefinition}
                    extensionsController={extensionsController}
                />
            )}
            <div className="form-group mt-4 mb-0">
                <button type="reset" className="btn btn-secondary mr-2" onClick={onDismiss}>
                    Cancel
                </button>
                <button type="submit" disabled={isLoading} className="btn btn-success">
                    {isLoading ? <LoadingSpinner className="icon-inline" /> : buttonText}
                </button>
            </div>
        </Form>
    )
}
