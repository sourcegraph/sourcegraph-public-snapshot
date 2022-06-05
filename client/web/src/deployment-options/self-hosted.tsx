import { FunctionComponent } from 'react'
import * as H from 'history'

import { BestForTitle } from '.'
import Install from '../components/Install'
import { PageRoutes } from '../routes.constants'
import { ButtonLink } from '@sourcegraph/wildcard'
import { PageTitle } from '../components/PageTitle'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import ArrowRightIcon from 'mdi-react/ArrowRightIcon'

import styles from './styles.module.scss'

interface IProps {
    history: H.History
}

export const SelfHosted: FunctionComponent<IProps> = ({ history }) => {
    return (
        <div className="flex flex-1 mt-5">
            <PageTitle title="Get Started with Sourcegraph Self-Hosted" />
            <div className={styles.hero}>
                <div className="container-xl py-5">
                    <h1 className="display-3 mb-2">
                        <strong>What's best for you?</strong>
                    </h1>
                    <p>From Amazon to Uber, the world's best developers use Sourcegraph every day.</p>
                </div>
            </div>
            <div className={`${styles.selfHostedRoot} py-5`}>
                <div className="row container-xl mx-auto py-5">
                    <div className="col-lg-6 pr-lg-4">
                        <div>
                            <ButtonLink className="p-0 mb-3" to={PageRoutes.DeploymentOptions}>
                                <ArrowLeftIcon /> DEPLOYMENT OPTIONS
                            </ButtonLink>

                            <h1 className="font-weight-bolder mb-2">
                                Sourcegraph <br />
                                Self-Hosted
                            </h1>

                            <p className="mt-4">
                                Deploy and control Sourcegraph in your own infrastructure, or use Docker to install
                                locally. Get started for free.
                            </p>

                            <BestForTitle />
                            <p>Teams and enterprises</p>

                            <p>
                                Collaborate with your team on any code host (including private hosts) and access
                                advanced security functionality.
                            </p>
                        </div>
                    </div>

                    <div className="col-lg-6 py-4 py-lg-0 pl-lg-4">
                        <Install />

                        <div className="d-flex flex-column align-items-start">
                            <a className="mt-5" href="https://info.sourcegraph.com/talk-to-a-developer" target="_blank">
                                Talk to an engineer <ArrowRightIcon />
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default SelfHosted
