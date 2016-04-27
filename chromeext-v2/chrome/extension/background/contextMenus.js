let windowId = 0;
const CONTEXT_MENU_ID = "sourcegraph_context_menu";

function closeIfExist() {
	if (windowId > 0) {
		chrome.windows.remove(windowId);
		windowId = chrome.windows.WINDOW_ID_NONE;
	}
}

// Consider adding back; proof-of-concept for standalone
// window.

// function popWindow(type) {
// 	closeIfExist();
// 	let options = {
// 		type: "popup",
// 		left: 100, top: 100,
// 		width: 800, height: 475
// 	};
// 	if (type === "open") {
// 		options.url = "window.html";
// 		chrome.windows.create(options, (win) => {
// 			windowId = win.id;
// 		});
// 	}
// }

// chrome.contextMenus.create({
// 	id: CONTEXT_MENU_ID,
// 	title: "Sourcegraph Chrome Extension",
// 	contexts: ["all"],
// 	documentUrlPatterns: [
// 		"https://github.com/*"
// 	]
// });

// chrome.contextMenus.onClicked.addListener((event) => {
// 	if (event.menuItemId === CONTEXT_MENU_ID) {
// 		popWindow("open");
// 	}
// });
