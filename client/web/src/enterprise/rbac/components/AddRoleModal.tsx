import React from 'react'

import { Button, Modal, Input, H3, Text, Label, ErrorAlert, Form, useCheckboxes, useForm, useField, SubmissionResult } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { PermissionsList } from './Permissions'
import { useCreateRole, PermissionsMap } from '../backend'
import { PermissionFields } from '../../../graphql-operations'

export interface AddRoleModalProps {
    onCancel: () => void
    afterCreate: () => void
    allPermissions: PermissionsMap
}

interface AddRoleModalFormValues {
    name: string
    permissions: string[]
}

const DEFAULT_FORM_VALUES: AddRoleModalFormValues = {
    name: '',
    permissions: []
}

export const AddRoleModal: React.FunctionComponent<React.PropsWithChildren<AddRoleModalProps>> = ({
    onCancel,
    afterCreate,
    allPermissions
}) => {
    const labelId = 'addRole'

    const [createRole, { loading, error }] = useCreateRole(afterCreate)
    const onSubmit = (values: AddRoleModalFormValues): SubmissionResult => {
        const { name, permissions } = values
        createRole({ variables: { name, permissions } })
    }

    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: DEFAULT_FORM_VALUES,
        onSubmit,
    })

    const { input: { isChecked, onBlur, onChange } } = useCheckboxes('permissions', formAPI)
    const nameInput = useField({
        name: 'name',
        formApi: formAPI,
    })

    const isButtonDisabled = nameInput.input.value.trimStart().length === 0 || formAPI.submitting
    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Add new role</H3>
            <Text>Enter a descriptive name for the role to be created.</Text>

            <Form onSubmit={handleSubmit} ref={ref}>
                <Label className="w-100">
                    <Text alignment="left" className="mb-2">Name</Text>

                    <Input
                        id="name"
                        autoComplete="off"
                        autoCapitalize="off"
                        autoFocus={true}
                        inputClassName="mb-2"
                        className="mb-0 form-group"
                        required={true}
                        spellCheck="false"
                        minLength={1}
                        // pattern="^[A-Za-z0-9_-]*$"
                        placeholder="The name of the role to be created."
                        {...nameInput.input}
                    />
                </Label>

                <PermissionsList allPermissions={allPermissions} isChecked={isChecked} onBlur={onBlur} onChange={onChange} />
                {error && !loading && <ErrorAlert error={error} className="my-2" />}

                <div className="d-flex justify-content-end">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        disabled={isButtonDisabled}
                        loading={formAPI.submitting}
                        variant="primary"
                        alwaysShowLabel={true}
                        label="Add role"
                    />
                </div>
            </Form>
        </Modal>
    )
}
