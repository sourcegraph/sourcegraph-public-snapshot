import type { FC } from 'react'

import { useBeforeUnloadPrompt } from '../../hooks'

interface BeforeUnloadPromptProps {
    when: boolean
    message: string
    beforeUnload?: boolean
}

export const BeforeUnloadPrompt: FC<BeforeUnloadPromptProps> = props => {
    const { when, message, beforeUnload } = props

    useBeforeUnloadPrompt(when ? message : false, { beforeUnload })

    return null
}
