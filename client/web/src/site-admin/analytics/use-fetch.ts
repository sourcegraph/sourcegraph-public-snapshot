import { useEffect, useState } from 'react'

export function useFetch<R>(fetch: () => Promise<R>): [R | undefined, boolean, Error | undefined] {
    const [data, setData] = useState<R>()
    const [error, setError] = useState<Error>()
    const [isLoading, setIsLoading] = useState(true)
    useEffect(() => {
        setIsLoading(true)
        setError(undefined)
        fetch()
            .then(data => {
                setData(data)
                setIsLoading(false)
            })
            .catch(error => {
                setError(error)
                setIsLoading(false)
            })
    }, [fetch])
    return [data, isLoading, error]
}
