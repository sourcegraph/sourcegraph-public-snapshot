import { Tab, useTabsContext } from '@reach/tabs'
import React, { useEffect } from 'react'

interface RepositoryTabProps {
    children: React.ReactNode
    onHandleDisplayAction: React.Dispatch<React.SetStateAction<boolean>>
}

export const RepositoryTab: React.FunctionComponent<RepositoryTabProps> = ({ onHandleDisplayAction, children }) => {
    const { selectedIndex } = useTabsContext()
    useEffect(() => {
        onHandleDisplayAction(selectedIndex === 0)
    }, [selectedIndex, onHandleDisplayAction])

    return <Tab>{children}</Tab>
}
