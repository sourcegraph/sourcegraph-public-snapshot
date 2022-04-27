/**
 * Utility for computing the PX equivalent of a value in REM.
 *
 * @param rem Size in REM
 */
export const convertREMToPX = (rem: number): number =>
    rem * parseFloat(getComputedStyle(document.documentElement).fontSize)
