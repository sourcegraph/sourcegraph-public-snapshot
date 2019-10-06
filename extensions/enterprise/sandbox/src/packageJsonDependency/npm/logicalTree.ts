/* eslint-disable @typescript-eslint/explicit-function-return-type */
// from https://raw.githubusercontent.com/npm/logical-tree/latest/index.js
'use strict'

import path from 'path'

export interface Dependency extends Node, NodeInfo {}

export interface Node {
    name: string
    version: string
    address: string | null
    dependencies: Map<string, Node>
    requiredBy: Set<Node>
    isRoot: boolean
}

interface NodeInfo {
    version: string
    optional?: boolean
    dev?: boolean
    bundled?: boolean
    resolved: string
    integrity: boolean
}

export class LogicalTree implements Node, NodeInfo {
    public version: string
    public optional?: boolean
    public dev?: boolean
    public bundled?: boolean
    public resolved: string
    public integrity: boolean
    public dependencies = new Map<string, Node>()
    public requiredBy = new Set<Node>()

    constructor(public readonly name: string, public readonly address: string | null, info: NodeInfo) {
        this.name = name
        this.version = info.version
        this.optional = !!info.optional
        this.dev = !!info.dev
        this.bundled = !!info.bundled
        this.resolved = info.resolved
        this.integrity = info.integrity
    }

    public get isRoot(): boolean {
        return !this.requiredBy.size
    }

    public addDep(dep: Node) {
        this.dependencies.set(dep.name, dep)
        dep.requiredBy.add(this)
        return this
    }

    public delDep(dep: Node) {
        this.dependencies.delete(dep.name)
        dep.requiredBy.delete(this)
        return this
    }

    public getDep(name: string) {
        return this.dependencies.get(name)
    }

    public path(prefix: string) {
        if (this.isRoot) {
            // The address of the root is the prefix itself.
            return prefix || ''
        }
        return path.join(prefix || '', 'node_modules', this.address!.replace(/:/g, '/node_modules/'))
    }

    // This finds cycles _from_ a given node: if some deeper dep has
    // its own cycle, but that cycle does not refer to this node,
    // it will return false.
    public hasCycle(_seen?: any, _from?: any) {
        if (!_seen) {
            _seen = new Set()
        }
        if (!_from) {
            _from = this
        }
        for (const dep of this.dependencies.values()) {
            if (_seen.has(dep)) {
                continue
            }
            _seen.add(dep)
            if (dep === _from || dep.hasCycle(_seen, _from)) {
                return true
            }
        }
        return false
    }

    public forEach(fn: (dep: Dependency, next: () => void) => void, _seen = new Set<LogicalTree>()) {
        if (_seen.has(this)) {
            return
        }
        _seen.add(this)
        fn(this, () => {
            for (const dep of this.dependencies.values()) {
                dep.forEach(fn, _seen)
            }
        })
    }
}

export function lockTree(pkg: any, pkgLock: any) {
    const tree = makeNode(pkg.name, null, pkg)
    const allDeps = new Map()
    Array.from(
        new Set(
            Object.keys(pkg.devDependencies || {})
                .concat(Object.keys(pkg.optionalDependencies || {}))
                .concat(Object.keys(pkg.dependencies || {}))
        )
    ).forEach(name => {
        let dep = allDeps.get(name)
        if (!dep) {
            const depNode = (pkgLock.dependencies || {})[name]
            dep = makeNode(name, name, depNode)
        }
        addChild(dep, tree, allDeps, pkgLock)
    })
    return tree
}

export function makeNode(name: any, address: any, opts?: any) {
    return new LogicalTree(name, address, opts || {})
}

function addChild(dep: any, tree: any, allDeps: any, pkgLock: any) {
    tree.addDep(dep)
    allDeps.set(dep.address, dep)
    const addr = dep.address
    const lockNode = atAddr(pkgLock, addr)
    if (!lockNode) {
        // TODO!(sqs): this seems to happen for nested package-lock.json files
        //
        // console.log('NO LOCK NODE', { pkgLock, addr, dep })
        return
    }
    Object.keys(lockNode.requires || {}).forEach(name => {
        const tdepAddr = reqAddr(pkgLock, name, addr)
        let tdep = allDeps.get(tdepAddr)
        if (!tdep) {
            tdep = makeNode(name, tdepAddr, atAddr(pkgLock, tdepAddr))
            addChild(tdep, dep, allDeps, pkgLock)
        } else {
            dep.addDep(tdep)
        }
    })
}

function reqAddr(pkgLock: any, name: any, fromAddr: any) {
    const lockNode = atAddr(pkgLock, fromAddr)
    const child = (lockNode.dependencies || {})[name]
    if (child) {
        return `${fromAddr}:${name}`
    }
    const parts = fromAddr.split(':')
    while (parts.length) {
        parts.pop()
        const joined = parts.join(':')
        const parent = atAddr(pkgLock, joined)
        if (parent) {
            const child = (parent.dependencies || {})[name]
            if (child) {
                return `${joined}${parts.length ? ':' : ''}${name}`
            }
        }
    }
    const err: Error & { pkgLock?: any; target?: any; from?: any } = new Error(
        `${name} not accessible from ${fromAddr}`
    )
    err.pkgLock = pkgLock
    err.target = name
    err.from = fromAddr
    throw err
}

function atAddr(pkgLock: any, addr: string) {
    if (!addr.length) {
        return pkgLock
    }
    const parts = addr.split(':')
    return parts.reduce((acc, next) => acc && (acc.dependencies || {})[next], pkgLock)
}
