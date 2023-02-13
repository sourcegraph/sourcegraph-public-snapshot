import { FC } from 'react'

interface CodeHostEditProps {}

/**
 * Renders edit UI for any supported code host type. (Github, Gitlab, ...)
 * Also performs edit, delete actions over opened code host connection
 */
export const CodeHostEdit: FC<CodeHostEditProps> = () => {
    return <>Edit UI</>
}

const CodeHostEditView: FC = () => {
    return <>Specific code host edit UI</>
}
