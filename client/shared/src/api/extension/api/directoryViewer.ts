import { type ProxyMarked, proxyMarker } from 'comlink'

import type { Directory } from '../../../codeintel/legacy-extensions/api'
import type { DirectoryViewerData, ViewerId } from '../../viewerTypes'

export class ExtensionDirectoryViewer implements ProxyMarked {
    public readonly [proxyMarker] = true
    public readonly viewerId: string
    public readonly type = 'DirectoryViewer'
    public isActive: boolean
    public directory: Directory
    public resource: string
    constructor(data: DirectoryViewerData & ViewerId) {
        this.isActive = data.isActive
        this.viewerId = data.viewerId
        this.resource = data.resource
        // Since directories don't have any state beyond the immutable URI,
        // we can set the model to a static object for now and don't need to track directory models in a Map.
        this.directory = {
            uri: new URL(data.resource),
        }
    }
}
