import { BUTTON_VARIANTS, BUTTON_SIZES } from './constants'

interface GetButtonStyleParameters {
    variant: typeof BUTTON_VARIANTS[number]
    outline?: boolean
}

export const getButtonStyle = ({ variant, outline }: GetButtonStyleParameters): string => {
    if (variant === 'link') {
        return 'btn-link'
    }

    return `btn${outline ? '-outline' : ''}-${variant}`
}

interface GetButtonSizeParameters {
    size: typeof BUTTON_SIZES[number]
}

export const getButtonSize = ({ size }: GetButtonSizeParameters): string => `btn-${size}`
