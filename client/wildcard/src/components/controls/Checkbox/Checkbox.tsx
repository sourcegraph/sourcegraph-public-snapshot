import React from 'react'
import { BaseControlInput } from '../BaseControlInput'

export const Checkbox: typeof BaseControlInput = React.forwardRef((props, reference) => (
    <BaseControlInput {...props} type="checkbox" ref={reference} />
))
