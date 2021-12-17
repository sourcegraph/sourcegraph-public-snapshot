import { escapeRegExp, find, filter } from 'lodash'
import { load, Kind as YAMLKind, YamlMap as YAMLMap, YAMLNode, YAMLSequence, YAMLScalar } from 'yaml-ast-parser'

const isYAMLMap = (node: YAMLNode): node is YAMLMap => node.kind === YAMLKind.MAP
const isYAMLSequence = (node: YAMLNode): node is YAMLSequence => node.kind === YAMLKind.SEQ
const isYAMLScalar = (node: YAMLNode): node is YAMLScalar => node.kind === YAMLKind.SCALAR

/**
 * A successful result from manipulating the raw batch spec YAML from its AST, even if the
 * manipulation was a no-op
 */
interface YAMLManipulationSuccess {
    success: true
    spec: string
}

/**
 * An unsuccessful result from manipulating the raw batch spec YAML from its AST
 */
interface YAMLManipulationFailure {
    success: false
    error: string
    spec: string
}

type YAMLManipulationResult = YAMLManipulationSuccess | YAMLManipulationFailure

/**
 * Searches the given batch spec YAML AST for a valid "repositoriesMatchingQuery"
 * directive and, if found, adds "-repo:<repo>" to the query string in order to exclude
 * the provided repo from the batch spec workspace results
 *
 * @param spec the raw batch spec YAML string
 * @param ast the batch spec YAML loaded as an AST root node, which should be a `YAMLMap`
 * @param repo the name of the repository to exclude from the batch spec
 * "repositoriesMatchingQuery"
 */
const appendExcludeRepoToQuery = (spec: string, ast: YAMLMap, repo: string): YAMLManipulationResult => {
    // Find the `YAMLMapping` node with the key "on"
    const onMapping = find(ast.mappings, mapping => mapping.key.value === 'on')
    // Take the sequence of values for the "on" key
    const onSequence = onMapping?.value

    if (!onSequence || !isYAMLSequence(onSequence)) {
        return { spec, success: false, error: 'Non-sequence value found for "on" key' }
    }

    // From the sequence, look for a `YAMLMap` node that contains a `YAMLMapping` node
    // with the key "repositoriesMatchingQuery"
    const queryMap: YAMLNode | undefined = find(
        onSequence.items,
        item => isYAMLMap(item) && !!find(item.mappings, mapping => mapping.key.value === 'repositoriesMatchingQuery')
    )

    if (!queryMap || !isYAMLMap(queryMap)) {
        // This just means there wasn't a "repositoriesMatchingQuery" node in the "on"
        // sequence, so return the unaltered spec
        return { spec, success: true }
    }

    // Extract the `YAMLMapping` node with the key "repositoriesMatchingQuery"
    const queryMapping = find(queryMap.mappings, mapping => mapping.key.value === 'repositoriesMatchingQuery')
    // Take the value for the "repositoriesMatchingQuery" key -- this should be a
    // `YAMLScalar` for the query search string
    const queryValue = queryMapping?.value

    if (!queryValue || !isYAMLScalar(queryValue)) {
        return { spec, success: false, error: 'Non-scalar value found for "repositoriesMatchingQuery" key' }
    }

    // Insert "-repo:" qualifier at the end of the query string
    // TODO: In the future this can be integrated into the batch spec under its own
    // "excludes" keyword instead.
    // If the value is quoted, we need to move the addition to the string to _within_
    // the string value.
    let slicePosition = queryValue.endPosition
    if (queryValue.doubleQuoted || queryValue.singleQuoted) {
        slicePosition--
    }
    return {
        success: true,
        spec: spec.slice(0, slicePosition) + ` -repo:${escapeRegExp(repo)}` + spec.slice(slicePosition),
    }
}

/**
 * Searches the given batch spec YAML AST for any valid "repository" directives that match
 * the provided repo name (and branch name, if applicable) and, if found, removes the
 * directive
 *
 * @param spec the raw batch spec YAML string
 * @param ast the batch spec YAML loaded as an AST root node, which should be a `YAMLMap`
 * @param repo the name of the repository to omit the "repository" directive for
 * @param branch the name of the repository branch to omit the "repository" directive for
 */
