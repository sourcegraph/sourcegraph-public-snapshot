import { FC } from 'react';

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations';
import { Button, Icon } from '@sourcegraph/wildcard';

import { getCodeHostIcon, getCodeHostName } from '../../helpers';

import styles from './CodeHostsPicker.module.scss';

const SUPPORTED_CODE_HOSTS = [
    ExternalServiceKind.GITHUB,
    ExternalServiceKind.GITLAB,
    ExternalServiceKind.BITBUCKETCLOUD,
]

export const CodeHostsPicker: FC = () => (
    <section>
        <header className={styles.header}>
            <span>Add another remote code host</span>
            <small className='text-muted'>Choose a provider from the list below</small>
        </header>

        <ul className={styles.list}>
            {SUPPORTED_CODE_HOSTS.map(codeHostType =>
                <li key={codeHostType}>
                    <Button variant='secondary' outline={true} className={styles.item}>
                        <Icon svgPath={getCodeHostIcon(codeHostType)} aria-hidden={true}/>
                        <span>{getCodeHostName(codeHostType)}</span>
                    </Button>
                </li>
            )}
        </ul>
    </section>
)
