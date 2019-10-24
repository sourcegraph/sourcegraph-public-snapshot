import { ChildProcess, spawn } from 'child_process'
import gulp from 'gulp'
import path from 'path'

export function build(): ChildProcess {
    return spawn('yarn', ['-s', 'run', 'build'], {
        stdio: 'inherit',
        shell: true,
        env: { ...process.env, NODE_OPTIONS: '--max_old_space_size=8192' },
    })
}

export function watch(): ChildProcess {
    return spawn('yarn', ['-s', 'run', 'dev'], {
        stdio: 'inherit',
        shell: true,
        env: { ...process.env, NODE_OPTIONS: '--max_old_space_size=8192' },
    })
}

const INTEGRATION_FILES = path.join(__dirname, './build/integration/**')
const INTEGRATION_ASSETS_DIRECTORY = path.join(__dirname, '../ui/assets/extension')

/**
 * Copies the phabricator extension over to the ui/assets folder so they can be served by the webapp. The package
 * is published from ./browser.
 */
export function copyIntegrationAssets(): NodeJS.ReadWriteStream {
    return gulp.src(INTEGRATION_FILES).pipe(gulp.dest(INTEGRATION_ASSETS_DIRECTORY))
}

export const watchIntegrationAssets = gulp.series(
    copyIntegrationAssets,
    async function watchIntegrationAssets(): Promise<void> {
        await new Promise<never>((_, reject) => {
            gulp.watch(INTEGRATION_FILES, copyIntegrationAssets).on('error', reject)
        })
    }
)
