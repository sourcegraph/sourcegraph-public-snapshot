import classNames from 'classnames'
import { upperFirst } from 'lodash'

import type { TYPOGRAPHY_ALIGNMENTS, TYPOGRAPHY_MODES, TYPOGRAPHY_WEIGHTS } from './constants'

import styles from './Typography.module.scss'

export interface TypographyProps {
    alignment?: typeof TYPOGRAPHY_ALIGNMENTS[number]
    mode?: typeof TYPOGRAPHY_MODES[number]
    as?: React.ElementType
}

export interface GetAlignmentStyleParameters {
    alignment?: typeof TYPOGRAPHY_ALIGNMENTS[number]
}

export interface GetModeStyleParameters {
    mode?: typeof TYPOGRAPHY_MODES[number]
}

export interface GetFontWeightStyleParameters {
    weight?: typeof TYPOGRAPHY_WEIGHTS[number]
}

export const getAlignmentStyle = ({ alignment }: GetAlignmentStyleParameters): string =>
    classNames(styles[`align${upperFirst(alignment)}` as keyof typeof styles])

export const getFontWeightStyle = ({ weight }: GetFontWeightStyleParameters): string =>
    classNames(styles[`fontWeight${upperFirst(weight)}` as keyof typeof styles])

export const getModeStyle = ({ mode }: GetModeStyleParameters): string => {
    switch (mode) {
        case 'single-line': {
            return styles.singleLine
        }
        case 'break-word': {
            return styles.breakWord
        }
        default: {
            return ''
        }
    }
}
