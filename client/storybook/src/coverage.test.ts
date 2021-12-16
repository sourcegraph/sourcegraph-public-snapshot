import path from 'path'
import { pathToFileURL } from 'url'

import initStoryshots from '@storybook/addon-storyshots'
import { puppeteerTest } from '@storybook/addon-storyshots-puppeteer'
import { PUPPETEER_REVISIONS } from 'puppeteer/lib/cjs/puppeteer/revisions'

import { recordCoverage } from '@sourcegraph/shared/src/testing/coverage'
import { getPuppeteerBrowser } from '@sourcegraph/shared/src/testing/driver'

// This test suite does not actually test anything.
// It just loads up the storybook in Puppeteer and records its coverage,
// so it can be tracked in Codecov.

initStoryshots({
    configPath: __dirname,
    suite: 'Storybook',
    test: puppeteerTest({
        chromeExecutablePath: getPuppeteerBrowser('chrome', PUPPETEER_REVISIONS.chromium).executablePath,
        storybookUrl: pathToFileURL(path.resolve(__dirname, '../storybook-static')).href,
        testBody: async page => {
            await recordCoverage(page.browser())
        },
    }),
})
