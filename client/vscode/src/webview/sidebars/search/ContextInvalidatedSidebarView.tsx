import React from 'react'

import { Button, H5, Text } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../../platform/context'

export interface ContextInvalidatedSidebarViewProps extends WebviewPageProps {}

export const ContextInvalidatedSidebarView: React.FunctionComponent<
    React.PropsWithChildren<ContextInvalidatedSidebarViewProps>
> = ({ extensionCoreAPI }) => (
    <div>
        <H5 className="mt-3 mb-2">Your Sourcegraph instance URL has changed.</H5>
        <Text>Please reload VS Code to use to Sourcegraph extension.</Text>
        <Button
            variant="primary"
            className="font-weight-normal w-100 my-1 border-0"
            onClick={() => extensionCoreAPI.reloadWindow()}
        >
            Reload VS Code
        </Button>
    </div>
)
