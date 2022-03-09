import React, { ReactElement } from 'react'
import { MultiValueGenericProps } from 'react-select'

import { Badge } from '../../Badge'

import styles from './MultiSelect.module.scss'
import { MultiSelectOption } from './types'

// Overwrite the multi value container with Wildcard `Badge`
export const MultiValueContainer = <OptionValue extends unknown = unknown>({
    innerProps: _innerProps,
    selectProps: _selectProps,
    ...props
}: MultiValueGenericProps<MultiSelectOption<OptionValue>, true>): ReactElement => (
    <Badge variant="secondary" className={styles.multiValueContainer} {...props} />
)
