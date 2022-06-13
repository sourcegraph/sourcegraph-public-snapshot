// eslint-disable-next-line no-restricted-imports
import isChromaticDefault from 'chromatic/isChromatic'

/**
 * `chromatic/isChromatic` wrapper that takes into account `process.env.CHROMATIC` for local testing.
 */
export const isChromatic = (): boolean => isChromaticDefault() || Boolean(process.env.CHROMATIC)
