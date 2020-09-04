import isChromatic from '../types/chromatic/isChromatic'

/**
 * Returns `true` if the code is running inside Chromatic or Percy,
 * where certain UI elements should be hidden that could cause flaky screenshots.
 */
export const isScreenshotTestEnvironment = (): boolean => isChromatic() || window.matchMedia('only percy').matches
