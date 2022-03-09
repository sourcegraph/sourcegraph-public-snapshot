import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import React, { ReactElement } from 'react'
import { components, DropdownIndicatorProps } from 'react-select'

import styles from './MultiSelect.module.scss'
import { MultiSelectOption } from './types'

// Overwrite the dropdown indicator with `ChevronDownIcon`
export const DropdownIndicator = <OptionValue extends unknown = unknown>(
    props: DropdownIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.DropdownIndicator {...props}>
        <ChevronDownIcon className={props.isDisabled ? styles.dropdownIconDisabled : styles.dropdownIcon} />
    </components.DropdownIndicator>
)
