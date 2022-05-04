import React, { ReactNode } from 'react'

import classNames from 'classnames'

import { Checkbox, RadioButton, Select, TextArea } from '@sourcegraph/wildcard'
import 'storybook-addon-designs'

import styles from './FormFieldVariants.module.scss'

type FieldVariants = 'standard' | 'invalid' | 'valid' | 'disabled'

interface WithVariantsProps {
    field: React.ComponentType<
        React.PropsWithChildren<{
            className?: string
            disabled?: boolean
            message?: ReactNode
            variant: FieldVariants
        }>
    >
}

const FieldMessageText = 'Helper text'

const FieldMessage: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({ className }) => (
    <small className={className}>{FieldMessageText}</small>
)

// Use this temporarily for form components which ones we haven't implemented in wilcard package yet
const WithVariantsAndMessageElements: React.FunctionComponent<React.PropsWithChildren<WithVariantsProps>> = ({
    field: Field,
}) => (
    <>
        <Field variant="standard" message={<FieldMessage className="field-message" />} />
        <Field variant="invalid" className="is-invalid" message={<FieldMessage className="invalid-feedback" />} />
        <Field variant="valid" className="is-valid" message={<FieldMessage className="valid-feedback" />} />
        <Field variant="disabled" disabled={true} message={<FieldMessage className="field-message" />} />
    </>
)

const WithVariants: React.FunctionComponent<React.PropsWithChildren<WithVariantsProps>> = ({ field: Field }) => (
    <>
        <Field variant="standard" message={FieldMessageText} />
        <Field variant="invalid" message={FieldMessageText} />
        <Field variant="valid" message={FieldMessageText} />
        <Field variant="disabled" message={FieldMessageText} />
    </>
)

export const FormFieldVariants: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <div className={styles.grid}>
        <WithVariantsAndMessageElements
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
            field={({ className, message, variant, ...props }) => (
                <Select
                    isCustomStyle={true}
                    className={className}
                    isValid={variant === 'invalid' ? false : variant === 'valid' ? true : undefined}
                    message={message}
                    disabled={variant === 'disabled'}
                    aria-label=""
                    {...props}
                >
                    <option>Option A</option>
                    <option>Option B</option>
                    <option>Option C</option>
                </Select>
            )}
        />
        <WithVariants
            field={({ className, message, variant, ...props }) => (
                <fieldset className="form-group">
                    <TextArea
                        message={message}
                        placeholder="This is sample content in a text area that spans four lines to see how it fits."
                        className={className}
                        rows={4}
                        isValid={variant === 'invalid' ? false : variant === 'valid' ? true : undefined}
                        disabled={variant === 'disabled'}
                        {...props}
                    />
                </fieldset>
            )}
        />
        <WithVariants
            field={({ className, message, variant, ...props }) => (
                <Checkbox
                    id={`inputFieldsetCheck - ${variant}`}
                    label="Checkbox"
                    className={className}
                    name={`inputFieldsetCheck - ${variant}`}
                    isValid={variant === 'invalid' ? false : variant === 'valid' ? true : undefined}
                    message={message}
                    disabled={variant === 'disabled'}
                    {...props}
                />
            )}
        />
        <WithVariants
            field={({ className, message, variant, ...props }) => (
                <RadioButton
                    id={`inputFieldsetRadio - ${variant}`}
                    className={className}
                    name={`inputFieldsetRadio - ${variant}`}
                    label="Radio button"
                    isValid={variant === 'invalid' ? false : variant === 'valid' ? true : undefined}
                    message={message}
                    disabled={variant === 'disabled'}
                    {...props}
                />
            )}
        />
    </div>
)
