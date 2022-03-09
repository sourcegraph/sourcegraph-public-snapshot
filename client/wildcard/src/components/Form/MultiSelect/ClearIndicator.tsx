import CloseIcon from 'mdi-react/CloseIcon'
import React, { ReactElement } from 'react'
import { components, ClearIndicatorProps } from 'react-select'

import styles from './MultiSelect.module.scss'
import { MultiSelectOption } from './types'

// Overwrite the clear indicator with `CloseIcon`
export const ClearIndicator = <OptionValue extends unknown = unknown>(
    props: ClearIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.ClearIndicator {...props}>
        <CloseIcon className={styles.clearIcon} />
    </components.ClearIndicator>
)
