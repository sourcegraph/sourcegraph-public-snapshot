import { call } from './api'
import { stage } from './Frame'

export const maintenance = async ({ healthy, onDone }: { healthy: boolean; onDone?: () => void }): Promise<void> => {
    await call('/api/operator/v1beta1/fake/maintenance/healthy', {
        method: 'POST',
        body: JSON.stringify({ healthy: healthy }),
    })
    call('/v1/appliance/status', {
        method: 'POST',
        body: JSON.stringify({ stage: 'maintenance' }),
    }).then(() => {
        if (onDone !== undefined) {
            onDone()
        }
    })
    if (onDone !== undefined) {
        onDone()
    }
}

export const changeStage = ({ action, data, onDone }: { action: stage; data?: string; onDone?: () => void }) => {
    call('/api/v1/appliance/status', {
        method: 'POST',
        body: JSON.stringify({ state: action, data }),
    }).then(() => {
        if (onDone) {
            onDone()
        }
    })
}
