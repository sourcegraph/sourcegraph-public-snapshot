export class WantDef {
	constructor(url) {
		this.url = url;
	}
}

export class DefFetched {
	constructor(url, def) {
		this.url = url;
		this.def = def;
	}
}

export class SelectDef {
	constructor(url) {
		this.url = url;
	}
}

export class SelectMultipleDefs {
	constructor(urls, left, top) {
		this.urls = urls;
		this.left = left;
		this.top = top;
	}
}

export class HighlightDef {
	constructor(url) {
		this.url = url;
	}
}

export class WantExample {
	constructor(defURL, index) {
		this.defURL = defURL;
		this.index = index;
	}
}

export class ExampleFetched {
	constructor(defURL, index, example) {
		this.defURL = defURL;
		this.index = index;
		this.example = example;
	}
}

export class NoExampleAvailable {
	constructor(defURL, index) {
		this.defURL = defURL;
		this.index = index;
	}
}

export class WantDiscussions {
	constructor(defURL) {
		this.defURL = defURL;
	}
}

export class DiscussionsFetched {
	constructor(defURL, discussions) {
		this.defURL = defURL;
		this.discussions = discussions;
	}
}

export class CreateDiscussion {
	constructor(defURL, title, description, callback) {
		this.defURL = defURL;
		this.title = title;
		this.description = description;
		this.callback = callback;
	}
}

export class CreateDiscussionComment {
	constructor(defURL, discussionID, body, callback) {
		this.defURL = defURL;
		this.discussionID = discussionID;
		this.body = body;
		this.callback = callback;
	}
}
