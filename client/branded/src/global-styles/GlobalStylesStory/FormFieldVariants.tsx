import classNames from 'classnames'
import React from 'react'
import 'storybook-addon-designs'

import styles from './FormFieldVariants.module.scss'

interface WithVariantsProps {
    field: React.ComponentType<{ className?: string; disabled?: boolean; message?: JSX.Element }>
}

const FieldMessage: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <small className={className}>Helper text</small>
)

const WithVariants: React.FunctionComponent<WithVariantsProps> = ({ field: Field }) => (
    <>
        <Field message={<FieldMessage />} />
        <Field className="is-invalid" message={<FieldMessage className="invalid-feedback" />} />
        <Field className="is-valid" message={<FieldMessage className="valid-feedback" />} />
        <Field disabled={true} message={<FieldMessage />} />
    </>
)

export const FormFieldVariants: React.FunctionComponent = () => {
    return (
        <div className={styles.grid}>
            <WithVariants
                field={({ className, message, ...props }) => (
                    <div className="form-group">
                        <input
                            type="text"
                            placeholder="Form field"
                            className={classNames('form-control', className)}
                            {...props}
                        />
                        {message}
                    </div>
                )}
            />
            <WithVariants
                field={({ className, message, ...props }) => (
                    <div className="form-group">
                        <select className={classNames('form-control', className)} {...props}>
                            <option>Option A</option>
                            <option>Option B</option>
                            <option>Option C</option>
                        </select>
                        {message}
                    </div>
                )}
            />
            <WithVariants
                field={({ className, message, ...props }) => (
                    <div className="form-group">
                        <textarea
                            placeholder="This is sample content in a text area that spans three lines to see how it fits."
                            className={classNames('form-control', className)}
                            {...props}
                        />
                        {message}
                    </div>
                )}
            />
            <WithVariants
                field={({ className, message, ...props }) => (
                    <div className="form-check">
                        <input
                            id="inputFieldsetCheck"
                            type="checkbox"
                            className={classNames('form-check-input', className)}
                            {...props}
                        />{' '}
                        <label className={classNames('form-check-label')} htmlFor="inputFieldsetCheck">
                            Checkbox
                        </label>
                    </div>
                )}
            />
        </div>
    )
}
