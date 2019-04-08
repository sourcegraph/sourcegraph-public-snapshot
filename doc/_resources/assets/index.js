window.sgdocs = (global => {
  let VERSION_SELECT_BUTTON,
    SEARCH_FORMS,
    CONTENT_NAV,
    BODY,
    BREADCRUMBS,
    BREADCRUMBS_DATA = [],
    MOBILE_NAV_BUTTON

  return {
    init: () => {
      SEARCH_FORMS = document.querySelectorAll('.search-form')
      VERSION_SELECTOR = document.querySelector('#version-selector')
      VERSION_SELECT_BUTTON = VERSION_SELECTOR.querySelector('#version-selector button')
      VERSION_OPTIONS = VERSION_SELECTOR.querySelector('#version-selector details-menu')
      CONTENT_NAV = document.querySelector('#content-nav')
      BODY = document.querySelector('body')
      BREADCRUMBS = document.querySelector('#breadcrumbs')
      MOBILE_NAV_BUTTON = BREADCRUMBS.querySelector('input[type="button"]')
      BREADCRUMBS_DATA = global.SGDOCS_BREADCRUMBS ? global.SGDOCS_BREADCRUMBS : []

      searchInit()
      versionSelectorInit()
      breadcrumbsInit()
      mobileNavInit()
      navInit()
    },
  }

  function searchInit() {
    SEARCH_FORMS.forEach(form => {
      form.addEventListener('submit', e => {
        const search = e.srcElement.querySelector('input[name="search"]').value
        e.preventDefault()
        window.location.href =
          'https://www.google.com/search?ie=UTF-8&q=site%3Adocs.sourcegraph.com+' + encodeURIComponent(search)
      })
    })
  }

  function versionSelectorInit() {
    function outsideVersionSelectorListener(event) {
      if (!event.target.closest('#version-selector')) {
        VERSION_OPTIONS.classList.toggle('show')
        hideMenu
      }
    }

    function escaped(e) {
      if (e.which === 27) {
        e.preventDefault()
        hideMenu()
      }
    }

    function hideMenu() {
      VERSION_OPTIONS.classList.toggle('show')
      document.removeEventListener('click', outsideVersionSelectorListener)
      document.removeEventListener('keydown', escaped)
    }

    VERSION_SELECT_BUTTON.addEventListener('click', e => {
      VERSION_OPTIONS.classList.toggle('show')
      document.addEventListener('click', outsideVersionSelectorListener)
      document.addEventListener('keydown', escaped)
    })
  }

  function breadcrumbsInit() {
    if (BREADCRUMBS_DATA.length === 0) {
      return
    }
  }

  function mobileNavInit() {
    MOBILE_NAV_BUTTON.addEventListener('click', e => {
      CONTENT_NAV.classList.toggle('mobile-show')
      BODY.classList.toggle('fix-body')
      BREADCRUMBS.classList.toggle('fixed')
    })
  }

  function navInit() {
    if (BREADCRUMBS_DATA[1]) {
      document
        .querySelector(`ul.content-nav-section[data-nav-section="${window.SGDOCS_BREADCRUMBS[1].Label}"]`)
        .classList.toggle('expanded')
    }

    document
      .querySelector(`ul.content-nav-section a[href="${BREADCRUMBS_DATA[BREADCRUMBS_DATA.length - 1].URL}"]`)
      .parentNode.classList.add('selected')

    document.querySelectorAll('button.content-nav-button').forEach(el => {
      el.addEventListener('click', e => e.srcElement.closest('.content-nav-section').classList.toggle('expanded'))
    })
  }
})(window)
