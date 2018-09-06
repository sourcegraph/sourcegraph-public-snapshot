import { RouteComponentProps } from 'react-router'

/**
 * Configuration for a route.
 *
 * @template C Context information that is passed to `render` and `condition`
 */
export interface RouteConfiguration<C extends object = {}> {
    /** Path of this route (appended to the current match) */
    path: string
    exact?: boolean
    render: ((context: C & RouteComponentProps<any>) => React.ReactNode)
    /** Whether this route should be handled */
    condition?: (context: C) => boolean
}
