import { useEffect, useState } from 'react'

import { parseQueryInt } from '../utils'

interface UseVisibleParameters {
    shouldCheck?: boolean
    searchParameters: URLSearchParams
}

export const useVisible = ({
    shouldCheck,
    searchParameters,
}: UseVisibleParameters): [number, (value: number) => void] => {
    const [visible, setVisible] = useState(0)

    useEffect(() => {
        const visible = shouldCheck && parseQueryInt(searchParameters, 'visible')

        if (visible) {
            setVisible(visible)
        }
    }, [shouldCheck, searchParameters])

    return [visible, setVisible]
}
