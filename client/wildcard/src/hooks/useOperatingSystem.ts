import { useEffect, useState } from 'react'

export function useOperatingSystem(): 'Windows' | 'MacOS' | 'Linux' | undefined {
    const [os, setOs] = useState<'Windows' | 'MacOS' | 'Linux'>()

    useEffect(() => {
        if (navigator.userAgent.indexOf('Win') !== -1) {
            setOs('Windows')
        } else if (navigator.userAgent.indexOf('Mac') !== -1) {
            setOs('MacOS')
        } else if (navigator.userAgent.indexOf('Linux') !== -1) {
            setOs('Linux')
        }
    }, [])

    return os
}
