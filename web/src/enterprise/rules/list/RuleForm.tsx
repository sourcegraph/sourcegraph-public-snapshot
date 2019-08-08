import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'

export interface RuleFormData extends Pick<GQL.IRule, 'name' | 'description'> {
    definition: string
}

interface Props {
    initialValue?: RuleFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss: () => void

    /** Called when the form is submitted. */
    onSubmit: (rule: RuleFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
}

/**
 * A form to create or edit a rule.
 */
export const RuleForm: React.FunctionComponent<Props> = ({
    initialValue = { name: '', description: '', definition: '' },
    onDismiss,
    onSubmit: onSubmitRule,
    buttonText,
    isLoading,
    className = '',
}) => {
    const [name, setName] = useState(initialValue.name)
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setName(e.currentTarget.value),
        []
    )
    useEffect(() => setName(initialValue.name), [initialValue.name])

    const [description, setDescription] = useState(initialValue.description)
    const onDescriptionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setDescription(e.currentTarget.value),
        []
    )
    useEffect(() => setDescription(initialValue.description), [initialValue.description])

    const [definition, setDefinition] = useState(initialValue.definition)
    const onDefinitionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setDefinition(e.currentTarget.value),
        []
    )
    useEffect(() => setDefinition(initialValue.definition), [initialValue.definition])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitRule({ name, description, definition })
        },
        [onSubmitRule, name, definition, description]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="d-flex align-items-end">
                <div className="form-group mb-0 mr-3">
                    <label htmlFor="rule-form__name">Name</label>
                    <input
                        type="text"
                        id="rule-form__name"
                        className="form-control"
                        required={true}
                        minLength={1}
                        size={16}
                        placeholder="Rule name"
                        value={name}
                        onChange={onNameChange}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group mb-0 flex-1 mr-3">
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
                <div className="form-group mb-0 mr-3">
                    <label htmlFor="rule-form__definition">Definition</label>
                    <input
                        type="text"
                        id="rule-form__definition"
                        className="form-control"
                        placeholder="Definition"
                        required={true}
                        size={7}
                        maxLength={7}
                        value={definition}
                        onChange={onDefinitionChange}
                    />
                </div>
                <div className="form-group mb-0">
                    <button type="reset" className="btn btn-secondary mr-2" onClick={onDismiss}>
                        Cancel
                    </button>
                    <button type="submit" disabled={isLoading} className="btn btn-success">
                        {isLoading ? <LoadingSpinner className="icon-inline" /> : buttonText}
                    </button>
                </div>
            </div>
        </Form>
    )
}
