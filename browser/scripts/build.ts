import signale from 'signale'
import webpack from 'webpack'
import config from '../config/webpack/prod.config'
import * as tasks from './tasks'

const buildChrome = tasks.buildChrome('prod')
const buildFirefox = tasks.buildFirefox('prod')

tasks.copyAssets('prod')

const compiler = webpack(config)

signale.await('Webpack compilation')

compiler.run((err, stats) => {
    console.log(stats.toString(tasks.WEBPACK_STATS_OPTIONS))

    if (stats.hasErrors()) {
        signale.error('Webpack compilation error')
        process.exit(1)
        return
    }
    signale.success('Webpack compilation done')

    buildChrome()
    buildFirefox()
    tasks.copyPhabricator()
    signale.success('Build done')
})
