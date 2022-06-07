import { useMemo } from 'react'

import { isMacPlatform as isMacPlatformFunc } from '@sourcegraph/common'

export const useModifierKeyLabel = (): string => {
    const isMacPlatform = useMemo(() => isMacPlatformFunc(), [])
    return useMemo(() => (isMacPlatform ? '⌘' : 'Ctrl'), [isMacPlatform])
}
