import React from 'react'
import addons, { types } from '@storybook/addons'
import { Icons, IconButton } from '@storybook/components'
import { useRedesignToggle, REDESIGN_CLASS_NAME } from '../../client/shared/src/util/useRedesignToggle'

const toggleRedesignClass = (element, isRedesignEnabled) => {
  element.classList.toggle(REDESIGN_CLASS_NAME, !isRedesignEnabled)
}

const updatePreview = isRedesignEnabled => {
  const iframe = document.getElementById('storybook-preview-iframe')

  if (!iframe) {
    return
  }

  const iframeDocument = iframe.contentDocument || iframe.contentWindow?.document
  const body = iframeDocument?.body

  toggleRedesignClass(body, isRedesignEnabled)
}

const updateManager = isRedesignEnabled => {
  const manager = document.querySelector('body')

  if (!manager) {
    return
  }

  toggleRedesignClass(manager, isRedesignEnabled)
}

const RedesignToggleStorybook = () => {
  const { isRedesignEnabled, setIsRedesignEnabled } = useRedesignToggle()

  const handleRedesignToggle = () => {
    setIsRedesignEnabled(!isRedesignEnabled)
    updatePreview(isRedesignEnabled)
    updateManager(isRedesignEnabled)
  }

  return (
    <IconButton
      key="redesign-toolbar"
      active={isRedesignEnabled}
      title={isRedesignEnabled ? 'Disable redesign theme' : 'Enable redesign theme'}
      onClick={handleRedesignToggle}
    >
      <Icons icon="beaker" />
    </IconButton>
  )
}

/**
 * Custom toolbar which renders button to toggle redesign theme global CSS class.
 */
addons.register('sourcegraph/redesign-toggle-toolbar', () => {
  addons.add('sourcegraph/redesign-toggle-toolbar', {
    title: 'Redesign toggle toolbar',
    type: types.TOOL,
    match: ({ viewMode }) => viewMode === 'story' || viewMode === 'docs',
    render: RedesignToggleStorybook,
  })
})
