import classNames from 'classnames'
import { upperFirst } from 'lodash'

import type { PANEL_POSITIONS } from './constants'

import styles from './Panel.module.scss'

interface GetDisplayStyleParameters {
    isFloating?: boolean
}

export const getDisplayStyle = ({ isFloating }: GetDisplayStyleParameters): string =>
    classNames(styles[`panel${upperFirst(isFloating ? 'fixed' : 'relative')}` as keyof typeof styles])

interface GetPositionStyleParameters {
    position?: typeof PANEL_POSITIONS[number]
}

export const getPositionStyle = ({ position }: GetPositionStyleParameters): string =>
    classNames(styles[`panel${upperFirst(position)}` as keyof typeof styles])
