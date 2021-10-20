import { Module } from 'webpack'

const LARGE_DEPENDENCIES = new Set(['monaco-editor'])

export const splitModuleDependenciesByName = (module: Module): string => {
    // get the name. E.g. node_modules/packageName/not/this/part.js or node_modules/packageName
    const packageName = module.context?.match(/[/\\]node_modules[/\\](.*?)([/\\]|$)/)?.[1]

    if (!packageName) {
        // This should never happen. If it does, please raise an issue and notify the Frontend Platform team.
        throw new Error('Could not generate dependency chunk: Module has no name.')
    }

    // We modify the prefix so we can easily differentiate between the few large vs many small dependencies that we have.
    const prefix = LARGE_DEPENDENCIES.has(packageName) ? 'npm-large' : 'npm'

    // npm package names are URL-safe, but some servers don't like @ symbols
    return `${prefix}.${packageName.replace('@', '')}`
}
