import className from 'classnames'
import { upperFirst } from 'lodash'

import type { ALERT_VARIANTS } from './constants'

import styles from './Alert.module.scss'

interface GetAlertStyleParameters {
    variant: typeof ALERT_VARIANTS[number]
}

export const getAlertStyle = ({ variant }: GetAlertStyleParameters): string =>
    className(styles[`alert${upperFirst(variant)}` as keyof typeof styles])
