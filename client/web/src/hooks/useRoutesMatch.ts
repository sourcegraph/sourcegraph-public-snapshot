import { useEffect, useState } from 'react'
import { matchPath, useLocation } from 'react-router'
import { LayoutRouteProps } from '../routes'

export const useRoutesMatch = (routes: readonly LayoutRouteProps<{}>[]): string | undefined => {
    const location = useLocation()
    const [match, setMatch] = useState<string | undefined>('')

    useEffect(() => {
        const newMatch = routes.find(({ path, exact }) => matchPath(location.pathname, { path, exact }))?.path
        setMatch(newMatch)
    }, [location.pathname, routes])

    return match
}
