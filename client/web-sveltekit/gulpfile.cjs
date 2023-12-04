// @ts-check
const { spawn } = require('child_process')

/**
 * buildSvelteKit is supposed to be used to build the SvelteKit
 * artifacts for the enterprise development build.
 */
function buildSvelteKit() {
    // We cannot progmatically start Vite because the SvelteKit plugin
    // expects the current working directory to be the project root
    // (client/web-sveltekit in our case). Hence we spawn a child
    // process with the correct working directory.
    return spawn('pnpm', ['run', 'build', '-l', 'error'], {
        stdio: 'inherit',
        cwd: __dirname,
        shell: true,
        env: { ...process.env },
    })
}

module.exports = {
    buildSvelteKit,
}
