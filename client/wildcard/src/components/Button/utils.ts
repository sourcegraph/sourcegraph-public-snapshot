import { upperFirst } from 'lodash'

import { BUTTON_VARIANTS, BUTTON_SIZES, BUTTON_DISPLAY } from './constants'

import styles from './Button.module.scss'

interface GetButtonStyleParameters {
    variant: typeof BUTTON_VARIANTS[number]
    outline?: boolean
}

export const getButtonStyle = ({ variant, outline }: GetButtonStyleParameters): string =>
    styles[`btn${outline ? 'Outline' : ''}${upperFirst(variant)}` as keyof typeof styles]

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