const removeRepoDirective = (spec: string, ast: YAMLMap, repo: string, branch: string): YAMLManipulationResult => {
    // Find the `YAMLMapping` node with the key "on"
    const onMapping = find(ast.mappings, mapping => mapping.key.value === 'on')
    // Take the sequence of values for the "on" key
    const onSequence = onMapping?.value

    if (!onSequence || !isYAMLSequence(onSequence)) {
        return { spec, success: false, error: 'Non-sequence value found for "on" key' }
    }

    // From the sequence, filter to any `YAMLMap` nodes that contain a `YAMLMapping` node
    // with the key "repository" and a `YAMLScalar` value whose value matches the repo
    // name (there may be none, one, or multiple `YAMLMap`s for different branches)
    const repositoryMatchMaps: YAMLMap[] = filter(
        onSequence.items,
        (item): item is YAMLMap =>
            isYAMLMap(item) &&
            !!find(
                item.mappings,
                mapping =>
                    mapping.key.value === 'repository' &&
                    isYAMLScalar(mapping.value) &&
                    // Compare the values case-insensitively
                    mapping.value.value.toLowerCase() === repo.toLowerCase()
            )
    )

    if (repositoryMatchMaps.length === 0) {
        // This just means there wasn't a matching "repository" directive in the "on"
        // sequence, so return the unaltered spec
        return { spec, success: true }
    }

    // If there's only one matching `YAMLMap` node, we can just remove it from the spec
    if (repositoryMatchMaps.length === 1) {
        const repositoryMatchMap = repositoryMatchMaps[0]
        return {
            success: true,
            spec:
                // NOTE: We also need to trim the sequence delimiter, which is not
                // included in the `YAMLMap`'s start position to end position range
                trimLastSequenceItemDelimiter(spec.slice(0, repositoryMatchMap.startPosition)) +
                spec.slice(repositoryMatchMap.endPosition),
        }
    }

    // Otherwise, if there are multiple matches, look for one that contains a
    // `YAMLMapping` node with the key "branch" and a `YAMLScalar` value whose value
    // matches the branch argument name
    const branchMatchMap: YAMLMap | undefined = find(
        repositoryMatchMaps,
        map =>
            !!find(
                map.mappings,
                mapping =>
                    mapping.key.value === 'branch' &&
                    isYAMLScalar(mapping.value) &&
                    // Compare the values case-insensitively
                    mapping.value.value.toLowerCase() === branch.toLowerCase()
            )
    )

    // If we found no branch match
    if (!branchMatchMap) {
        // This just means none of the matching "repository" directives also matched in
        // the "branch" specified, so return the unaltered spec
        return { spec, success: true }
    }

    // Otherwise, remove the matching `YAMLMap` node from the spec
    return {
        success: true,
        spec:
            // NOTE: We also need to trim the sequence delimiter, which is not included in
            // the `YAMLMap`'s start position to end position range
            trimLastSequenceItemDelimiter(spec.slice(0, branchMatchMap.startPosition)) +
            spec.slice(branchMatchMap.endPosition),
    }
}

/**
 * Trims the final sequence delimiter (i.e. a set of newlines, spaces, and a dash) from
 * the given slice of raw batch spec.
 *
 * This is "sorry-pls-don't-hate-me"-level hack but unfortunately the easiest way around a
 * limitation of working with the YAML AST. The YAML AST parser will not include these
 * characters itself in the character "range" for a node, i.e. they will be present an
 * indeterminate number of characters before `node.startPosition`. So removing a node from
 * a sequence in the spec without also invoking this helper would leave an "empty"
 * sequence item behind and result in parsing errors, like:
 *
 * ```yaml
 * on:
 *   - repository: github.com/sourcegraph/sourcegraph
 *   -
 *   - repository: github.com/sourcegraph/about
 * ```
 *
 * @param specSlice the slice of a raw batch spec YAML string from the beginning up to and
 * including the last set of sequence delimiter characters (one or more newlines followed
 * by zero or more spaces, then a single dash, and then zero or more spaces)
 */
const trimLastSequenceItemDelimiter = (specSlice: string): string =>
    // Trim the last instance of one or more newlines, zero or more spaces, a single dash,
    // and then zero or more spaces, e.g. "\n  - "
    specSlice.replace(/\n+\s*-\s*$/, '')

/**
 * Modifies the provided raw batch spec YAML string in order to exclude a repo resolved in
 * the workspaces preview from the "repositoriesMatchingQuery" value and remove any single
 * "repository" directive that matches the repo name (and branch name, if applicable).
 *
 * @param spec the raw batch spec YAML string
 * @param repo the name of the repository to omit from the batch spec
 * @param branch the name of the repository branch to match when omitting from the batch
 * spec
 */
