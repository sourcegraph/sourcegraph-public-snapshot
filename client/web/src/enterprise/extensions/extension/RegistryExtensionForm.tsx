import * as React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Input, Select } from '@sourcegraph/wildcard'

import {
    EXTENSION_NAME_MAX_LENGTH,
    EXTENSION_NAME_VALID_PATTERN,
    publisherName,
    RegistryPublisher,
} from '../../../extensions/extension/extension'

export const RegistryPublisherFormGroup: React.FunctionComponent<{
    /** The current publisher value. */
    value?: Scalars['ID']

    /** The viewer's authorized publishers, undefined while loading, or an error. */
    publishersOrError: 'loading' | RegistryPublisher[] | ErrorLike

    disabled?: boolean
    onChange?: React.FormEventHandler<HTMLSelectElement>

    className?: string
}> = ({ value, publishersOrError, disabled, onChange, className }) => (
    <>
        {isErrorLike(publishersOrError) ? (
            <ErrorAlert error={publishersOrError} />
        ) : (
            <Select
                label="Publisher"
                onChange={onChange}
                required={true}
                disabled={disabled || publishersOrError === 'loading'}
                value={value}
                aria-label="Publisher"
                message="The owner of this extension. This can't be changed after creation."
                className={className}
            >
                {publishersOrError === 'loading' ? (
                    <option disabled={true}>Loading...</option>
                ) : (
                    publishersOrError.map(publisher => (
                        <option key={publisher.id} value={publisher.id}>
                            {publisherName(publisher)}
                        </option>
                    ))
                )}
            </Select>
        )}
    </>
)

export const RegistryExtensionNameFormGroup: React.FunctionComponent<{
    value: string
    disabled?: boolean
    onChange: React.FormEventHandler<HTMLInputElement>
    className?: string
}> = ({ value, disabled, onChange, className }) => (
    <Input
        label="Name"
        type="text"
        name="extension-name"
        onChange={onChange}
        required={true}
        autoFocus={true}
        spellCheck={false}
        autoCapitalize="off"
        autoCorrect="off"
        autoComplete="off"
        value={value}
        pattern={EXTENSION_NAME_VALID_PATTERN}
        maxLength={EXTENSION_NAME_MAX_LENGTH}
        disabled={disabled}
        message="The name for this extension."
        className={className}
    />
)
