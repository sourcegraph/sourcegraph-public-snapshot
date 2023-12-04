import React from 'react'

import { ConnectionForm } from '../../../../components/FilteredConnection/ui'
import type { ConnectionFormProps } from '../../../../components/FilteredConnection/ui/ConnectionForm'

import styles from './ConnectionPopoverForm.module.scss'

type ConnectionPopoverFormProps = ConnectionFormProps

export const ConnectionPopoverForm: React.FunctionComponent<React.PropsWithChildren<ConnectionPopoverFormProps>> = ({
    inputClassName,
    formClassName,
    ...rest
}) => <ConnectionForm inputClassName={styles.connectionPopoverInput} formClassName={formClassName} {...rest} />
