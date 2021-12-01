import classNames from 'classnames'
import React, { useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { Button } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { Badge } from '../../../../../components/Badge'
import { FeedbackPromptContent } from '../../../../../nav/Feedback/FeedbackPrompt'
import { Popover } from '../../../../insights/components/popover/Popover'

import styles from './Sidebar.module.scss'

const SIZE_STORAGE_KEY = 'catalog-sidebar-size'

interface Props {
    className?: string
}

export const Sidebar: React.FunctionComponent<Props> = ({ className, ...props }) => (
    <Resizable
        defaultSize={200}
        handlePosition="right"
        storageKey={SIZE_STORAGE_KEY}
        className={styles.resizable}
        element={<SidebarContent className={classNames('border-right w-100', className)} {...props} />}
    />
)

const SidebarContent: React.FunctionComponent<Props & { className?: string }> = ({ className, children }) => (
    <div className={classNames('d-flex flex-column', className)}>
        <h2 className="h6 font-weight-bold pt-2 px-2 pb-0 mb-0 d-none">
            <Link to="/catalog" className="d-flex align-items-center text-muted">
                <CatalogIcon className="icon-inline mr-1" /> Catalog
            </Link>
        </h2>
        {children}
    </div>
)
