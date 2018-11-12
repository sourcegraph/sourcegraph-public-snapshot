import { ChildProcess, spawn } from 'child_process'
import gulp from 'gulp'
import path from 'path'

export function build(): ChildProcess {
    return spawn('yarn', ['-s', 'run', 'build'], {
        stdio: 'inherit',
        shell: true,
    })
}

export function watch(): ChildProcess {
    return spawn('yarn', ['-s', 'run', 'dev'], {
        stdio: 'inherit',
        shell: true,
    })
}

const PHABRICATOR_EXTENSION_FILES = path.join(__dirname, './build/phabricator/**')

/**
 * Copies the phabricator extension over to the ui/assets folder so they can be served by the webapp. The package
 * is published from ./client/browser.
 */
export function phabricator(): NodeJS.ReadWriteStream {
    return gulp.src(PHABRICATOR_EXTENSION_FILES).pipe(gulp.dest('../../ui/assets/extension'))
}

export const watchPhabricator = gulp.series(phabricator, async function watchPhabricator(): Promise<void> {
    await new Promise<never>((_, reject) => {
        gulp.watch(PHABRICATOR_EXTENSION_FILES, phabricator).on('error', reject)
    })
})
