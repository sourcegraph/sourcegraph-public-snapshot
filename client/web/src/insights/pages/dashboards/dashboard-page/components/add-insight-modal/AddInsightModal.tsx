import Dialog from '@reach/dialog';
import { VisuallyHidden } from '@reach/visually-hidden';
import classnames from 'classnames';
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useState } from 'react';
import { noop } from 'rxjs';

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings';

import { Settings } from '../../../../../../schema/settings.schema';
import { InsightVisibility } from '../../../../../components/insight-visibility/InsightVisibility';
import { createDashboardsMap } from '../../../../../components/insight-visibility/utils';
import {
    isRealDashboard,
    RealInsightDashboard
} from '../../../../../core/types';
import { useDashboards } from '../../../../../hooks/use-dashboards/use-dashboards';
import { useInsight } from '../../../../../hooks/use-insight/use-insight';

import styles from './AddInsightModal.module.scss'

export interface AddInsightModalProps extends SettingsCascadeProps<Settings> {
    onClose?: () => void
    insightId: string
}

export const AddInsightModal: React.FunctionComponent<AddInsightModalProps> = props => {
    const {
        insightId,
        settingsCascade,
        onClose = noop,
    } = props

    const dashboards = useDashboards(settingsCascade)
    const insight = useInsight({ settingsCascade, insightId })
    const [pickedDashboards, setPickedDashboards] = useState<Record<string, RealInsightDashboard>>(() => {
        if (!insight) {
            return {}
        }

        // Calculate initial value for the insight visibility (dashboard list)
        return createDashboardsMap(...insight.dashboards)
    })

    const handleChange = (dashboards: Record<string, RealInsightDashboard>): void => {
        console.log(dashboards)

        setPickedDashboards(dashboards)
    }

    if (!insight) {
        // implement no insight state
        return <span>The insight with { insightId } id wasn't found.</span>
    }

    const realDashboards = dashboards.filter(isRealDashboard)

    return (
        <Dialog className={styles.modal} onDismiss={close}>
            <button type='button' className={classnames('btn btn-icon', styles.closeButton)} onClick={onClose}>
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon/>
            </button>

            <h2>Add insight to dashboard</h2>

            <span>Assign ‘{insight.title}’ to Dashboards</span>

            {/* eslint-disable-next-line react/forbid-elements */}
            <form noValidate={true} >

                <h3>Dashboards</h3>
                <InsightVisibility
                    value={pickedDashboards}
                    dashboards={realDashboards}
                    onChange={handleChange}/>
            </form>
        </Dialog>
    )
}
