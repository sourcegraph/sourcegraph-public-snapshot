import { FC } from 'react'

import { Button, Input, getDefaultInputProps, useControlledField } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import { Repository } from '../LocalRepositoriesStep'

interface LocalRepositoryFormProps {
    onCancel: () => void
    repositoryInEdit?: Repository
}

export const LocalRepositoryForm: FC<LocalRepositoryFormProps> = ({ onCancel, repositoryInEdit }) => {
    const repositoryPath = useControlledField({
        value: repositoryInEdit?.url || '',
        name: 'repositoryPath',
        // submitted: form.formAPI.submitted,
        // formTouched: form.formAPI.touched,
        // validators: {
        //     sync: isTabActive ? syncAccessTokenValidator : undefined,
        //     async: isTabActive ? accessTokenAsyncValidator : undefined,
        // },
        // onChange: value => configurationField.input.onChange(modify(configurationField.input.value, ['token'], value)),
    })
    // TODO: On Submit > Implement local repo discovery (getDiscoveredLocalRepos()) --> https://github.com/sourcegraph/sourcegraph/issues/48128

    // TODO: How to give BE file picker option for edit mode? When !!repositoryInEdit

    return (
        <div className="d-flex w-100">
            <Input
                label="Project path"
                {...getDefaultInputProps(repositoryPath)}
                placeholder="user/path/repo-1"
                className="mb-0 col-5"
            />

            <div className="d-flex align-items-end mb-1 col-5">
                <LoaderButton
                    type="submit"
                    variant="primary"
                    size="sm"
                    label="Connect"
                    alwaysShowLabel={true}
                    loading={false}
                    disabled={false}
                />

                <Button size="sm" onClick={onCancel} variant="secondary" className="ml-2">
                    Cancel
                </Button>
            </div>
        </div>
    )
}
