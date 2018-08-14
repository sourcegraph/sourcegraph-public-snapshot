import { noop } from 'lodash'
import signale from 'signale'
import webpack from 'webpack'
import config from '../webpack/dev.config'
import * as autoReloading from './auto-reloading'
import * as tasks from './tasks'

signale.config({ displayTimestamp: true })

const triggerReload = process.env.AUTO_RELOAD === 'false' ? noop : autoReloading.initializeServer()

const buildChrome = tasks.buildChrome('dev')
const buildFirefox = tasks.buildFirefox('dev')

tasks.copyAssets('dev')

const compiler = webpack(config)

signale.info('Running webpack')

compiler.hooks.watchRun.tap('Notify', () => signale.await('Compiling...'))

compiler.watch(
    {
        aggregateTimeout: 300,
        poll: 1000,
    },
    (err, stats) => {
        console.log(stats.toString('errors-only'))

        if (stats.hasErrors()) {
            return
        }

        tasks.buildSafari('dev')
        buildChrome()
        buildFirefox()
        triggerReload()
    }
)
