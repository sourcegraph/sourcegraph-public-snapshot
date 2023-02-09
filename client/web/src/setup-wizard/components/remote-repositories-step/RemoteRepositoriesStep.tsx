import { FC, HTMLAttributes, ReactElement } from 'react';

import classNames from 'classnames';
import { Routes, Route } from 'react-router-dom-v5-compat';

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations';
import { Button, Container, Icon, Text } from '@sourcegraph/wildcard';

import { CustomNextButton } from '../setup-steps';

import { CodeHostsNavigation } from './components/navigation';
import { getCodeHostIcon, getCodeHostName } from './helpers';

import styles from './RemoteRepositoriesStep.module.scss'

interface RemoteRepositoriesStepProps extends HTMLAttributes<HTMLDivElement> {}

export const RemoteRepositoriesStep: FC<RemoteRepositoriesStepProps> = props => {
    const { className, ...attributes } = props

    return (
        <div {...attributes} className={classNames(className, styles.root)}>
            <Text className='mb-2'>Connect remote code hosts where your source code lives.</Text>

            <section className={styles.content}>
                <Container className={styles.contentNavigation}>
                    <CodeHostsNavigation className={styles.navigation}/>
                </Container>

                <Container className={styles.contentMain}>
                    <Routes>
                        <Route path="" element={<CodeHostPicker/>} />
                        <Route path=":codehost/create" element={<span>Hello creation UI</span>} />
                    </Routes>
                </Container>
            </section>

            <CustomNextButton label="Custom next step label" disabled={true} />
        </div>
    )
}

enum ExternalCodeHostType {
  GitHub,
  GitLab,
  BitBucket
}

const SUPPORT_CODE_HOSTS = [
    ExternalServiceKind.GITHUB,
    ExternalServiceKind.GITLAB,
    ExternalServiceKind.BITBUCKETCLOUD,
]

interface CodeHostPickerProps {}

function CodeHostPicker(props: CodeHostPickerProps): ReactElement {
    return (
        <section className={styles.codeHostPicker}>
            <header className={styles.codeHostPickerHeader}>
                <span>Add another remote code host</span>
                <small className='text-muted'>Choose a provider from the list below</small>
            </header>

            <ul className={styles.codeHostPickerList}>
                { SUPPORT_CODE_HOSTS.map(codeHostType =>
                    <li key={codeHostType}>
                        <Button variant='secondary' outline={true} className={styles.codeHostPickerItem}>
                            <Icon svgPath={getCodeHostIcon(codeHostType)} aria-hidden={true}/>
                            <span>{getCodeHostName(codeHostType)}</span>
                        </Button>
                    </li>
                )}
            </ul>
        </section>
    )
}
