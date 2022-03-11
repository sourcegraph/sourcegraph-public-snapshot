import { useMemo } from 'react'

import { isMacPlatform as isMacPlatformFn } from '@sourcegraph/common'

export const useModifierKeyLabel = (): string => {
    const isMacPlatform = useMemo(() => isMacPlatformFn(), [])
    return useMemo(() => (isMacPlatform ? '⌘' : 'Ctrl'), [isMacPlatform])
}
