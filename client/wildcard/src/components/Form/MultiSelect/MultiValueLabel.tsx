import { ReactElement } from 'react'

import { MultiValueGenericProps } from 'react-select'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Remove extra wrappers around multi value label
export const MultiValueLabel = <OptionValue extends unknown = unknown>({
    innerProps: _innerProps,
    selectProps: _selectProps,
    ...props
}: MultiValueGenericProps<MultiSelectOption<OptionValue>, true>): ReactElement => (
    <span className={styles.multiValueLabel} {...props} />
)
