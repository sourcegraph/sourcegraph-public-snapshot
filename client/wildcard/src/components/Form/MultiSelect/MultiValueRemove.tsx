import { ReactElement } from 'react'

import { components, MultiValueRemoveProps } from 'react-select'

import { MultiSelectOption } from './types'

import styles from './MultiSelect.module.scss'
import { mdiClose } from "@mdi/js";
import { Icon } from "@sourcegraph/wildcard";

// Overwrite the multi value remove indicator with `CloseIcon`
export const MultiValueRemove = <OptionValue extends unknown = unknown>(
    props: MultiValueRemoveProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.MultiValueRemove {...props}>
        <Icon className={styles.removeIcon} svgPath={mdiClose} inline={false} aria-hidden={true} />
    </components.MultiValueRemove>
)
