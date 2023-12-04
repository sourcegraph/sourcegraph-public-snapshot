/**
 * Mute Polly.js error output on CI that's happening despite having `logging: false`.
 * It's a known bug that was fixed in the next major version.
 *
 * The PR that tries to fix the issue: https://github.com/Netflix/pollyjs/pull/377
 * The PR that fixed the issue in the next major version: https://github.com/Netflix/pollyjs/pull/427
 *
 * Logger logic that is suppressed by this module:
 * https://github.com/davidNHK/pollyjs/blob/3e876a8cc0b28e8ef422763762bdab10027bb25d/packages/%40pollyjs/core/src/-private/logger.js
 *
 */
if (process.env.CI || process.env.LOG_BROWSER_CONSOLE === 'false') {
  const originalLog = console.log
  const originalError = console.error
  const originalGroup = console.group
  const originalGroupEnd = console.groupEnd

  let areLogsEnabled = true

  console.group = title => {
    if (areLogsEnabled === false) {
      return
    }

    // Rely on the recording prefix used by Polly to group the error output.
    if (typeof title === 'string' && (title.startsWith('[SG_POLLY]') || title.startsWith('Errored ➞'))) {
      areLogsEnabled = false

      // Re-enable logging on `groupEnd`.
      console.groupEnd = () => {
        areLogsEnabled = true
        console.groupEnd = originalGroupEnd
      }

      return
    }

    return originalGroup(title)
  }

  console.log = (...args) => {
    if (areLogsEnabled) {
      originalLog(...args)
    }
  }

  console.error = (...args) => {
    if (areLogsEnabled) {
      originalError(...args)
    }
  }
}
