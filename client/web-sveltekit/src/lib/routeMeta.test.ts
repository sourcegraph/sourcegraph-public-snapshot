import { test, expect } from 'vitest'

import { generateRouteMeta_testOnly as generateRouteMeta } from './routeMeta'

test('pages inherit layout data properly', () => {
    const rawRouteMeta = {
        '../routes/a/b/+page.ts': {
            serverRouteName: 'ab',
        },
        '../routes/a/b/+layout.ts': {
            layoutProp: 'ab',
        },
        '../routes/a/+page.ts': {
            serverRouteName: 'a',
        },
        '../routes/+layout.ts': {
            layoutProp: 'root',
            rootProp: 'a',
        },
    }

    const routeMeta = generateRouteMeta(rawRouteMeta)

    expect(routeMeta).toEqual({
        '/a/b': {
            serverRouteName: 'ab',
            layoutProp: 'ab',
            rootProp: 'a',
        },
        '/a': {
            serverRouteName: 'a',
            layoutProp: 'root',
            rootProp: 'a',
        },
    })
})
