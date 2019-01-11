import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { Select } from '../../../components/Select'
import {
    EXTENSION_NAME_MAX_LENGTH,
    EXTENSION_NAME_VALID_PATTERN,
    publisherName,
    RegistryPublisher,
} from '../../../extensions/extension/extension'

export const RegistryPublisherFormGroup: React.FunctionComponent<{
    className?: string

    /** The current publisher value. */
    value?: GQL.ID

    /** The viewer's authorized publishers, undefined while loading, or an error. */
    publishersOrError: 'loading' | RegistryPublisher[] | ErrorLike

    disabled?: boolean
    onChange?: React.FormEventHandler<HTMLSelectElement>
}> = ({ className = '', value, publishersOrError, disabled, onChange }) => (
    <div className={`form-group ${className}`}>
        <label htmlFor="extension-registry-create-extension-page__publisher">Publisher</label>
        {isErrorLike(publishersOrError) ? (
            <div className="alert alert-danger">{publishersOrError.message}</div>
        ) : (
            <Select
                id="extension-registry-create-extension-page__publisher"
                onChange={onChange}
                required={true}
                disabled={disabled || publishersOrError === 'loading'}
                value={value}
            >
                {publishersOrError === 'loading' ? (
                    <option disabled={true}>Loading...</option>
                ) : (
                    publishersOrError.map((p, i) => (
                        <option key={p.id} value={p.id}>
                            {publisherName(p)}
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
    <div className={`form-group ${className}`}>
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
