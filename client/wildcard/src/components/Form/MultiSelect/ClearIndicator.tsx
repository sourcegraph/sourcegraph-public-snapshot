import { ReactElement } from 'react'

import CloseIcon from 'mdi-react/CloseIcon'
import { components, ClearIndicatorProps } from 'react-select'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the clear indicator with `CloseIcon`
export const ClearIndicator = <OptionValue extends unknown = unknown>(
    props: ClearIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.ClearIndicator {...props}>
        <CloseIcon className={styles.clearIcon} />
    </components.ClearIndicator>
)
