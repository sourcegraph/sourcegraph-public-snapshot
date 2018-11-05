import * as React from 'react'
import { Button } from '../../../shared/components/Button'

const info =
    'This repository does not have a staging area enabled and we were unable to apply the patch set for this diff.'

export interface StagingAreaInformationProps {
    className?: string
    style?: React.CSSProperties
    iconStyle?: React.CSSProperties
}

export const StagingAreaInformation: React.StatelessComponent<StagingAreaInformationProps> = (
    props: StagingAreaInformationProps
) => (
    <div style={{ display: 'inline-block' }}>
        <Button
            url="https://about.sourcegraph.com/docs/features/phabricator-extension#staging-areas"
            style={props.style}
            iconStyle={props.iconStyle}
            className={props.className}
            ariaLabel={info}
            label="Unable to resolve diff"
        />
    </div>
)
