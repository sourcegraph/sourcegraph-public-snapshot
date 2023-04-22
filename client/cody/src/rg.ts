import { exec as exec_ } from 'child_process'
import * as fs from 'fs'
import * as os from 'os'
import path from 'path'
import { promisify } from 'util'

const exec = promisify(exec_)

export async function getRgPath(extensionPath: string): Promise<string> {
    if (process.env.MOCK_RG_PATH) {
        return process.env.MOCK_RG_PATH
    }

    try {
        const target = await getTarget()
        const resourcesDir = path.join(extensionPath, 'resources', 'bin')
        const files = await new Promise<string[]>((resolve, reject) => {
            fs.readdir(resourcesDir, (err, files) => {
                if (err) {
                    reject(err)
                    return
                }
                resolve(files)
            })
        })
        for (const file of files) {
            if (file.includes(target)) {
                return path.join(resourcesDir, file)
            }
        }
        console.log(
            'Did not find bundled `rg` (if running in development, you probably need to run scripts/download-rg.sh). Falling back to the `rg` on $PATH.'
        )
        return 'rg'
    } catch (error) {
        console.error(error)
        return 'rg'
    }
}

// Code below this line copied from https://github.com/microsoft/vscode-ripgrep

async function isMusl(): Promise<boolean> {
    let stderr
    try {
        stderr = (await exec('ldd --version')).stderr
    } catch (error) {
        stderr = error.stderr
    }
    if (stderr.indexOf('musl') > -1) {
        return true
    }
    return false
}

async function getTarget(): Promise<string> {
    const arch = process.env.npm_config_arch || os.arch()

    switch (os.platform()) {
        case 'darwin':
            return arch === 'arm64' ? 'aarch64-apple-darwin' : 'x86_64-apple-darwin'
        case 'win32':
            return arch === 'x64'
                ? 'x86_64-pc-windows-msvc'
                : arch === 'arm'
                ? 'aarch64-pc-windows-msvc'
                : 'i686-pc-windows-msvc'
        case 'linux':
            return arch === 'x64'
                ? 'x86_64-unknown-linux-musl'
                : arch === 'arm'
                ? 'arm-unknown-linux-gnueabihf'
                : arch === 'armv7l'
                ? 'arm-unknown-linux-gnueabihf'
                : arch === 'arm64'
                ? (await isMusl())
                    ? 'aarch64-unknown-linux-musl'
                    : 'aarch64-unknown-linux-gnu'
                : arch === 'ppc64'
                ? 'powerpc64le-unknown-linux-gnu'
                : arch === 's390x'
                ? 's390x-unknown-linux-gnu'
                : 'i686-unknown-linux-musl'
        default:
            throw new Error('Unknown platform: ' + os.platform())
    }
}
