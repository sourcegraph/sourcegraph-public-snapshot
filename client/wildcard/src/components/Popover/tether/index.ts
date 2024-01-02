// Tether primitive models and structures
export * from './models/tether-models'
export * from './models/geometry/point'
export * from './models/geometry/rectangle'

// The main entry point of tooltip positioning engine
export { createTether } from './services/tether-registry'

// Shared type interfaces
export type { Tether } from './services/types'
export type { TetherInstanceAPI } from './services/tether-registry'
export { TetherAPI } from './services/tether-registry'
