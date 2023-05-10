import {FC} from 'react';
import {PageTitle} from '../../../components/PageTitle';
import {
    Container,
    PageHeader,
} from '@sourcegraph/wildcard'
import styles from '../../insights/admin-ui/CodeInsightsJobs.module.scss';
import './own-status-page-styles.scss'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'


const data = [{
    'name': 'recent-contributors',
    'description': 'Calculates recent contributors one job per repository.',
    'isEnabled': true
},
    {
        'name': 'recent-views',
        'description': 'Calculates recent viewers from the events stored inside Sourcegraph.',
        'isEnabled': false
    }
]

export const OwnStatusPage: FC = () => (
    <div>
        <PageTitle title="Own status page"/>
        <PageHeader
            headingElement="h2"
            path={[{text: 'Own status page'}]}
            description="List of Own signal indexers and their status"
            className="mb-3"
        />
        <Container className={styles.root}>
            {data.map(job => (
                <li key={job.name} className={styles.job}>
                    <div className={styles.jobHeader}>
                        <h3 className={styles.jobName}>{job.name}</h3>
                        <span className={styles.jobStatus}>
                            <Toggle
                                id="job-enabled"
                                value={job.isEnabled}
                                />
                        </span>
                    </div>
                    <p className={styles.jobDescription}>{job.description}</p>
                </li>
            ))}
        </Container>
    </div>
)
