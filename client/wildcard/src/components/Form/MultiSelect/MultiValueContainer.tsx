import React, { ReactElement } from 'react'

import { MultiValueGenericProps } from 'react-select'

import { Badge } from '../../Badge'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the multi value container with Wildcard `Badge`
export const MultiValueContainer = <OptionValue extends unknown = unknown>({
    innerProps: _innerProps,
    selectProps: _selectProps,
    ...props
}: MultiValueGenericProps<MultiSelectOption<OptionValue>, true>): ReactElement => (
    <Badge variant="secondary" className={styles.multiValueContainer} {...props} />
)
