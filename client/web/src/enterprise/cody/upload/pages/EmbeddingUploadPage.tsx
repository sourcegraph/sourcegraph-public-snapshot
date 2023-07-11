import { FC, useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container, ErrorAlert, Link, PageHeader } from '@sourcegraph/wildcard'

import { CodyColorIcon } from '../../../../cody/chat/CodyPageIcon'
import { PageTitle } from '../../../../components/PageTitle'

export interface EmbeddingUploadPageProps extends TelemetryProps {}

export const EmbeddingUploadPage: FC<EmbeddingUploadPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('EmbeddingUploadPage')
    }, [telemetryService])
    return (
        <>
            <PageTitle title="Embedding Third-party Data Upload" />
            <PageHeader
                headingElement="h2"
                path={[
                    { icon: CodyColorIcon, text: 'Cody' },
                    {
                        text: 'Upload',
                    },
                ]}
                description={
                    <>
                        Rules that control keeping embeddings up-to-date. See the{' '}
                        <Link target="_blank" to="/help/cody/explanations/policies">
                            documentation
                        </Link>{' '}
                        for more details.
                    </>
                }
                className="mb-3"
            />
        </>
    )
}
