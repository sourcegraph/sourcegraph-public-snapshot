export class AdvanceProgressStep {}

export class SelectOrg {
	constructor(org) {
		this.org = org;
	}
}

export class SelectItems {
	constructor(items, type, selectAll) {
		this.items = items;
		this.type = type;
		this.selectAll = selectAll;
	}
}
export class SelectItem {
	constructor(itemIndex, type, select) {
		this.itemIndex = itemIndex;
		this.type = type;
		this.select = select;
	}
}
