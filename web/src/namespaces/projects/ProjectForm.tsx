import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Form } from '../../components/Form'

export interface ProjectFormData extends Pick<GQL.IProject, 'name'> {}

interface Props {
    initialValue?: ProjectFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss: () => void

    /** Called when the form is submitted. */
    onSubmit: (project: ProjectFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
}

/**
 * A form to create or edit a project.
 */
export const ProjectForm: React.FunctionComponent<Props> = ({
    initialValue = { name: '' },
    onDismiss,
    onSubmit: onSubmitProject,
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

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitProject({ name })
        },
        [name, onSubmitProject]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="form-row align-items-end">
                <div className="form-group mb-md-0 col-md-3">
                    <label htmlFor="new-project-form__name">Name</label>
                    <input
                        type="text"
                        id="new-project-form__name"
                        className="form-control"
                        required={true}
                        minLength={1}
                        placeholder="Project name"
                        value={name}
                        onChange={onNameChange}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group mb-md-0 col-md-3 text-right">
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
