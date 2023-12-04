import React from 'react'

import { noop } from 'lodash'

import {
    Button,
    Modal,
    Input,
    H3,
    Text,
    Label,
    ErrorAlert,
    Form,
    useCheckboxes,
    useForm,
    useField,
    type SubmissionResult,
} from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { useCreateRole, type PermissionsMap } from '../backend'

import { PermissionsList } from './Permissions'

export interface CreateRoleModalProps {
    onCancel: () => void
    afterCreate: () => void
    allPermissions: PermissionsMap
}

interface CreateRoleModalFormValues {
    name: string
    permissions: string[]
}

const DEFAULT_FORM_VALUES: CreateRoleModalFormValues = {
    name: '',
    permissions: [],
}

export const CreateRoleModal: React.FunctionComponent<React.PropsWithChildren<CreateRoleModalProps>> = ({
    onCancel,
    afterCreate,
    allPermissions,
}) => {
    const labelId = 'createRole'

    const [createRole, { loading, error }] = useCreateRole(afterCreate)
    const onSubmit = (values: CreateRoleModalFormValues): SubmissionResult => {
        const { name, permissions } = values
        // We handle any error by destructuring the query result directly
        createRole({ variables: { name, permissions } }).catch(noop)
    }

    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: DEFAULT_FORM_VALUES,
        onSubmit,
    })

    const {
        input: { isChecked, onBlur, onChange },
    } = useCheckboxes('permissions', formAPI)
    const nameInput = useField({
        name: 'name',
        formApi: formAPI,
    })

    const isButtonDisabled = nameInput.input.value.trimStart().length === 0 || formAPI.submitting
    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Add new role</H3>
            <Text>Enter a unique, descriptive name for the role.</Text>

            <Form onSubmit={handleSubmit} ref={ref}>
                <Label className="w-100">
                    <Text alignment="left" className="mb-2">
                        Role name
                    </Text>

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
                        {...nameInput.input}
                    />
                </Label>

                <PermissionsList
                    allPermissions={allPermissions}
                    isChecked={isChecked}
                    onBlur={onBlur}
                    onChange={onChange}
                    roleName={nameInput.input.value}
                />
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
                        label="Create"
                    />
                </div>
            </Form>
        </Modal>
    )
}
