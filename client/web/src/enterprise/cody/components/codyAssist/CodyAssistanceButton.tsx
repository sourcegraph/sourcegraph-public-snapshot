import React, { forwardRef } from 'react'

import { Button, ForwardReferenceComponent, type ButtonProps } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../CodyIcon'

interface Props extends Omit<ButtonProps, 'as' | 'ref'> {
    text?: string | null
}

/**
 * A button to get contextual assistance from Cody.
 */
export const CodyAssistanceButton = forwardRef(function CodyAssistanceButton(props, reference) {
    const { text = 'Cody assistance', ...otherProps } = props
    return (
        <Button ref={reference} {...otherProps}>
            <CodyIcon /> {text}
        </Button>
    )
}) as ForwardReferenceComponent<'button', Props>
