import { Client } from '../client'
import { WindowShowMessageFeature } from './message'

const create = (): {
    client: Client
    feature: WindowShowMessageFeature
} => {
    const client: Client = { onNotification: () => void 0, onRequest: () => void 0 } as any
    const feature = new WindowShowMessageFeature(client, () => void 0, async () => null)
    return { client, feature }
}

describe('WindowShowMessageFeature', () => {
    it('initializes', () => {
        const { feature } = create()
        feature.initialize()
    })
})
