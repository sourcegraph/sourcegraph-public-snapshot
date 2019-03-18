// @ts-ignore
import babelPluginIstanbul from 'babel-plugin-istanbul'
import babelify from 'babelify'
import browserify from 'browserify'
import getStream from 'get-stream'
import { DOMWindow, JSDOM } from 'jsdom'
import ora from 'ora'
// @ts-ignore
import tsify from 'tsify'

interface DOMModuleSandbox {
    /**
     * The JSDOM window the module was loaded into
     */
    window: DOMWindow

    /**
     * The loaded module
     */
    module: any
}

export interface TestBundle {
    /**
     * Creates a JSDOM instance and loads the module into the window
     */
    load(): DOMModuleSandbox
}

/**
 * Bundles a given module for testing purposes
 *
 * @param modulePath The absolute path to the TypeScript module file including extension
 */
export const createTestBundle = async (modulePath: string): Promise<TestBundle> => {
    const bundler = browserify(modulePath, {
        // Generate sourcemaps for debugging
        debug: true,
        // Expose the module under this global variable
        standalone: 'moduleUnderTest',
        // rootDir for sourcemaps
        basedir: __dirname + '/../../src',
    })
    bundler.plugin(tsify, { project: __dirname + '/../../tsconfig.test.json' })

    // If running through nyc, instrument bundle for coverage reporting too
    if (process.env.NYC_CONFIG) {
        bundler.transform(babelify.configure({ plugins: [babelPluginIstanbul], extensions: ['.tsx', '.ts'] }))
    }

    const spinner = ora(`Bundling ${modulePath}`)
    bundler.on('file', (file, id) => {
        spinner.text = `Bundling ${id}`
    })
    spinner.start()
    try {
        const bundle = await getStream(bundler.bundle())
        return {
            load(): DOMModuleSandbox {
                const jsdom = new JSDOM('', { runScripts: 'dangerously' }) as JSDOM & {
                    window: { moduleUnderTest: any }
                }
                jsdom.window.eval(bundle)
                return { window: jsdom.window, module: jsdom.window.moduleUnderTest }
            },
        }
    } finally {
        spinner.stop()
    }
}
