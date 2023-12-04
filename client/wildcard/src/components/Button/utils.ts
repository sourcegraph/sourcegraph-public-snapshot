import classNames from 'classnames'
import { upperFirst } from 'lodash'

import type { BUTTON_VARIANTS, BUTTON_SIZES, BUTTON_DISPLAY } from './constants'

import styles from './Button.module.scss'

interface GetButtonStyleParameters {
    variant: typeof BUTTON_VARIANTS[number]
    outline?: boolean
}

export const getButtonStyle = ({ variant, outline }: GetButtonStyleParameters): string =>
    classNames(styles[`btn${upperFirst(variant)}` as keyof typeof styles], outline && styles.btnOutline)

interface GetButtonSizeParameters {
    size: typeof BUTTON_SIZES[number]
}

export const getButtonSize = ({ size }: GetButtonSizeParameters): string =>
    styles[`btn${upperFirst(size)}` as keyof typeof styles]

interface GetButtonDisplayParameters {
    display: typeof BUTTON_DISPLAY[number]
}

export const getButtonDisplay = ({ display }: GetButtonDisplayParameters): string =>
    styles[`btn${upperFirst(display)}` as keyof typeof styles]

/**
 * Returns the class name to style a button with the given options. This can be
 * used to for generating the right CSS class combination for plain DOM buttons,
 * but it should be used sparingly.
 */
export function getButtonClassName({
    variant,
    display,
    size,
    outline,
}: Partial<GetButtonStyleParameters & GetButtonSizeParameters & GetButtonDisplayParameters> = {}): string {
    return classNames(
        styles.btn,
        variant && getButtonStyle({ variant, outline }),
        display && getButtonDisplay({ display }),
        size && getButtonSize({ size })
    )
}
