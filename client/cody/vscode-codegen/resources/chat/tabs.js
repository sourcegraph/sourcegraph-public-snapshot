/* eslint-disable no-undef */
class TabsController {
	constructor(model) {
		this.model = model
		this.tabs = [...document.querySelectorAll('li[data-tab]')]
		this.containers = [...document.querySelectorAll('[data-tab-target]')]

		const onTabClick = e => {
			const tabName = e.target.dataset.tab
			this.setSelectedTab(tabName)
		}
		this.tabs.forEach(tab => tab.addEventListener('click', onTabClick))

		this.renderTab()
	}

	setSelectedTab(newSelectedTab) {
		const changed = this.model.selectedTab !== newSelectedTab
		if (changed) {
			this.model.selectedTab = newSelectedTab
			this.renderTab()
		}
	}

	renderTab() {
		const tabElement = document.querySelector(`[data-tab="${this.model.selectedTab}"]`)
		const tabTargetContainer = document.querySelector(`[data-tab-target="${this.model.selectedTab}"]`)

		for (const tab of this.tabs) {
			tab.classList.remove('tab-menu-item-selected')
		}

		for (const container of this.containers) {
			container.classList.remove('show')
		}

		tabElement.classList.add('tab-menu-item-selected')
		tabTargetContainer.classList.add('show')
	}
}
