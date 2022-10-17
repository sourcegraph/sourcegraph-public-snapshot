const styles = getComputedStyle(document.documentElement)

export const VIEWPORT_SM = parseInt(styles.getPropertyValue('--viewport-sm'), 10)
export const VIEWPORT_MD = parseInt(styles.getPropertyValue('--viewport-md'), 10)
export const VIEWPORT_LG = parseInt(styles.getPropertyValue('--viewport-lg'), 10)
export const VIEWPORT_XL = parseInt(styles.getPropertyValue('--viewport-xl'), 10)

export const VIEWPORTS = {
    sm: VIEWPORT_SM,
    md: VIEWPORT_MD,
    lg: VIEWPORT_LG,
    xl: VIEWPORT_XL,
}
