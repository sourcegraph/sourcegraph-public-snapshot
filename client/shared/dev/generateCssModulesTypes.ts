import { spawn } from 'child_process'
import path from 'path'

const REPO_ROOT = path.join(__dirname, '../../..')
const CSS_MODULES_GLOB = path.resolve(__dirname, '../../*/src/**/*.module.{scss,css}')
const JETBRAINS_CSS_MODULES_GLOB = path.resolve(__dirname, '../../jetbrains/webview/**/*.module.{scss,css}')
const TSM_COMMAND = `pnpm exec tsm --logLevel error "{${CSS_MODULES_GLOB},${JETBRAINS_CSS_MODULES_GLOB}}" --includePaths node_modules client`
const [BIN, ...ARGS] = TSM_COMMAND.split(' ')

/**
 * Generates the TypeScript types CSS modules.
 */
export function cssModulesTypings(extraArgs: string[] = []): Promise<void> {
    return new Promise((resolve, reject) => {
        const process = spawn(BIN, [...ARGS, ...extraArgs], {
            stdio: 'inherit',
            shell: true,
            cwd: REPO_ROOT,
        })
        process.on('exit', code => {
            if (code) {
                reject(new Error(`exited with code ${code}`))
            } else {
                resolve()
            }
        })
        process.on('error', err => reject(err))
    })
}

/**
 * Watch CSS modules and generates the TypeScript types for them.
 */
export function watchCSSModulesTypings(): Promise<void> {
    return cssModulesTypings(['--watch'])
}

if (require.main === module) {
    async function main(args: string[]) {
        cssModulesTypings(args)
    }
    main(process.argv.slice(2)).catch(error => {
        console.error(error)
        process.exit(1)
    })
}
