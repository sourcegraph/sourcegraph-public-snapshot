chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
    if(request.query === "whoami"){
         console.log("providing tab information");
         sendResponse({tabUrl:sender.tab.url});
    }
});

