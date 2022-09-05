import { useEffect, useState } from 'react'

export function isBlockInputFocused(id: string): boolean {
    if (!document.activeElement) {
        return false
    }
    const activeElement = document.activeElement as HTMLElement
    if (!activeElement.closest(`[data-block-id="${id}"]`)) {
        return false
    }
    const activeTagName = activeElement.tagName.toLowerCase()
    return activeTagName === 'input' || activeTagName === 'textarea' || activeElement.contentEditable === 'true'
}

export function useIsBlockInputFocused(id: string): boolean {
    const [isInputFocused, setIsInputFocused] = useState(false)

    useEffect(() => {
        const handleFocusChange = (): void => {
            setIsInputFocused(isBlockInputFocused(id))
        }
        document.addEventListener('focusin', handleFocusChange)
        document.addEventListener('focusout', handleFocusChange)
        return () => {
            document.removeEventListener('focusin', handleFocusChange)
            document.removeEventListener('focusout', handleFocusChange)
        }
    }, [id])

    return isInputFocused
}