export const excludeRepo = (spec: string, repo: string, branch: string): YAMLManipulationResult => {
    let ast = load(spec)

    if (!isYAMLMap(ast) || ast.errors.length > 0) {
        return { spec, success: false, error: 'Spec not parseable' }
    }

    // First, try to update the "repositoriesMatchingQuery" string with "-repo:<repo>"
    const appendToQueryResult = appendExcludeRepoToQuery(spec, ast, repo)

    if (!appendToQueryResult.success) {
        return appendToQueryResult
    }

    // Re-parse the AST from the updated result
    ast = load(appendToQueryResult.spec)

    if (!isYAMLMap(ast) || ast.errors.length > 0) {
        return { spec, success: false, error: 'Could not parse spec after updating "repositoriesMatchingQuery"' }
    }

    // Then, also update in case we need to remove any single repository directives that
    // match the repo and branch name
    const removeRepoResult = removeRepoDirective(appendToQueryResult.spec, ast, repo, branch)

    return removeRepoResult
}

/**
 * Checks for a valid "on: " sequence within the provided YAML AST parsed from the input
 * batch spec.
 *
 * @param ast the `YAMLMap` node parsed from the input batch spec
 */
const hasOnStatement = (ast: YAMLMap): boolean => {
    // Find the `YAMLMapping` node with the key "on"
    const onMapping = find(ast.mappings, mapping => mapping.key.value === 'on')
    // Take the sequence of values for the "on" key
    const onSequence = onMapping?.value

    return Boolean(onSequence && isYAMLSequence(onSequence) && onSequence.items.length > 0)
}

/**
 * Checks for a valid "importChangesets: " sequence within the provided YAML AST parsed
 * from the input batch spec.
 *
 * @param ast the `YAMLMap` node parsed from the input batch spec
 */
const hasImportChangesetsStatement = (ast: YAMLMap): boolean => {
    // Find the `YAMLMapping` node with the key "importing changesets"
    const importMapping = find(ast.mappings, mapping => mapping.key.value === 'importChangesets')
    // Take the sequence of values for the "importChangesets" key
    const importSequence = importMapping?.value

    return Boolean(importSequence && isYAMLSequence(importSequence) && importSequence.items.length > 0)
}

/**
 * Checks for a valid "on" or "importChangesets: " sequence within the provided raw batch
 * spec YAML string.
 *
 * @param spec the raw batch spec YAML string
 */
export const hasOnOrImportChangesetsStatement = (spec: string): boolean => {
    const ast = load(spec)

    if (!isYAMLMap(ast) || ast.errors.length > 0) {
        return false
    }

    return hasOnStatement(ast) || hasImportChangesetsStatement(ast)
}

/**
 * Checks whether or not the provided raw batch spec YAML string is a minimal batch spec
 * (i.e. the type auto-created for a brand new draft batch change) or something that the
 * user has touched. If the spec is not parseable, as it might be for an in-progress draft
 * batch spec, this function will return `false`.
 *
 * @param spec the raw batch spec YAML string to check
 */
export const isMinimalBatchSpec = (spec: string): boolean => {
    const ast = load(spec)

    if (!isYAMLMap(ast) || ast.errors.length > 0) {
        return false
    }

    return ast.mappings.length === 1 && ast.mappings[0].key.value === 'name'
}

/**
 * Replaces the "name" value of the provided `librarySpec` with the provided `name`. If
 * `librarySpec` or its "name" is not properly parsable, just returns the original
 * `librarySpec`.
 *
 * @param librarySpec the raw batch spec YAML example code from a library spec
 * @param name the name of the batch change to be inserted
 */
export const insertNameIntoLibraryItem = (librarySpec: string, name: string): string => {
    const ast = load(librarySpec)

    if (!isYAMLMap(ast) || ast.errors.length > 0) {
        return librarySpec
    }

    // Find the `YAMLMapping` node with the key "name"
    const nameMapping = find(ast.mappings, mapping => mapping.key.value === 'name')

    if (!nameMapping || !isYAMLScalar(nameMapping.value)) {
        return librarySpec
    }

    // Stitch the new "name" value into the spec
    return (
        librarySpec.slice(0, nameMapping.value.startPosition) + name + librarySpec.slice(nameMapping.value.endPosition)
    )
}
