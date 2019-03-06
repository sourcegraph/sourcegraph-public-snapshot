//Index.js
// Docs site JS code


/**
 * Toggle search dropdown on button click
 * Accepts the ID for the dropdown to toggle
 * @param {string} dropdownName
 */
const toggleDropdown = dropdownName => {
  const dropdown = document.getElementById(dropdownName)
  dropdown.classList.toggle('show')
}

// Listen for click outside search form to close dropdowns
document.addEventListener('click', e => {
  // Consts for dropdown areas (dropdown and button)
  const searchDropdownArea = document.getElementById('search-form')
  const versionDropdownArea = document.getElementById('version-form')

  // Const for navbar
  const navArea = document.getElementById('globalNav')

  // Consts for the actual dropdown element
  const searchDropdown = document.getElementById('searchDropdown')
  const versionDropdown = document.getElementById('versionDropdown')

  // Const for navbar checkbox
  const navState = document.getElementById('nav-state')

  let targetElement = e.target

  do {
    if (targetElement == searchDropdownArea) {
      // If clicked area is in search dropdown remove all other dropdowns
      versionDropdown.classList.remove('show')
      return
    } else if (targetElement == versionDropdownArea) {
      // If clicked area is in version dropdown remove all other dropdowns
      searchDropdown.classList.remove('show')
      return
    } else if (targetElement == navArea) {
      // Because this is ran after the other two check it is okay to close all dropdowns
      // even through they are contained in nav.
      versionDropdown.classList.remove('show')
      searchDropdown.classList.remove('show')
      return
    }
    targetElement = targetElement.parentNode
  } while (targetElement)

  // If clicked area is outside all dropdowns remove all other dropdowns
  versionDropdown.classList.remove('show')
  searchDropdown.classList.remove('show')

  // Close dropdown on mobile
  navState.checked = false
})

// Open and close nav section
const toggleNavSection = navSection => {
  const section = document.getElementById(navSection)
  section.classList.toggle('expanded')
}

// Only open nav section
const openNavSection = navSection => {
  const section = document.getElementById(navSection)
  if (section.classList.contains('expanded')) {
    // Section is already expanded
    return
  } else {
    section.classList.add('expanded')
    return
  }
}

// Open nav section based off current page
const breadcrumbNavToggle = () => {
  // Check if breadcrumb is defined
  if (breadcrumbs.length === 0) {
    // If breadcrumb isn't defined (docs home) do nothing
    return
  }

  var currentCatagory = breadcrumbs[1].Label

  // Open nav pannel based off current catagory
  switch (currentCatagory) {
    case 'user':
      openNavSection('contentNavUser')
      break
    case 'admin':
      openNavSection('contentNavAdmin')
      break
    case 'extensions':
      openNavSection('contentNavExtension')
      break
    case 'dev':
      openNavSection('contentNavDev')
      break
    case 'api':
      openNavSection('contentNavAPI')
      break
    case 'integration':
      openNavSection('contentNavIntegration')
      break
    default:
      openNavSection('contentNavUser')
      break
  }
}

// Toggle content nav on mobile
const toggleContentNav = () => {
  const contentNav = document.getElementById('contentNav')
  const body = document.getElementById('body')
  const breadcrumbs = document.getElementById('breadcrumbs')
  contentNav.classList.toggle('mobile-show')
  body.classList.toggle('fix-body')
  breadcrumbs.classList.toggle('fixed')
}
