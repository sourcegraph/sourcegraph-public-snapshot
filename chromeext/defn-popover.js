var ajaxCache = {};
function sourcegraph_activateDefnPopovers(el) {
  var request;
  var activeA;

  $(window).on('popstate', function(){
    if (showingPopover) {
      popover.classList.remove("visible");
    }
  });

  el.addEventListener("mouseout", function(ev) {
    var t = __sg_getTarget(ev.toElement);
      setTimeout(hidePopover);
      activeA = null;
  });
  el.addEventListener("mouseover", function(ev) {
    var t = __sg_getTarget(ev.target);
    if (!t) return;
    // if (activeA != t) console.log('ACTIVE:', t.href);
    if (activeA != t) {
      activeA = t;
      //need to add /.ui to URL, so split and take everything after sourcegraph.com
      var URLend = activeA.href.split('https://sourcegraph.com')[1]
      var URLtoUse = 'https://sourcegraph.com/.api/repos' + URLend;
      ajaxGet(URLtoUse, function(html) {
        if (activeA) showPopover(html);
      });
    }
    preShowingX = ev.pageX;
    preShowingY = ev.pageY;
    positionPopover(ev.pageX, ev.pageY);
  });

  function __sg_getTarget(t) {
    while (t && t.tagName == "SPAN") { t = t.parentNode; }
    if (t && t.tagName == "A" && t.classList.contains("sgdef")) return t;
  }

  var popover;
  var showingPopover;
  var preShowingX, preShowingY;
  function showPopover(html) {
    showingPopover = true;
    if (!popover) {
      popover = document.createElement("div");
      popover.classList.add("sourcegraph-popover");
      popover.style.position = "absolute";
      document.body.appendChild(popover);
    }
    positionPopover(preShowingX, preShowingY);
    popover.innerHTML = html;
    popover.classList.add("visible");
    popover.style.display = "block";
  }

  function hidePopover() {
    if (!popover) return;
    showingPopover = false;
    setTimeout(function() {
      if (!showingPopover) popover.style.display = "none";
    }, 100);
    popover.classList.remove("visible");
  }



  function positionPopover(x, y) {
    if (!popover || !showingPopover) return;
    popover.style.top = (y + 15) + "px";
    popover.style.left = (x + 15) + "px";
  }

  function ajaxGet(url, cb) {
    if (ajaxCache[url]) {
      //console.log('cached', ajaxCache)
      cb(ajaxCache[url]);
      return;
    }
    request = new XMLHttpRequest();
    request.open('GET', url, true);
    request.onload = function() {
      console.log(request)
      var response = (JSON.parse(request.response));
      if (request.status >= 200 && 400 > request.status) {
        console.log('success')
        var html;
        console.log(response)
        if (response.Data) {
          console.log('data')
          if (response.DocHTML){
            html = "<div><span class='title'>" + response.QualifiedName.__html +"</span>\n<span class='p'>"+ response.DocHTML.__html + "</span>\n<span class='repo'>" + response.Repo + "</span></div>";  
          }
          else {
            html = "<div><span class='title'>" + response.QualifiedName.__html + "</span></br><span class='repo'>" + response.Repo + "</span></div>";
          }
          ajaxCache[url] = html;
          cb(html);
        }
      } 
      else if (request.readyState > 1) {
        console.error("Sourcegraph error getting definition info.", JSON.stringify(request));
      }
    }
    
    request.onerror = function() { console.error("Sourcegraph error getting definition info."); };
    request.send();
  }
}


