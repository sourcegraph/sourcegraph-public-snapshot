// tslint:disable:no-string-literal

import chalk from 'chalk'
import errorStackParser from 'error-stack-parser'
import fs from 'fs'
import os from 'os'
import path from 'path'
import * as shelljs from 'shelljs'
import * as shell from 'shelljs-exec-proxy'
import tmp from 'tmp'
import yargs from 'yargs'

export function main(argv: string[]): void {
    chalk.enabled = true
    chalk.level = 1

    const buildType = { choices: ['dev', 'dist'], default: 'dev' }

    // tslint:disable-next-line:no-unused-expression
    yargs
        .command('execute', 'Builds and runs the symbols package.', { buildType }, runExecutable as (
            args: { buildType: string }
        ) => void)
        .command(
            'buildDockerImage',
            'Builds the symbols Docker image.',
            {
                dockerImageName: { type: 'string', demand: true },
                buildType,
            },
            buildDockerImage as (args: { dockerImageName: string; buildType: string }) => void
        )
        .demandCommand(1, 'You have to specify a command.')
        .strict()
        .parse(argv)
}

interface MuslGcc {
    command: string
    installationHelp: string
}

const muslGccByPlatform: { [platform in NodeJS.Platform]?: MuslGcc } = {
    darwin: {
        command: 'x86_64-linux-musl-gcc',
        installationHelp: 'Run `brew install FiloSottile/musl-cross/musl-cross`.',
    },
    linux: {
        command: 'musl-gcc',
        installationHelp: 'Install the musl-tools package (e.g. on Ubuntu, run `apt-get install musl-tools`).',
    },
}

const libsqlite3PcreFilenameByPlatform: { [platform in NodeJS.Platform]?: string } = {
    darwin: 'libsqlite3-pcre.dylib',
    linux: 'libsqlite3-pcre.so',
}

const repositoryRoot: string = process.cwd()

/**
 * Runs the command silently and checks that the exit code is 0.
 */
function testSilently(...command: string[]): boolean {
    const oldSilent = shell.config.silent
    shell.config.silent = true
    const code = shell[command[0]](...command.slice(1)).code
    shell.config.silent = oldSilent
    return code === 0
}

/**
 * Logs an error after cleaning up the stack (removing internal frames, trimming $PWD, etc.).
 */
function logError(error: Error): void {
    console.log(chalk.redBright.red(error.message))
    console.log(
        errorStackParser
            .parse(error)
            .filter(({ fileName }) => !/^internal/.test(fileName || '?'))
            .map(frame => ({ ...frame, fileName: frame.fileName!.replace(`${repositoryRoot}/`, '') }))
            .map(({ fileName, lineNumber, columnNumber, functionName }) =>
                /node_modules/.test(fileName!)
                    ? chalk.gray(`    ${fileName}:${lineNumber}:${columnNumber} (in ${functionName || '?'})`)
                    : `    ${fileName}:${lineNumber}:${columnNumber} (in ${chalk.cyan(functionName || '?')})`
            )
            .join('\n')
    )
}

/**
 * Only prints output when the command fails. Exits the process on failure.
 */
function runArrayOrShell(command: string[] | string): shelljs.ExecOutputReturnValue {
    const oldSilent = shell.config.silent
    shell.config.silent = true
    const result = typeof command === 'string' ? shell.exec(command) : shell[command[0]](...command.slice(1))
    shell.config.silent = oldSilent
    if (result.code !== 0) {
        logError(
            new Error(
                `Command exited with code ${result.code}: ${typeof command === 'string' ? command : command.join(' ')}`
            )
        )
        process.stdout.write(result.stdout)
        process.stderr.write(result.stderr)
        process.exit(result.code)
    }
    return result
}

/**
 * Only prints output when the command fails. Exits the process on failure.
 */
function run(...command: string[]): shelljs.ExecOutputReturnValue {
    return runArrayOrShell(command)
}

/**
 * Only prints output when the command fails. Exits the process on failure.
 */
function runShell(command: string): shelljs.ExecOutputReturnValue {
    return runArrayOrShell(command)
}

type BuildType = 'dev' | 'dist'

