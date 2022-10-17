import { ReactElement } from 'react'

import { mdiClose } from '@mdi/js'
import { components, ClearIndicatorProps } from 'react-select'

import { Icon } from '../../Icon'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'

// Overwrite the clear indicator with `CloseIcon`
export const ClearIndicator = <OptionValue extends unknown = unknown>(
    props: ClearIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.ClearIndicator {...props}>
        <Icon className={styles.clearIcon} svgPath={mdiClose} inline={false} aria-hidden={true} />
    </components.ClearIndicator>
)
