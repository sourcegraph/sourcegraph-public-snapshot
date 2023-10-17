import classNames from 'classnames'

import {
    getAlignmentStyle,
    type GetAlignmentStyleParameters,
    getFontWeightStyle,
    type GetFontWeightStyleParameters,
    getModeStyle,
    type GetModeStyleParameters,
} from '../utils'

import typographyStyles from '../Typography.module.scss'
import styles from './Label.module.scss'

interface GetLabelClassNameParameters
    extends GetAlignmentStyleParameters,
        GetModeStyleParameters,
        GetFontWeightStyleParameters {
    size?: 'small' | 'base'
    weight?: 'regular' | 'medium' | 'bold'
    isUnderline?: boolean
    isUppercase?: boolean
}

export function getLabelClassName({
    mode,
    size,
    weight,
    alignment,
    isUnderline,
    isUppercase,
}: GetLabelClassNameParameters = {}): string {
    return classNames(
        styles.label,
        isUnderline && styles.labelUnderline,
        isUppercase && styles.labelUppercase,
        size === 'small' && typographyStyles.small,
        weight && getFontWeightStyle({ weight }),
        alignment && getAlignmentStyle({ alignment }),
        mode && getModeStyle({ mode }),
        mode === 'single-line' && styles.labelSingleLine
    )
}
