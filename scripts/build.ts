import signale from 'signale'
import webpack from 'webpack'
import config from '../webpack/prod.config'
import * as tasks from './tasks'

const buildChrome = tasks.buildChrome('prod')
const buildFirefox = tasks.buildFirefox('prod')

signale.info('[Copy assets]')
signale.info('--------------------------------')
tasks.copyAssets('prod')

const compiler = webpack(config)

signale.info('[Webpack Build]')
signale.info('--------------------------------')

compiler.run((err, stats) => {
    console.log(stats.toString('normal'))

    if (stats.hasErrors()) {
        process.exit(1)
        return
    }

    tasks.buildSafari('prod')
    buildChrome()
    buildFirefox()
    tasks.copyPhabricator()
})
