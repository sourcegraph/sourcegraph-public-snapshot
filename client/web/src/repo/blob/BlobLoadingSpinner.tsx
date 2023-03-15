import { useEffect, useState } from 'react'

import { useLocation } from 'react-router-dom'

import { Link, LoadingSpinner } from '@sourcegraph/wildcard'

export function BlobLoadingSpinner(): JSX.Element {
    const [afterOneSec, setAfterOneSec] = useState(false)
    const [afterThreeSec, setAfterThreeSec] = useState(false)
    const [afterSixSec, setAfterSixSec] = useState(false)
    const [afterNineSec, setAfterNineSec] = useState(false)

    useEffect(() => {
        const timeout = setTimeout(() => setAfterOneSec(true), 1000)
        return () => clearTimeout(timeout)
    }, [])
    useEffect(() => {
        const timeout = setTimeout(() => setAfterThreeSec(true), 3000)
        return () => clearTimeout(timeout)
    }, [])
    useEffect(() => {
        const timeout = setTimeout(() => setAfterSixSec(true), 6000)
        return () => clearTimeout(timeout)
    }, [])
    useEffect(() => {
        const timeout = setTimeout(() => setAfterNineSec(true), 9000)
        return () => clearTimeout(timeout)
    }, [])

    const location = useLocation()

    return (
        <div className="d-flex mt-3 justify-content-center">
            {afterOneSec ? (
                <div className="d-flex flex-column align-items-center">
                    <LoadingSpinner />
                    <div className="text-muted mt-2">
                        {afterNineSec ? (
                            <>
                                It’s taking much longer than expected to load this file. Try{' '}
                                <Link to={location.pathname} onClick={reload}>
                                    reloading the page
                                </Link>
                                .
                            </>
                        ) : afterSixSec ? (
                            'Loading a whole lot of bits and bytes here...'
                        ) : afterThreeSec ? (
                            'Loading a large file...'
                        ) : (
                            ''
                        )}
                    </div>
                </div>
            ) : null}
        </div>
    )
}

function reload(event: React.MouseEvent<HTMLAnchorElement>): void {
    window.location.reload()
    event.preventDefault()
}
