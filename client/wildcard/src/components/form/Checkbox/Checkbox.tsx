import React from 'react'

import { BaseControlInput } from '../internal/BaseControlInput'

export const Checkbox: typeof BaseControlInput = React.forwardRef((props, reference) => (
    <BaseControlInput {...props} type="checkbox" ref={reference} />
))
