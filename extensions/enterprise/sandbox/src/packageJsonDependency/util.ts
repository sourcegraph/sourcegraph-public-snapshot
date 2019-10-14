/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { openTextDocument } from '../dependencyManagement/util'
import semver from 'semver'
import { FormattingOptions, Segment } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import * as sourcegraph from 'sourcegraph'
import { Observable, combineLatest } from 'rxjs'
import { map } from 'rxjs/operators'
import { DependencySpecification, DependencyDeclaration, DependencyResolution } from '../dependencyManagement'
import { PackageJsonDependencyQuery } from '.'
import { LogicalTree } from './npm/logicalTree'

// function computeDiffs(files: { old: sourcegraph.TextDocument; newText?: string }[]): sourcegraph.WorkspaceEdit {
//     const edit = new sourcegraph.WorkspaceEdit()
//     for (const { old, newText } of files) {
//         // TODO!(sqs): handle creation/removal
//         if (old.text !== undefined && newText !== undefined && old.text !== newText) {
//             edit.replace(
//                 new URL(old.uri),
//                 new sourcegraph.Range(new sourcegraph.Position(0, 0), old.positionAt(old.text.length)),
//                 newText
//             )
//         }
//     }
//     return edit
// }

const PACKAGE_JSON_FORMATTING_OPTIONS: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

export const editPackageJson = (
    doc: sourcegraph.TextDocument,
    operations: { path: Segment[]; value: any }[],
    workspaceEdit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit => {
    for (const op of operations) {
        const edits = setProperty(doc.text!, op.path, op.value, PACKAGE_JSON_FORMATTING_OPTIONS)
        for (const edit of edits) {
            workspaceEdit.replace(
                new URL(doc.uri),
                new sourcegraph.Range(doc.positionAt(edit.offset), doc.positionAt(edit.offset + edit.length)),
                edit.content
            )
        }
    }
    return workspaceEdit
}

export const traversePackageJsonLockfile = (
    tree: LogicalTree,
    parsedQuery: PackageJsonDependencyQuery
): Pick<DependencySpecification<PackageJsonDependencyQuery>, 'declarations' | 'resolutions'> => {
    // TODO!(sqs): make this identify all deps that need changing
    const declarations: DependencyDeclaration[] = []
    const resolutions: DependencyResolution[] = []
    // eslint-disable-next-line ban/ban
    tree.forEach((dep, next) => {
        if (dep.name === parsedQuery.name && semver.satisfies(dep.version, parsedQuery.parsedVersionRange)) {
            declarations.push({
                name: dep.name,
                requestedVersion: dep.version,
                direct: dep.requiredBy.has(tree),
                location: 'TODO!(sqs)' as any,
            })
            resolutions.push({ name: dep.name, resolvedVersion: dep.version })
        } else {
            // eslint-disable-next-line callback-return
            next()
        }
    })
    return { declarations, resolutions }
}

export const provideDependencySpecification = (
    packageJson: URL,
    lockfile: URL,
    query: PackageJsonDependencyQuery,
    getDependencyFromPackageJsonAndLockfile: (
        packageJson: string,
        lockfile: string,
        query: PackageJsonDependencyQuery
    ) => Pick<DependencySpecification<PackageJsonDependencyQuery>, 'declarations' | 'resolutions'>
): Observable<DependencySpecification<PackageJsonDependencyQuery> | null> =>
    combineLatest([openTextDocument(packageJson), openTextDocument(lockfile)]).pipe(
        map(([packageJson, lockfile]) => {
            if (packageJson === null || lockfile === null) {
                return null
            }
            const partialSpec: Pick<DependencySpecification<PackageJsonDependencyQuery>, 'query'> = {
                query,
            }
            try {
                // TODO!(sqs): support multiple versions in lockfile/package.json
                const { declarations, resolutions } = getDependencyFromPackageJsonAndLockfile(
                    packageJson.text!,
                    lockfile.text!,
                    query
                )
                if (declarations.length === 0 && resolutions.length === 0) {
                    return null
                }
                const spec: DependencySpecification<PackageJsonDependencyQuery> = {
                    ...partialSpec,
                    declarations: declarations.map(decl => ({
                        ...decl,
                        location: {
                            uri: new URL(packageJson.uri),
                            // TODO!(sqs): this is not exact anyway, needs to traverse json file
                            range: findMatchRange(packageJson.text!, `"${query.name}"`),
                        },
                    })),
                    resolutions: resolutions.map(res => ({
                        ...res,
                        location: {
                            uri: new URL(lockfile.uri),
                            // TODO!(sqs): this differs from yarn.lock vs package-lock.json and is not exact anyway, needs to traverse json file
                            range: findMatchRange(packageJson.text!, query.name),
                        },
                    })),
                }
                return spec
            } catch (err) {
                const spec: DependencySpecification<PackageJsonDependencyQuery> = {
                    ...partialSpec,
                    declarations: [],
                    resolutions: [],
                    error: Object.assign(
                        new Error(
                            `Unable to get dependency specification from package.json and lockfile (package ${JSON.stringify(
                                query.name
                            )} in ${packageJson.uri}): ${err.message}`
                        ),
                        { data: { packageJson: packageJson.uri, query }, wrapped: err }
                    ),
                }
                return spec
            }
        })
    )

function findMatchRange(text: string, str: string): sourcegraph.Range | undefined {
    for (const [i, line] of text.split('\n').entries()) {
        const j = line.indexOf(str)
        if (j !== -1) {
            return new sourcegraph.Range(i, j, i, j + str.length)
        }
    }
    return undefined
}
