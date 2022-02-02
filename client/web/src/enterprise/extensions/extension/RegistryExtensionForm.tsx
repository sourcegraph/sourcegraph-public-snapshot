import classNames from 'classnames'
import * as React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Select } from '@sourcegraph/wildcard'

import {
    EXTENSION_NAME_MAX_LENGTH,
    EXTENSION_NAME_VALID_PATTERN,
    publisherName,
    RegistryPublisher,
} from '../../../extensions/extension/extension'

export const RegistryPublisherFormGroup: React.FunctionComponent<{
    className?: string

    /** The current publisher value. */
    value?: Scalars['ID']

    /** The viewer's authorized publishers, undefined while loading, or an error. */
    publishersOrError: 'loading' | RegistryPublisher[] | ErrorLike

    disabled?: boolean
    onChange?: React.FormEventHandler<HTMLSelectElement>
}> = ({ className = '', value, publishersOrError, disabled, onChange }) => (
    <div className={className}>
        {isErrorLike(publishersOrError) ? (
            <ErrorAlert error={publishersOrError} />
        ) : (
            <Select
                selectClassName="form-control"
                label="Publisher"
                id="extension-registry-create-extension-page__publisher"
                onChange={onChange}
                required={true}
                disabled={disabled || publishersOrError === 'loading'}
                value={value}
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
        <small className="form-help text-muted">
            The owner of this extension. This can't be changed after creation.
        </small>
    </div>
)

export const RegistryExtensionNameFormGroup: React.FunctionComponent<{
    className?: string
    value: string
    disabled?: boolean
    onChange: React.FormEventHandler<HTMLInputElement>
}> = ({ className = '', value, disabled, onChange }) => (
    <div className={classNames('form-group', className)}>
        <label htmlFor="extension-registry-form__name">Name</label>
        <input
            type="text"
            name="extension-name"
            className="form-control"
            id="extension-registry-form__name"
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
        />
        <small className="form-help text-muted">The name for this extension.</small>
    </div>
)
