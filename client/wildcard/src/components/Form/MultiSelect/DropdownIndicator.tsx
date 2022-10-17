import { ReactElement } from 'react'

import { mdiChevronDown } from '@mdi/js'
import { components, DropdownIndicatorProps } from 'react-select'

import { Icon } from '../../Icon'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the dropdown indicator with `ChevronDownIcon`
export const DropdownIndicator = <OptionValue extends unknown = unknown>(
    props: DropdownIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.DropdownIndicator {...props}>
        <Icon
            className={props.isDisabled ? styles.dropdownIconDisabled : styles.dropdownIcon}
            svgPath={mdiChevronDown}
            inline={false}
            aria-hidden={true}
        />
    </components.DropdownIndicator>
)
