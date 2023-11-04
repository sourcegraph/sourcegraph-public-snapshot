import path from 'path'

import { BAZEL, defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        environmentMatchGlobs: [
            // TODO(sqs): can't use jsdom because it breaks simmerjs
            // (https://github.com/jsdom/jsdom/issues/3612#issuecomment-1778560104)
            ['src/**/domFunctions.test.tsx', 'happy-dom'],
        ],

        setupFiles: [
            'src/testSetup.test.ts',
            '../testing/src/reactCleanup.ts',
            '../testing/src/fetch.js',
            '../testing/src/mockUniqueId.ts',
        ],

        // For some reason, watch mode fails with `Error: Failed to terminate worker` unless
        // singleThread is true. See https://github.com/vitest-dev/vitest/issues/3077.
        singleThread: true,
    },

    plugins: BAZEL
        ? [
              {
                  // The github/codeHost.tsx file imports sourcegraph-mark.svg, but this is not
                  // needed for any tests. Just ignore it.
                  name: 'no-sourcegrah-mark-svg',
                  resolveId(id) {
                      if (id.endsWith('/sourcegraph-mark.svg')) {
                          return { id, external: true }
                      }
                      return undefined
                  },
              },
          ]
        : undefined,
})
