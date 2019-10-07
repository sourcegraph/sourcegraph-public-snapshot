'use strict'

import semver from 'semver'
import { makeNode } from '../npm/logicalTree'
import { parse as parseLockfile } from '@yarnpkg/lockfile'

export function yarnLogicalTree(pkg, yarnLock) {
    yarnLock = parseLockfile(yarnLock).object
    const tree = makeNode(pkg.name, null, pkg)
    const allDeps = new Map()
    const pkgDeps = Object.assign({}, pkg.devDependencies || {}, pkg.optionalDependencies || {}, pkg.dependencies || {})
    Object.keys(pkgDeps).forEach(name => {
        const semverString = pkgDeps[name]
        let dep = allDeps.get(`${name}@${semverString}`)
        if (!dep) {
            const semverString = pkgDeps[name]
            const depNode = yarnLock[`${name}@${semverString}`]
            if (!semver.validRange(semverString)) {
                // eg. file, url, etc.
                depNode.version = semverString
            }
            dep = { node: makeNode(name, name, depNode), semverString }
        }
        addChild(dep, { node: tree }, allDeps, yarnLock)
    })
    return tree
}

function addChild(dep, treeNode, allDeps, yarnLock) {
    const tree = treeNode.node
    const { node, semverString } = dep
    const lockNode = yarnLock[`${node.name}@${semverString}`]
    if (!lockNode) {
        return // TODO!(sqs)??? otherwise has an exception
    }
    const dependencies = Object.assign({}, lockNode.optionalDependencies || {}, lockNode.dependencies || {})
    tree.addDep(node)
    allDeps.set(`${node.name}@${semverString}`, dep)
    Object.keys(dependencies).forEach(name => {
        const tdepSemver = dependencies[name]
        let tdep = allDeps.get(`${name}@${tdepSemver}`)
        if (!tdep) {
            const tdepNode = yarnLock[`${name}@${tdepSemver}`]
            if (!tdepNode) {
                throw new Error(`${name} not accessible from ${node.name}`)
            }
            tdepNode.optional = lockNode.optionalDependencies ? lockNode.optionalDependencies[name] : false
            tdep = { node: makeNode(name, name, tdepNode), semverString: tdepSemver }
            addChild(tdep, dep, allDeps, yarnLock)
        } else {
            node.addDep(tdep.node)
        }
    })
}
