// Theme
const currentThemePreference = () => localStorage.getItem('theme-preference') || 'auto'
const themePreferenceButtons = () => document.querySelectorAll('body > #sidebar #theme button[data-theme-preference]')
const currentTheme = (pref = currentThemePreference()) => pref === 'auto' ? window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light' : pref
const applyThemePreference = (pref) => {
  localStorage.setItem('theme-preference', pref)
  applyTheme()
}
const applyTheme = () => {
  const pref = currentThemePreference()
  for (const button of themePreferenceButtons()) {
    button.classList.toggle('active', button.dataset.themePreference === pref)
  }

  const theme = currentTheme(pref)
  document.body.classList.toggle('theme-light', theme === 'light')
  document.body.classList.toggle('theme-dark', theme === 'dark')
}
document.addEventListener('DOMContentLoaded', () => {
  window.matchMedia('(prefers-color-scheme: dark)').addListener(e => applyTheme())
  applyTheme()
  for (const button of themePreferenceButtons()) {
    button.addEventListener('click', e => {
      applyThemePreference(e.currentTarget.dataset.themePreference)
    });
  }
  // Toggle theme on Option+T/Alt+T/Meta+T
  document.body.addEventListener('keydown', e => {
    if ((e.metaKey || e.altKey) && e.key === 't') {
      const theme = currentTheme();
      applyThemePreference(theme === 'light' ? 'dark' : 'light')
    }
  })
})

// Sidebar
const pagePath = location.pathname
const quote = str => JSON.stringify(str.replace(/[^a-zA-Z0-9._\/-]/g, ''))
document.addEventListener('DOMContentLoaded', () => {
  const style = document.createElement('style')
  for (const link of document.querySelectorAll('body > #sidebar .nav-section.tree a')) {
    const current = link.pathname === pagePath
    const expand = current || pagePath.startsWith(link.pathname + '/')
    const subsection = link.pathname.split('/').length >= 3

    const item = link.parentNode
    item.classList.toggle('current', current)
    item.classList.toggle('expand', expand)
    item.classList.toggle('active-subsection', subsection && expand)
    item.classList.toggle('collapse', !expand)
  }
})

// JSON Schema keys
document.addEventListener('DOMContentLoaded', () => {
  const schemaDocs = document.querySelectorAll('.json-schema-doc')
  // Find spans that contain a key and swap them for anchors for hover functionality
  schemaDocs.forEach(schemaDoc => {
    schemaDoc.querySelectorAll(`span`).forEach(el => {
      const keyNameMatch = el.innerText.match(/^"(.*)"/)
      const isKey = el.nextSibling && el.nextSibling.textContent.includes(':')
      if (!isKey || !keyNameMatch) {
        return
      }

      // Add a named anchor to get the hover functionality we need
      const keyText = keyNameMatch[1]
      const id = keyText.replace(/\./g, '-')
      const anchor = document.createElement('a')
      anchor.id = id
      anchor.className = 'schema-doc-key'
      anchor.href = '#' + id
      anchor.rel = 'nofollow'
      anchor.textContent = `"${keyText}"`
      anchor.style.color = el.style.color
      el.replaceWith(anchor)
    })
  })
  // If URL hash is set and matches a schema key, scroll to it
  if (window.location.hash) {
    let targetKey = document.getElementById(window.location.hash.slice(1))
    if (targetKey && Array.from(schemaDocs).some(e => e.contains(targetKey))) {
      setTimeout(() => {
        // Scroll to slightly above the element so the docstring is visible.
        targetKey.scrollIntoView(true)
        window.scrollBy(0, -150)
      }, 0)
    }
  }
})

// Conversion tracking for copying the quickstart command
document.addEventListener('DOMContentLoaded', () => {
  const startSourcegraphCommand = document.querySelector('.start-sourcegraph-command') // Assumes only one per page
  const gaConversionOnStartSourcegraphCommands = () => {
    // TODO(sqs): doesn't seem to be working: https://github.com/sourcegraph/sourcegraph/issues/14323
    if (window && window.gtag) {
      window.gtag('event', 'conversion', {
        'send_to': 'AW-868484203/vOYoCOCUj7EBEOuIkJ4D',
      });
    }
  }
  if (startSourcegraphCommand) {
    startSourcegraphCommand.addEventListener('click', gaConversionOnStartSourcegraphCommands)
  }
})

// Cloud CTA clicks
document.addEventListener('DOMContentLoaded', () => {
  const cloudCTAs = document.querySelectorAll('.cloud-cta')

  cloudCTAs.forEach(cloudCTA => {
    cloudCTA.addEventListener('click', () => {
      if (window && window.plausible) {
        window.plausible('ClickedOnFreeTrialCTA')
      }
    })
  })
})

// Promise to wait on for DOMContentLoaded
const domContentLoadedPromise = new Promise(resolve => {
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', resolve)
  } else {
    resolve()
  }
})

// Import Sourcegraph's EventLogger
const importEventLoggerPromise = import('https://cdn.skypack.dev/@sourcegraph/event-logger')

Promise.all([domContentLoadedPromise, importEventLoggerPromise]).then(([_, { EventLogger }]) => {
  const eventLogger = new EventLogger('https://sourcegraph.com')

  // Log a Docs page view event
  const eventArguments = { path: window.location.pathname }
  console.log('ViewStaticPage', eventArguments) // TODO: remove this console log
  eventLogger.log('ViewStaticPage', eventArguments, eventArguments)

  // Log download links which have a "data-download-name" attribute
  // This is used to track Sourcegraph App download links
  document.addEventListener('click', event => {
    if (event.target.matches('a[data-download-name]')) {
      const downloadName = event.target.getAttribute('data-download-name')
      const eventArguments = {
        path: window.location.pathname,
        downloadSource: 'docs',
        downloadName,
      }

      console.log('DownloadClick', eventArguments) // TODO: remove console log
      eventLogger.log('DownloadClick', eventArguments, eventArguments)
    }
  })
})
