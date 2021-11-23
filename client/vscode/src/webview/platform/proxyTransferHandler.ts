import * as Comlink from 'comlink'
import { isObject } from 'lodash'

import { hasProperty } from '@sourcegraph/shared/src/util/types'

// TODO create comlink folder, add tests w/ Puppeteer. These tests will be useful even if we move to a
// from-scratch node <-> web RPC solution.

export function generateUUID(): string {
    return new Array(4)
        .fill(0)
        .map(() => Math.floor(Math.random() * Number.MAX_SAFE_INTEGER).toString(16))
        .join('-')
}

// TODO wire type (serialized)

export type RelationshipType = 'webToWeb' | 'webToNode'

export interface NestedConnectionData {
    nestedConnectionId: string
    proxyMarkedValue?: object
    panelId: string
    relationshipType?: RelationshipType
}

export function isNestedConnection(value: unknown): value is NestedConnectionData {
    return isObject(value) && hasProperty('nestedConnectionId')(value) && hasProperty('proxyMarkedValue')(value)
}

export function isProxyMarked(value: unknown): value is Comlink.ProxyMarked {
    return isObject(value) && (value as Comlink.ProxyMarked)[Comlink.proxyMarker]
}
