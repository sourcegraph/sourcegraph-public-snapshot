// eslint-disable-next-line no-restricted-imports
import isChromaticDefault from 'chromatic/isChromatic'

export const isChromatic = (): boolean => isChromaticDefault() || Boolean(process.env.CHROMATIC)
