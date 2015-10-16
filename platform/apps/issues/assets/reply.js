window.onload = function() {
  pencils = drawPencils();
}

function edit(replyNum) {
  return function() {
		var headElem = document.querySelector("head[data-csrf-token]");
		var csrfToken = headElem.dataset.csrfToken;

    id = 'reply'+replyNum;
    reply = document.getElementById(id);
    renderedReply = reply.innerHTML;
    u = window.location.pathname+'/'+replyNum;
    req = new XMLHttpRequest();
    req.onload = function() {
      raw = this.responseText
        reply.innerHTML = '<form action="'+u+'/edit" method="POST">'
				+ '<input type="hidden" name="csrf_token" value="'+csrfToken+'">'
        + '<textarea name="reply" rows="10" cols=80">'+raw+'</textarea>'
        + '<br>'
        + '<input type="submit" value="Submit">'
        + '<button type="button" id="discard'+replyNum+'">Discard</button>'
        + '</form>';
      discardButton = document.getElementById('discard'+replyNum);
      discardButton.onclick = function() {
        reply.innerHTML = renderedReply;
      }
    }
    raw = u+'/raw';
    req.open('get', raw, true);
    req.send();
  }
}

function drawPencils() {
  pencils = document.getElementsByClassName('edit');
  for (i = 0; i < pencils.length; i++) {
    p = pencils[i];
    p.onclick = edit(i+1);
  }
}
