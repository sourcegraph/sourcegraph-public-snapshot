export const DEFAULT_EXTENSIONS = [
  '.css',
  '.scss',
  '.sass',
  '.stylus',
  '.styl',
  '.less',
]

export let oldHandlers = {}

export function noOp() {}

export function restore() {
  for (const ext in oldHandlers) {
    if (oldHandlers[ext] === undefined) {
      delete require.extensions[ext]
    } else {
      require.extensions[ext] = oldHandlers[ext]
    }
  }

  oldHandlers = {}
}

export default function register(extensions = DEFAULT_EXTENSIONS, handler = noOp) {
  restore()

  for (const ext of extensions) {
    oldHandlers[ext] = require.extensions[ext]
    require.extensions[ext] = handler
  }
}

// Run at import
register()
