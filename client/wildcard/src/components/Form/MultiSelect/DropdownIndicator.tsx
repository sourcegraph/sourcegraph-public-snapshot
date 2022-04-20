import { ReactElement } from 'react'

import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import { components, DropdownIndicatorProps } from 'react-select'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the dropdown indicator with `ChevronDownIcon`
export const DropdownIndicator = <OptionValue extends unknown = unknown>(
    props: DropdownIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.DropdownIndicator {...props}>
        <ChevronDownIcon className={props.isDisabled ? styles.dropdownIconDisabled : styles.dropdownIcon} />
    </components.DropdownIndicator>
)
