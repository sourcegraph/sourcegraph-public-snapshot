javascript:
(function(){
    // This bookmarklet jumps between the same URL paths on
    // http://localhost:3000 and https://sourcegraph.com.
    var p = window.location.protocol + "//" + window.location.host;
    var p2 = (p=="http://localhost:3000") ? "https://sourcegraph.com" : "http://localhost:3000";
    var dst=window.location.href.replace(p,p2);window.location=dst;
})()
