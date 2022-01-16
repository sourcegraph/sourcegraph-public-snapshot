import { useEffect } from 'react'

import { useDistinctValue } from '../../../hooks/use-distinct-value'

interface UseKeyboardProps {
    detectKeys: (string | number)[]
    keyevent?: 'keydown' | 'keyup'
}

export function useKeyboard(props: UseKeyboardProps, callback: (event: Event) => void): void {
    const { keyevent = 'keyup', detectKeys } = props
    const keys = useDistinctValue(detectKeys)

    useEffect(() => {
        const handleEvent = (event: KeyboardEvent): void => {
            if (keys.includes(event.key)) {
                return callback(event)
            }
        }

        window.document.addEventListener(keyevent, handleEvent)

        return () => {
            window.document.removeEventListener(keyevent, handleEvent)
        }
    }, [callback, keyevent, keys])
}
