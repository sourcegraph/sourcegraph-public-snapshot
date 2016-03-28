var consumingSpan, annotating; 


function main() {
	mainCall();
	
	var pageScript = document.createElement('script');
	pageScript.innerHTML = '$(document).on("pjax:success", function () { var evt = new Event("PJAX_PUSH_STATE_0923"); document.dispatchEvent(evt); });';
	document.querySelector('body').appendChild(pageScript);
	$(document).one('PJAX_PUSH_STATE_0923', function() {
		console.log('pjax')
		mainCall();
	})
}
function mainCall() {

	var fileElem = document.querySelector('.file .blob-wrapper')
	if (fileElem) {
		getAnnotations()
	}

}

function getAnnotations() {
	url = document.URL;
	urlsplit = url.split("/");
	user = urlsplit[3]
	repo = urlsplit[4]
	branch = 'master';
	if (urlsplit[6] !== null && (urlsplit[5]==='tree' || urlsplit[5]==='blob')) {
		branch = urlsplit[6];
	}
	path = urlsplit[7];
	if (urlsplit.length > 8){
		for (var i = 8; i < urlsplit.length; i++){
			path = path + "/" + urlsplit[i] 
		}
	}

	
	$.ajax ({
		dataType: "json",
		method: "GET",
		url: "https://sourcegraph.com/.api/annotations?Entry.RepoRev.URI=github.com/"+user+"/"+repo+"&Entry.RepoRev.Rev="+branch+"&Entry.RepoRev.CommitID=&Entry.Path="+path+"&Range.StartByte=0&Range.EndByte=0"
	}).done(getSourcegraphRefLinks)
}


function getSourcegraphRefLinks(data) {
	var annsByStartByte = {};
	var annsByEndByte = {};
	for (var i = 0; i < data.Annotations.length; i++){
		if (data.Annotations[i].URL !== undefined) {
			var ann = data.Annotations[i];
			annsByStartByte[ann.StartByte] = ann;
			annsByEndByte[ann.EndByte] = ann;
		}
	}
	traverseDOM(annsByStartByte, annsByEndByte);
}


function traverseDOM(annsByStartByte, annsByEndByte){
	console.time("traverseDOM"); 

	var table = document.querySelector('table');
	var count = 0;
	var refLink;
	
	//get output HTML for each line and replace the original <td>
	for (var i = 0; i < table.rows.length; i++){
		var output = "";
		// Keep track of which row we're at.
		var row = table.rows[i];

		// Code is always the second <td> element.
		//CODE.INNERHTML IS WHAT WE WANT TO REPLACE WITH OUR STRING
		var code = row.cells[1]
		//console.log(code.innerHTML)
		var children = code.childNodes; // We want the children of the <td>
		var startByte = count;
		count += code.textContent.length;
		if (code.textContent !== "\n") {
    		count++; // newline
    	}
		//Go through each childNode
		for (var j = 0; j < children.length; j++) {

			//console.log(startByte)
			//console.log(children[j])
			var childNodeChars;

			if (children[j].nodeType === Node.TEXT_NODE){
				childNodeChars = children[j].nodeValue.split("")
			}
			
			else {
				childNodeChars = children[j].outerHTML.split("");    
			}
			
			//console.log(childNodeChars)

			//when we are returning the <span> element, we don't want to increment startByte
			consumingSpan = false;

			//keep track of whether we are currently on a definition with a link
			annotating = false;

            //go through each char of childNodes
            for (var k = 0; k < childNodeChars.length; k++) {
            	if (childNodeChars[k] === "<") {
            		consumingSpan = true;
            	}
            		
            	if (!consumingSpan) {
            		output += next(childNodeChars[k], startByte, annsByStartByte, annsByEndByte)
            		startByte++  
                }
                else {
                	output += childNodeChars[k]
                }

                if (childNodeChars[k] === ">") {
            		consumingSpan = false;
            	}


            }
		}
		//console.log(output)
		code.innerHTML = output;
		var newRows = row.cells[1].childNodes
		for (var n = 0; n < newRows.length; n++){
			sourcegraph_activateDefnPopovers(newRows[n])
				
		}
	
	}
	if (document.getElementsByClassName('sourcegraph-popover visible').length !== 0){
		document.getElementsByClassName('sourcegraph-popover visible')[0].classList.remove('visible')
	}
	console.timeEnd("traverseDOM")

}



function next(c, byteCount, annsByStartByte, annsByEndByte) {
	//console.log(annsByStartByte !== undefined, byteCount);
	//console.log("byteCount", byteCount, c);

	

	var matchDetails = annsByStartByte[byteCount];
	//console.log(matchDetails)
	
	//if there is a match
	if (annotating===false && matchDetails !== undefined) { 
		if (annsByStartByte[byteCount].EndByte - annsByStartByte[byteCount].StartByte === 1){
			return '<a href="https://sourcegraph.com' + matchDetails.URL+'" target="tab" class="sgdef">'+c+'</a>'
		}
		
		annotating = true;
		//console.log(byteCount) 
		return '<a href="https://sourcegraph.com' + matchDetails.URL+'" target="tab" class="sgdef">'+c
	}

	//if we reach the end of the child node - our counter = endByte of the match, annotating = false, close the tag.  
	if (annotating===true && annsByEndByte[byteCount+1]) {
		//console.log('end')
		//console.log(byteCount)
		annotating = false;
		return c+"</a>"
	}
	
	else {

		return c
	}


}
	
main();