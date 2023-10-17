import React, { useEffect, useMemo } from 'react'

import type * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ReloadIcon from 'mdi-react/ReloadIcon'
import { isRouteErrorResponse, useRouteError } from 'react-router-dom'

import { asError, logger } from '@sourcegraph/common'
import { Button, Code, Text } from '@sourcegraph/wildcard'

import { isChunkLoadError } from '../monitoring'

import { HeroPage } from './HeroPage'

interface Props {
    /**
     * The current location, or null if there is no location (such as the root component, which is above the
     * react-router component).
     */
    location: H.Location | null

    /**
     * Extra context to aid with debugging
     */
    extraContext?: JSX.Element

    /**
     * Custom render logic in place of <HeroPage>
     */
    render?: (error: Error) => JSX.Element

    /**
     * Classname to pass to <HeroPage>
     */
    className?: string
}

interface State {
    error?: Error
}

/**
 * A [React error boundary](https://reactjs.org/docs/error-boundaries.html) that catches errors from
 * its children. If an error occurs, it displays a nice error page instead of a blank page and reports the error to Sentry.
 *
 * Components should handle their own errors (and must not rely on this error boundary). This error
 * boundary is a last resort in case of an unexpected error.
 */
export class ErrorBoundary extends React.PureComponent<React.PropsWithChildren<Props>, State> {
    public state: State = {}

    public static getDerivedStateFromError(error: any): Pick<State, 'error'> {
        return { error: asError(error) }
    }

    public componentDidCatch(error: unknown, errorInfo: React.ErrorInfo): void {
        if (typeof Sentry !== 'undefined') {
            Sentry.withScope(scope => {
                for (const [key, value] of Object.entries(errorInfo)) {
                    scope.setExtra(key, value)
                }
                Sentry.captureException(error)
            })
        }
    }

    public componentDidUpdate(previousProps: Props): void {
        if (previousProps.location !== this.props.location) {
            // Reset error state when location changes, so that the user can try navigating to a different page to
            // clear the error.
            /* eslint react/no-did-update-set-state: warn */
            this.setState({ error: undefined })
        }
    }

    public render(): React.ReactNode | null {
        if (this.state.error !== undefined) {
            return (
                <ErrorBoundaryMessage
                    error={this.state.error}
                    extraContext={this.props.extraContext}
                    render={this.props.render}
                    className={this.props.className}
                />
            )
        }

        return this.props.children
    }
}

/**
 * A React component that can be used within a react router errorElement callback. It extracts the
 * route error and displays it nicely on the page.
 */
export const RouteError: React.FC = () => {
    const routeError = useRouteError()
    const error = useMemo(() => {
        if (isRouteErrorResponse(routeError)) {
            return new Error(routeError.data ? routeError.data : `${routeError.status} ${routeError.statusText}`)
        }
        return asError(routeError)
    }, [routeError])
    useEffect(() => {
        logger.error(error)
        if (typeof Sentry !== 'undefined') {
            Sentry.captureException(error)
        }
    }, [error])
    return <ErrorBoundaryMessage error={error} />
}

interface ErrorBoundaryMessageProps {
    error: Error
    // Extra context to aid with debugging
    extraContext?: JSX.Element
    // Custom render logic in place of <HeroPage>
    render?: (error: Error) => JSX.Element
    // className to pass to <HeroPage>
    className?: string
}
const ErrorBoundaryMessage: React.FC<ErrorBoundaryMessageProps> = ({ error, extraContext, render, className }) => {
    if (isChunkLoadError(error)) {
        // This means that the JavaScript assets that correspond to the deploy version currently
        // running are no longer available, likely because a redeploy occurred after the user
        // initially loaded this page.
        return (
            <HeroPage
                icon={ReloadIcon}
                title="Reload required"
                subtitle={
                    <div className="container">
                        <Text>A new version of Sourcegraph is available.</Text>
                        <Button onClick={hardReload} variant="primary">
                            Reload to update
                        </Button>
                    </div>
                }
            />
        )
    }

    if (render) {
        return render(error)
    }

    return (
        <HeroPage
            icon={AlertCircleIcon}
            title="Error"
            className={className}
            subtitle={
                <div className="container">
                    <Text>
                        Sourcegraph encountered an unexpected error. If reloading the page doesn't fix it, contact your
                        site admin or Sourcegraph support.
                    </Text>
                    <Text>
                        <Code className="text-wrap">{error.message}</Code>
                    </Text>
                    {extraContext}
                </div>
            }
        />
    )
}

function hardReload(): void {
    window.location.reload()
}