/**
 * Builds the symbols executable.
 */
function buildExecutable({ outputPath, buildType }: { outputPath: string; buildType: BuildType }): void {
    const symbolsPackage = 'github.com/sourcegraph/sourcegraph/cmd/symbols'
    const gcFlagsByBuildType: { [buildType in BuildType]: string } = {
        dev: 'all=-N -l',
        dist: '',
    }
    const tagsByBuildType: { [buildType in BuildType]: string } = {
        dev: 'dev delve',
        dist: 'dist',
    }
    const gcFlags = gcFlagsByBuildType[buildType]
    const tags = tagsByBuildType[buildType]
    console.log('Building the symbols executable...')
    run('go', 'build', '-buildmode', 'exe', '-gcflags', gcFlags, '-tags', tags, '-o', outputPath, symbolsPackage)
    console.log('Building the symbols executable... done')
}

/**
 * Builds the PCRE extension to sqlite3.
 */
function buildLibsqlite3Pcre({ outputPath }: { outputPath: string }): void {
    const sqlite3PcreRepositoryDirectory = tmp.dirSync().name

    run('git', 'clone', 'https://github.com/ralight/sqlite3-pcre', sqlite3PcreRepositoryDirectory)
    shelljs.pushd('-q', sqlite3PcreRepositoryDirectory)
    runShell(
        `gcc -fno-common -dynamiclib pcre.c -o ${outputPath} $(pkg-config --cflags sqlite3 libpcre) $(pkg-config --libs libpcre)`
    )
    shelljs.popd('-q')
}

/**
 * Builds and runs the symbols executable.
 */
function runExecutable({ buildType }: { buildType: BuildType }): void {
    const libsqlite3PcrePath = path.join(
        repositoryRoot,
        libsqlite3PcreFilenameByPlatform[os.platform()] || 'libsqlite3-pcre.so'
    )
    if (!fs.existsSync(libsqlite3PcrePath)) {
        buildLibsqlite3Pcre({ outputPath: libsqlite3PcrePath })
    }
    const outputPath = tmp.tmpNameSync({ prefix: 'symbols' })
    buildExecutable({ outputPath, buildType })
    // TODO uncomment
    // buildDockerImage({ dockerImageName: 'dev-symbols', buildType: 'dev' })
    shell.env['LIBSQLITE3_REGEXP'] = libsqlite3PcrePath
    shell.env['CTAGS_COMMAND'] = shell.env['CTAGS_COMMAND'] || 'cmd/symbols/universal-ctags-dev'
    shell.env['CTAGS_PROCESSES'] = shell.env['CTAGS_PROCESSES'] || '1'
    shell.exec(outputPath)
}

/**
 * Builds the symbols Docker image.
 */
function buildDockerImage({ dockerImageName, buildType }: { dockerImageName: string; buildType: BuildType }): void {
    const muslGcc = muslGccByPlatform[os.platform()]
    if (!muslGcc) {
        console.log(`Unsupported OS platform ${os.platform()}`)
        process.exit(1)
        return
    }

    if (!testSilently('command', '-v', muslGcc.command)) {
        console.log(`Couldn't find musl C compiler (needed by sqlite3) ${muslGcc.command}. ${muslGcc.installationHelp}`)
        process.exit(1)
    }

    const dockerBuildContext = tmp.dirSync().name

    shell.env['CC'] = muslGcc.command
    shell.env['GO111MODULE'] = 'on'
    shell.env['GOARCH'] = 'amd64'
    shell.env['GOOS'] = 'linux'
    shell.env['CGO_ENABLED'] = '1' // to build the sqlite3 library
    const symbolsOut = path.join(dockerBuildContext, 'symbols')
    buildExecutable({ outputPath: symbolsOut, buildType })

    shelljs.cp('-R', 'cmd/symbols/.ctags.d', dockerBuildContext)

    console.log(`Building the ${dockerImageName} Docker image...`)
    run('docker', 'build', '--quiet', '-f', 'cmd/symbols/Dockerfile', '-t', dockerImageName, dockerBuildContext)
    console.log(`Building the ${dockerImageName} Docker image... done`)
}
