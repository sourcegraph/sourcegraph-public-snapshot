import React, { useMemo } from 'react'

import { mdiCloudDownload } from '@mdi/js'
import { parseISO } from 'date-fns'
import formatDistance from 'date-fns/formatDistance'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable, Link, Alert, Icon, Code, H2, H3, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import { fetchSiteUpdateCheck } from './backend'

import styles from './SiteAdminUpdatesPage.module.scss'

interface Props extends TelemetryProps {}

/**
 * A page displaying information about available updates for the server.
 */
export const SiteAdminDebugDataExportPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ telemetryService }) => {

    return (
        <div>
            <PageTitle title="Debug data export - Admin" />
            <H2>Debug data export</H2>
            <p>Seamlessly export data from your Sourcegrpah instance for easy debugging</p>

            <div>
            <H3>Data Included</H3>
            </div>

            <div>
            Data Included
            </div>


            <table>
            <td>Data Included</td>
            <td>Export Options</td>
            </table>

        </div>
    )
}
