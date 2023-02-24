// Tether primitive models and structures
export * from './models/geometry/point'
export * from './models/geometry/rectangle'
export * from './models/tether-models'
// The main entry point of tooltip positioning engine
export { createTether } from './services/tether-registry'
export type { TetherInstanceAPI } from './services/tether-registry'
// Shared type interfaces
export type { Tether } from './services/types'
