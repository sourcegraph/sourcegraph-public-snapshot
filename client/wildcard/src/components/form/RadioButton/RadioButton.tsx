import React from 'react'

import { BaseControlInput } from '../internal/BaseControlInput'

export const RadioButton: typeof BaseControlInput = React.forwardRef((props, reference) => (
    <BaseControlInput {...props} type="radio" ref={reference} />
))
