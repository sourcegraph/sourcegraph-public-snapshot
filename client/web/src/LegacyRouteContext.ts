import { createContext, useContext } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BatchChangesProps } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { BreadcrumbSetters, BreadcrumbsProps } from './components/Breadcrumbs'
import type { LegacyLayoutProps } from './LegacyLayout'
import { ThemePreferenceProps } from './theme'

export interface LegacyLayoutRouteComponentProps
    extends Omit<LegacyLayoutProps, 'match'>,
        ThemeProps,
        ThemePreferenceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        CodeIntelligenceProps,
        BatchChangesProps {
    isSourcegraphDotCom: boolean
    isMacPlatform: boolean
}

interface Props {
    render: (props: LegacyLayoutRouteComponentProps) => JSX.Element
    condition?: (props: LegacyLayoutRouteComponentProps) => boolean
}

/**
 * A wrapper component for React router route entrypoints that still need access to the legacy
 * route context and prop drilling.
 */
export const LegacyRoute = ({ render, condition }: Props): JSX.Element | null => {
    const context = useContext(LegacyRouteContext)
    if (!context) {
        throw new Error('LegacyRoute must be used inside a LegacyRouteContext.Provider')
    }

    if (condition && !condition(context)) {
        return null
    }

    return render(context)
}

export const LegacyRouteContext = createContext<LegacyLayoutRouteComponentProps | null>(null)
