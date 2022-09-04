import { useEffect } from 'react'

export const useScript = (src: string, onLoad: () => void, onError: (error: Error) => void) => {
    useEffect(() => {
        const script = document.createElement('script')
        script.src = src
        script.async = true
        script.addEventListener('load', () => onLoad())
        script.addEventListener('error', ({ error }) =>
            onError(typeof error === 'string' ? new Error(error) : new Error())
        )
        document.body.appendChild(script)
        return () => script.remove()
    }, [onError, onLoad, src])
}
