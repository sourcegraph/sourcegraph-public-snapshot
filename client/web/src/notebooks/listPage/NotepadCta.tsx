import React from 'react'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { ProductStatusBadge, Button, H3, Text } from '@sourcegraph/wildcard'

import { NotepadIcon } from '../../search/Notepad'

export const NOTEPAD_CTA_ID = 'notepad-cta'

interface NotepadCTAProps {
    onEnable: () => void
    onCancel: () => void
}

export const NotepadCTA: React.FunctionComponent<React.PropsWithChildren<NotepadCTAProps>> = ({
    onEnable,
    onCancel,
}) => {
    const assetsRoot = window.context?.assetsRoot || ''
    const isLightTheme = useIsLightTheme()

    return (
        <div>
            <H3 id={NOTEPAD_CTA_ID} className="d-inline-block">
                <NotepadIcon /> Enable notepad
            </H3>{' '}
            <ProductStatusBadge status="beta" />
            <div className="d-flex align-items-center">
                <img
                    className="flex-shrink-0 mr-3"
                    src={`${assetsRoot}/img/notepad-illustration-${isLightTheme ? 'light' : 'dark'}.svg`}
                    alt="notepad illustration"
                />
                <Text>
                    The notepad adds a toolbar to the bottom right of search results and file pages to help you create
                    notebooks from your code navigation activities.
                </Text>
            </div>
            <div className="float-right mt-2">
                <Button className="mr-2" variant="secondary" size="sm" onClick={onCancel}>
                    Cancel
                </Button>
                <Button variant="primary" onClick={onEnable} size="sm">
                    Enable notepad
                </Button>
            </div>
        </div>
    )
}
