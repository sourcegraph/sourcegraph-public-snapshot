import React, { ReactElement } from 'react'
import { MultiValueGenericProps } from 'react-select'

import { MultiSelectOption } from './types'

// Remove extra wrappers around multi value label
export const MultiValueLabel = <OptionValue extends unknown = unknown>({
    innerProps: _innerProps,
    selectProps: _selectProps,
    ...props
}: MultiValueGenericProps<MultiSelectOption<OptionValue>, true>): ReactElement => <span {...props} />
