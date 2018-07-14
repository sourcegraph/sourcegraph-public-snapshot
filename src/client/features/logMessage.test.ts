import { Client } from '../client'
import { WindowLogMessageFeature } from './logMessage'

const create = (): {
    client: Client
    feature: WindowLogMessageFeature
} => {
    const client: Client = { onNotification: () => void 0 } as any
    const feature = new WindowLogMessageFeature(client, () => void 0)
    return { client, feature }
}

describe('WindowLogMessageFeature', () => {
    it('initializes', () => {
        const { feature } = create()
        feature.initialize()
    })
})
