import React, { useEffect, useState } from 'react'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Link } from '@sourcegraph/wildcard'

import { CloudCtaBanner } from '../../components/CloudCtaBanner'
import { eventLogger } from '../../tracking/eventLogger'

interface CloudCtaHomepageProps {
    authenticatedUser?: AuthenticatedUser | null
}

export const CloudCtaHomepage: React.FunctionComponent<CloudCtaHomepageProps> = ({ authenticatedUser }) => {
    const [cloudCtaVariant, setCloudCtaVariant] = useState<CloudCtaBanner['variant'] | string>('filled')
    useEffect(() => {
        const searchParams = new URL(window.location.href).searchParams
        const uxParam = searchParams.get('cta')
        const allowedVariants: { [key: string]: string | undefined } = {
            a: 'filled',
            b: 'underlined',
            c: 'outlined',
            d: undefined,
        }

        if (uxParam && Object.keys(allowedVariants).includes(uxParam)) {
            setCloudCtaVariant(allowedVariants[uxParam])
        }
    }, [])

    return (
        <div className="d-table mx-auto">
            <CloudCtaBanner variant={cloudCtaVariant}>
                To search across your private repositories,{' '}
                <Link
                    to={buildCloudTrialURL(authenticatedUser)}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'HomeUnderSearch' })}
                >
                    try Sourcegraph Cloud
                </Link>
                .
            </CloudCtaBanner>
        </div>
    )
}
