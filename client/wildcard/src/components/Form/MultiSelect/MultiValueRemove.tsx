import { ReactElement } from 'react'

import CloseIcon from 'mdi-react/CloseIcon'
import { components, MultiValueRemoveProps } from 'react-select'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the multi value remove indicator with `CloseIcon`
export const MultiValueRemove = <OptionValue extends unknown = unknown>(
    props: MultiValueRemoveProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.MultiValueRemove {...props}>
        <CloseIcon className={styles.removeIcon} />
    </components.MultiValueRemove>
)
