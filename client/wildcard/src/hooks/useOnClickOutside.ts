import { type RefObject, useEffect } from 'react'

export function useOnClickOutside(reference: RefObject<HTMLElement>, handler: (event: Event) => void): void {
    useEffect(() => {
        const listener = (event: Event): void => {
            // Do nothing if clicking ref's element or descendent elements
            if (!reference.current || reference.current.contains(event.target as Node)) {
                return
            }

            handler(event)
        }

        document.addEventListener('mouseup', listener)
        document.addEventListener('touchstart', listener)

        return () => {
            document.removeEventListener('mouseup', listener)
            document.removeEventListener('touchstart', listener)
        }
    }, [reference, handler])
}
