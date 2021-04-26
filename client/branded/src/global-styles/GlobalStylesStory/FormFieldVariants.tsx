/* eslint-disable react/jsx-no-bind */
import classNames from 'classnames'
import React from 'react'
import 'storybook-addon-designs'

import styles from './FormFieldVariants.module.scss'

type FieldVariants = 'standard' | 'invalid' | 'valid' | 'disabled'

interface WithVariantsProps {
    field: React.ComponentType<{
        className?: string
        disabled?: boolean
        message?: JSX.Element
        variant: FieldVariants
    }>
}

const FieldMessage: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <small className={className}>Helper text</small>
)

const WithVariants: React.FunctionComponent<WithVariantsProps> = ({ field: Field }) => (
    <>
        <Field variant="standard" message={<FieldMessage className="field-message" />} />
        <Field variant="invalid" className="is-invalid" message={<FieldMessage className="invalid-feedback" />} />
        <Field variant="valid" className="is-valid" message={<FieldMessage className="valid-feedback" />} />
        <Field variant="disabled" disabled={true} message={<FieldMessage className="field-message" />} />
    </>
)

export const FormFieldVariants: React.FunctionComponent = () => (
    <div className={styles.grid}>
        <WithVariants
            field={({ className, message, ...props }) => (
                <fieldset className="form-group">
                    <input
                        type="text"
                        placeholder="Form field"
                        className={classNames('form-control', className)}
                        {...props}
                    />
                    {message}
                </fieldset>
            )}
        />
        <WithVariants
            field={({ className, message, ...props }) => (
                <fieldset className="form-group">
                    <select className={classNames('custom-select', className)} {...props}>
                        <option>Option A</option>
                        <option>Option B</option>
                        <option>Option C</option>
                    </select>
                    {message}
                </fieldset>
            )}
        />
        <WithVariants
            field={({ className, message, ...props }) => (
                <fieldset className="form-group">
                    <textarea
                        placeholder="This is sample content in a text area that spans four lines to see how it fits."
                        className={classNames('form-control', className)}
                        rows={4}
                        {...props}
                    />
                    {message}
                </fieldset>
            )}
        />
        <WithVariants
            field={({ className, message, variant, ...props }) => (
                <fieldset className="form-check">
                    <input
                        id={`inputFieldsetCheck - ${variant}`}
                        type="checkbox"
                        className={classNames('form-check-input', className)}
                        {...props}
                    />
                    <label className="form-check-label" htmlFor={`inputFieldsetCheck - ${variant}`}>
                        Checkbox
                    </label>
                    {message}
                </fieldset>
            )}
        />
        <WithVariants
            field={({ className, message, variant, ...props }) => (
                <fieldset className="form-check">
                    <input
                        id={`inputFieldsetRadio - ${variant}`}
                        type="radio"
                        className={classNames('form-check-input', className)}
                        name={`inputFieldsetRadio - ${variant}`}
                        {...props}
                    />
                    <label className="form-check-label" htmlFor={`inputFieldsetRadio - ${variant}`}>
                        Radio button
                    </label>
                    {message}
                </fieldset>
            )}
        />
    </div>
)
