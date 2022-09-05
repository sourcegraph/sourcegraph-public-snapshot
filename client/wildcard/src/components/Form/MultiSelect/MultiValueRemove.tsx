import { ReactElement } from 'react'

import { mdiClose } from '@mdi/js'
import { components, MultiValueRemoveProps } from 'react-select'

import { Icon } from '../../Icon'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the multi value remove indicator with `CloseIcon`
export const MultiValueRemove = <OptionValue extends unknown = unknown>(
    props: MultiValueRemoveProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.MultiValueRemove {...props}>
        <Icon className={styles.removeIcon} svgPath={mdiClose} inline={false} aria-hidden={true} />
    </components.MultiValueRemove>
)
