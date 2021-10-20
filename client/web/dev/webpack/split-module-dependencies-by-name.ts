import { Module } from 'webpack'

export const splitModuleDependenciesByName = (module: Module): string => {
    // get the name. E.g. node_modules/packageName/not/this/part.js or node_modules/packageName
    const packageName = module.context?.match(/[/\\]node_modules[/\\](.*?)([/\\]|$)/)?.[1]

    if (!packageName) {
        // This should never happen. If it does, please raise an issue and notify the Frontend Platform team.
        throw new Error('Could not generate dependency chunk: Module has no name.')
    }

    // npm package names are URL-safe, but some servers don't like @ symbols
    return `npm.${packageName.replace('@', '')}`
}
