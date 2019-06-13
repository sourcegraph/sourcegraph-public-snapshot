import { hex } from 'wcag-contrast'

const BLACK = '#000'
const WHITE = '#fff'

/**
 * Returns the hex color code (either black or white) for foreground text that has the most contrast
 * with the given background color.
 *
 * @param backgroundColor A hex color code (e.g., `#a7bb8c`).
 */
export const contrastingForegroundColor = (backgroundColor: string): typeof BLACK | typeof WHITE => {
    const blackContrast = hex(backgroundColor, BLACK)
    const whiteContrast = hex(backgroundColor, WHITE)
    return blackContrast > whiteContrast ? BLACK : WHITE
}
