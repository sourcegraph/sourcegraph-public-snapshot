import { ReactElement } from 'react'

import { components, ClearIndicatorProps } from 'react-select'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'
import { mdiClose } from "@mdi/js";
import { Icon } from "@sourcegraph/wildcard";

// Overwrite the clear indicator with `CloseIcon`
export const ClearIndicator = <OptionValue extends unknown = unknown>(
    props: ClearIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.ClearIndicator {...props}>
        <Icon className={styles.clearIcon} svgPath={mdiClose} inline={false} aria-hidden={true} />
    </components.ClearIndicator>
)
