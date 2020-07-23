window.sgdocs = (() => {
  let VERSION_SELECT_BUTTON,
    CONTENT_NAV,
    BREADCRUMBS,
    BREADCRUMBS_DATA = [],
    MOBILE_NAV_BUTTON,
    START_SOURCEGRAPH_COMMAND_SNIPPET

  return {
    init: breadcrumbs => {
      BREADCRUMBS_DATA = breadcrumbs ? breadcrumbs : []
      BREADCRUMBS = document.querySelector('#breadcrumbs')
      BREADCRUMBS_MOBILE = document.querySelector('#breadcrumbs-mobile')

      VERSION_SELECTOR = document.querySelector('#version-selector')
      VERSION_SELECT_BUTTON = VERSION_SELECTOR.querySelector('#version-selector button')
      VERSION_OPTIONS = VERSION_SELECTOR.querySelector('#version-selector details-menu')

      CONTENT_NAV = document.querySelector('#content-nav')

      MOBILE_NAV_BUTTON = BREADCRUMBS_MOBILE.querySelector('input[type="button"]')

      START_SOURCEGRAPH_COMMAND_SNIPPET = document.querySelector('.start-sourcegraph-command') // Assumes only one per page

      versionSelectorInit()
      mobileNavInit()
      navInit()
      scrollNavToSelected()
      docsVersionLinks()
      startSourcegraphCommandInit()
      setTimeout(schemaLinkCheck, 0) // Browser scrolls straight to element without this
    },

    scrollToElement: scrollToElement
  }

  /**
   * Smoothly scroll to an element
   *
   * @param {HTMLElement} element Element to scroll to
   * @param {number} elementOffsetTop Optionally reduce vertical scroll distance
   */
  function scrollToElement(parent = document.body, element, elementOffsetTop = 0) {
    if (!element) {
      return
    }

    parent.scrollTo({
      top: element.offsetTop - elementOffsetTop,
      left: 0,
      behavior: 'smooth',
    })
  }

  function versionSelectorInit() {
    function outsideVersionSelectorListener(event) {
      if (!event.target.closest('#version-selector')) {
        hideMenu()
      }
    }

    function escaped(e) {
      if (e.which === 27) {
        e.preventDefault()
        hideMenu()
      }
    }
    document.addEventListener('keydown', escaped)

    function hideMenu() {
      VERSION_OPTIONS.classList.remove('show')
      document.removeEventListener('click', outsideVersionSelectorListener)
      document.removeEventListener('keydown', escaped)
    }

    VERSION_SELECT_BUTTON.addEventListener('click', e => {
      VERSION_OPTIONS.classList.toggle('show')
      document.addEventListener('click', outsideVersionSelectorListener)
      document.addEventListener('keydown', escaped)
    })
  }

  function mobileNavInit() {
    MOBILE_NAV_BUTTON.addEventListener('click', e => {
      CONTENT_NAV.classList.toggle('mobile-show')
      document.body.classList.toggle('fix-document.body')
      BREADCRUMBS_MOBILE.classList.toggle('fixed')
    })
  }

  function navInit() {
    if (BREADCRUMBS_DATA[1]) {
      document
        .querySelector(`ul.content-nav-section[data-nav-section="${BREADCRUMBS_DATA[1].Label}"]`)
        .classList.toggle('expanded')
      document
        .querySelector(`ul.content-nav-section a[href="${BREADCRUMBS_DATA[BREADCRUMBS_DATA.length - 1].URL}"]`)
        .parentNode.classList.add('selected')
    }

    document.querySelectorAll('button.content-nav-button').forEach(el => {
      el.addEventListener('click', e => e.srcElement.closest('.content-nav-section').classList.toggle('expanded'))
    })

    document.querySelectorAll('.content-nav a').forEach(el => (el.title = el.text.trim()))
  }

  function scrollNavToSelected() {
    setTimeout(() => scrollToElement(CONTENT_NAV, CONTENT_NAV.querySelector('.selected')), 50)
  }

  /**
   * Make the docs nav branch aware by dynamically adding the branch prefix if non-master.
   *
   * NOTE: This will cause 404s in instances where the nav has been updated in master to
   * point to a new file which doesn't exist in a previous branch, but there is no easy fix
   * as the document.html template always reflects master because only the docs content is
   * loaded dynamically.
   */
  function docsVersionLinks() {
    let urlPath = window.location.pathname

    if (urlPath.indexOf('@') === -1) {
      return
    }

    let branchPrefix = urlPath.match(/(@[\d\w\.-]+)/)[0]
    document.querySelectorAll('#content-nav a').forEach(e => e.setAttribute('href', `/${branchPrefix}${e.getAttribute('href')}`))
  }

  /**
   * Check the URL to see if navigation to a schema key is desired
   */
  function schemaLinkCheck() {
    const schemaDocSelector = '.json-schema-doc'
    const offsetTop = document.querySelector('.global-navbar').clientHeight + 20
    const schemaDocs = document.querySelectorAll(schemaDocSelector)

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

        // Temporarily change the id of the anchor to prevent the browser
        // from scrolling to the element
        anchor.addEventListener('click', e => {
          let originalId = e.target.id
          e.target.id = `${originalId}-id-miss`
          setTimeout(() => e.target.id = originalId, 1000)
        })
      })
    })

    // If URL hash is set and matches a schema key, scroll to it
    let targetKey = document.querySelector(`${schemaDocSelector} ${window.location.hash}`)
    if (window.location.hash && targetKey) {
      scrollToElement(targetKey, offsetTop)
    }
  }

  function gaConversionOnStartSourcegraphCommands() {
    if (window && window.gtag) {
      window.gtag('event', 'conversion', {
        'send_to': 'AW-868484203/vOYoCOCUj7EBEOuIkJ4D',
      });
    }
  }

  function startSourcegraphCommandInit() {
    if (!START_SOURCEGRAPH_COMMAND_SNIPPET) {
      return
    }

    START_SOURCEGRAPH_COMMAND_SNIPPET.addEventListener('click', gaConversionOnStartSourcegraphCommands)
  }

})()
