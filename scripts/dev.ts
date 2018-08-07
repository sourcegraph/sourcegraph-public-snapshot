import webpack from 'webpack'
import config from '../webpack/dev.config'
import * as autoReloading from './auto-reloading'
import * as tasks from './tasks'

// tslint:disable-next-line:no-empty
const noop = () => {}
const triggerReload = process.env.AUTO_RELOAD === 'false' ? noop : autoReloading.initializeServer()

const buildChrome = tasks.buildChrome('dev')
const buildFirefox = tasks.buildFirefox('dev')

console.log('Copying Assets...')
tasks.copyAssets('dev')
console.log('Done copying assets.')

const compiler = webpack(config)

console.info('[Webpack Dev]')
console.info('--------------------------------')

compiler.watch(
    {
        aggregateTimeout: 300,
        poll: 1000,
    },
    (err, stats) => {
        console.log(stats.toString('normal'))

        if (stats.hasErrors()) {
            return
        }

        tasks.buildSafari('dev')
        buildChrome()
        buildFirefox()
        triggerReload()
    }
)
