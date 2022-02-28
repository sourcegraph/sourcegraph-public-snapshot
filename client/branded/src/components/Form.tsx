import classNames from 'classnames'
import * as React from 'react'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

interface FormProps extends React.DetailedHTMLProps<React.FormHTMLAttributes<HTMLFormElement>, HTMLFormElement> {
    children: React.ReactNode
}

export const Form = React.forwardRef((props, reference) => {
    const { as: Component = 'form', children, className, onInvalid, ...otherProps } = props

    const [wasValidated, setWasValidated] = React.useState(false)

    const localOnInvalid = (event: React.InvalidEvent<HTMLFormElement>): void => {
        setWasValidated(true)
        if (onInvalid) {
            onInvalid(event)
        }
    }

    return (
        // eslint-disable-next-line react/forbid-elements
        <form
            ref={reference}
            className={classNames(className, wasValidated && 'was-validated')}
            onInvalid={localOnInvalid}
            {...otherProps}
        >
            {children}
        </form>
    )
}) as ForwardReferenceComponent<'form', FormProps>
