import CloseIcon from 'mdi-react/CloseIcon'
import React, { ReactElement } from 'react'
import { components, MultiValueRemoveProps } from 'react-select'

import styles from './MultiSelect.module.scss'
import { MultiSelectOption } from './types'

// Overwrite the multi value remove indicator with `CloseIcon`
export const MultiValueRemove = <OptionValue extends unknown = unknown>(
    props: MultiValueRemoveProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.MultiValueRemove {...props}>
        <CloseIcon className={styles.removeIcon} />
    </components.MultiValueRemove>
)
