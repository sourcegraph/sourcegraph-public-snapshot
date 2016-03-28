//Sourcegraph Token Code Search
var url, query, user, repo, branch, 
	original, table, nomatch, notextmatch,
	taburl, 
	getDefs, getText, prevFile, commitID,
	repoIsGo = true;
var current = '';
var logo2 = document.createElement('img');
logo2.src = chrome.extension.getURL("assets/src4.png");
var search = document.createElement('img');
search.src = chrome.extension.getURL("assets/search.png");
var t = document.createElement('img')
t.src = chrome.extension.getURL ("assets/t.png")

//get response from background page 
chrome.runtime.sendMessage({query: 'whoami'}, function(response){
	comboUsed = false;
	taburl = response.tabUrl;
	var urlsplit = taburl.split("/")
	user = urlsplit[3];
	repo = urlsplit[4];
	branch = 'master';
	if (urlsplit[6] !== null && (urlsplit[5]==='tree' || urlsplit[5]==='blob')) {
		branch = urlsplit[6];
	}
	//if repo page
	if (document.getElementsByClassName('repository-meta').length !== 0){
		commitID = document.getElementsByClassName('commit-tease-sha')[0].href.split("/")[6]
	}
	//if file page
	if (document.getElementsByClassName('file').length !== 0) {
		commitID =document.getElementsByClassName('js-permalink-shortcut')[0].href.split("/")[6]
	}
	table = "<div class='column one-fourth2 codesearch-aside' id='toRemove'> <nav class='menu' data-pjax=''> <a role='button' id='seeDefs' class='menu-item'><svg aria-hidden='true' class='octicon octicon-code' height='16' role='img' version='1.1' viewBox='0 0 14 16' width='14'><path d='M9.5 3l-1.5 1.5 3.5 3.5L8 11.5l1.5 1.5 4.5-5L9.5 3zM4.5 3L0 8l4.5 5 1.5-1.5L2.5 8l3.5-3.5L4.5 3z'></path></svg>Code<span class='counter' id='codeCounter'></span></a><a role ='button' id='seeText' class='menu-item'><img id='t' src="+t.src+">Text<span class='counter' id='textCounter'></span></a> </nav></div><div class='column three-fourths2 codesearch-results' id='toRemove' style='float:right'><div class='repository-content' id='toRemove'> <div class='breadcrumb flex-table'> <div class='flex-table-item'> <span class='bold'><a href=/"+user+"/"+repo+">"+repo+"</a></span> / </div> <input type='text' name='query' autocomplete='off' spellcheck='false' autofocus='' id='tree-finder-field2' data-results='tree-finder-results' style='width:99%' class='tree-finder-input2' role='search' js-tree-finder-field js-navigation-enable flex-table-item-primary'><div class='flex-table-item'><button id='sg-search-submit-button' class='btn btn-sm empty-icon right js-show-file-finder' type='submit' tabindex='3'>Search</button></div></div><div id='loadingDiv' style='display:none'>Searching...</div ><div class='tree-finder clearfix' data-pjax=''><div class='flash-messages js-notice'> <div class='flash' ><form accept-charset='UTF-8' action='/sourcegraph/go-git/dismiss-tree-finder-help' class='flash-close js-notice-dismiss' data-form-nonce='9e84d03d8bcc6640b285af494d66a530ef543a51' data-remote='true' method='post'><div style='margin:0;padding:0;display:inline'><input name='utf8' type='hidden' value='✓'><input name='authenticity_token' type='hidden' value='mP8EUglfiCcfAl1tEEOFKkAhyNAQG/mxzCkwmqqhKapITZjnk06XW6lB6kmSxo6NLU6sI+cwDHdqrUlZiewlBA=='></div> </form> Type in a query and press <kbd>enter</kbd> to view results.  Press <kbd>esc</kbd> to exit. </br>  Powered by <a href='https://sourcegraph.com'>Sourcegraph</a>. </div> </div> <table id='tree-finder-results2' class='tree-browser css-truncate' cellpadding='0' cellspacing='0' width='100%' style='border-bottom:1px solid #;cacaca;width:100%'> <tbody class='tree-browser-result-template js-tree-browser-result-template'> <tr class='js-navigation-item tree-browser-result'><td> <a class='css-truncate-target js-navigation-open js-tree-finder-path'></a></td> </tr> </tbody> </table></div>"; 

	var pageScript = document.createElement('script');
	pageScript.innerHTML = '$(document).on("pjax:success", function () { var evt = new Event("PJAX_PUSH_STATE_0923"); document.dispatchEvent(evt); });';
	document.querySelector('body').appendChild(pageScript);
	document.addEventListener('PJAX_PUSH_STATE_0923', function() {
		if (repoIsGo===true) {
			addSearchButton();
		}
	})

});


$(window).on('popstate', function(){
	try{document.getElementById('sg-search-button-container').addEventListener("click", buttonClick)}catch(err){};
	try{document.getElementById('sg-search-submit-button').addEventListener("click", clickSubmitButton)}catch(err){};
	try{document.getElementById('seeText').addEventListener("click", showtext)}catch(err){};
	try{document.getElementById('seeDefs').addEventListener("click", showdefs)}catch(err){};
});


$(document).ready(function(){
	var currentURL = document.URL;
	var splitURL = currentURL.split('/')
	user = splitURL[3]
	repo = splitURL[4]
	if (((document).getElementsByClassName('entry-title')).length !== 0) {
		checkLanguageAjax(user, repo);	
	}
		
});


//Checks if the language of the repository is Go  
function checkLanguageAjax(user, repo){
	checkLang = $.ajax ({
		method: "GET",
		url: "https://api.github.com/repos/"+user+"/"+repo+"/languages"
	}).done(function(e){
		//console.log(e);
		if (e["Go"]) {
			addSearchButton();
			repoIsGo = true;
			return;		
		}
	});
	checkLang.fail(function(e){
		repoIsGo=true;
		return;
	})
	repoIsGo = false;
	return;
}




//insert search button
function addSearchButton (){
	var buttonHeader = document.querySelector('ul.pagehead-actions');
	var sgButton = buttonHeader.querySelector('#sg-search-button-container');
	if (!sgButton) {
		sgButton = document.createElement('li');
		sgButton.id = 'sg-search-button-container';
		buttonHeader.insertBefore(sgButton, buttonHeader.firstChild);
	}
	sgButton.innerHTML = "<a id='sg-search-button' class='btn btn-sm minibutton sg-button tooltipped tooltipped-s' aria-label='Find code definitions in this repository.\nKeyboard shortcut: shift-T'><img id='searchlogo' src="+search.src+" style='vertical-align:text-top' height='14' width='14'> Search code</a>";
	document.getElementById('sg-search-button-container').addEventListener("click", buttonClick);
}

//handler when search button is clicked
function buttonClick(){
	countScrolls=1;

	//store value of current page
	if ($('.container.new-discussion-timeline').not(':has(#toRemove)')) {
		original = $('.container.new-discussion-timeline').children().html();
	}

	if (document.getElementById('toRemove')) {
		$('div').remove("#toRemove");
	}

	
	//hide current page, show search bar 
	$('.container.new-discussion-timeline').children().hide();
	$('.container.new-discussion-timeline').append(table);
	/*
	if (getDefs !== undefined) {
		getDefs.abort();
	}*/

	//delay before focusing on search bar so T doesn't show up
	setTimeout(function(){
		$('.tree-finder-input2:last').focus();
	}, 1);
	
	current='';
	document.getElementById('sg-search-button-container').addEventListener("click", buttonClick);
	document.getElementById('sg-search-submit-button').addEventListener("click", clickSubmitButton);
	//document.getElementById('seeText').addEventListener("click", showtext);
	document.getElementById('seeDefs').addEventListener("click", showdefs);
	$('#seeDefs:last').addClass(' selected');

}


//handler for clicking submit button
function clickSubmitButton(){
	//table that replaces existing one during a search (does not include search bar)
	var table2 = "<table id='tree-finder-results2' class='tree-browser css-truncate' cellpadding='0' cellspacing='0' style='border-bottom:1px solid #;cacaca'> <tbody class='tree-browser-result-template js-tree-browser-result-template' aria-hidden='true'> <tr class='js-navigation-item tree-browser-result'><td> <a class='css-truncate-target js-navigation-open js-tree-finder-path'></a> </td> </tr> </tbody> </table>";
	
	query = $('.tree-finder-input2:last').val();

	//condition because we don't want to replace table if the query is the same and enter is pressed
	if (current !== query ) {
		$('.flash-messages').remove();
		
		//add logo if not already present
		if (document.getElementById('logo')===null || (!(document.getElementById('logo').offsetWidth >0) && !(document.getElementById.offsetHeight >0))) {
  			$('.tree-finder.clearfix:last').after("<div  width='100%' align='right' class='logodiv'><a href=http://sourcegraph.com><img id='logo' src="+logo2.src+" style='opacity:0.6;'></a></div>");
		}
		
		(treefinderarray[treefinderarray.length-1]).innerHTML=table2;
	    
	    if (getDefs !== undefined) {
	    	getDefs.abort();
	    }
	    if (getText !== undefined) {
			getText.abort();
		}
		try{$('.nomatch').remove();}catch(err){}
		ajaxCall();
	}    
	$('.tree-finder-input2:last').focus();
	current=query;

	/*amplitude.logEvent('SEARCH');*/
}





//events for key presses: get search screen when shift+t, submit + get request when enter, go back to previous page when esc 
document.addEventListener('keydown', keyboardevents);
var treefinderarray = document.getElementsByClassName('tree-finder');
function keyboardevents (e) {
	if (e.which===84 && e.shiftKey && (e.target.tagName.toLowerCase()) !=='input' && (e.target.tagName.toLowerCase())!=='textarea') {
		if (repoIsGo){

			countScrolls=1;

			if ($('.container.new-discussion-timeline').not(':has(#toRemove)')) {
				original = $('.container.new-discussion-timeline').children().html();
			}

			if (document.getElementById('toRemove')) {
				$('div').remove("#toRemove");
			}


		//hide current page, show search bar 
		$('.container.new-discussion-timeline').children().hide();
		$('.container.new-discussion-timeline').append(table);
		
		if (getDefs !== undefined) {
			getDefs.abort();
		}
		if (getText !== undefined) {
			getText.abort();
		}

		//delay before focusing on search bar so T doesn't show up
		setTimeout(function(){
			$('.tree-finder-input2:last').focus();
		}, 1);
		
		//set default to definition
		$('#seeDefs:last').addClass(' selected');

		current='';
		try{document.getElementById('sg-search-button-container').addEventListener("click", buttonClick);}catch(err){};
		try{document.getElementById('sg-search-submit-button').addEventListener("click", clickSubmitButton)}catch(err){};
		//document.getElementById('seeText').addEventListener("click", showtext);
		//document.getElementById('seeDefs').addEventListener("click", showdefs);
	}
}


	//press enter key to submit
	else if (e.which===13 && (e.target.tagName.toLowerCase())==='input') {
		e.stopImmediatePropagation();
		countScrolls=1;
    	//table that replaces existing one during a search (does not include search bar)
    	var table2 = "<table id='tree-finder-results2' class='tree-browser css-truncate' cellpadding='0' cellspacing='0' style='border-bottom:1px solid #;cacaca'> <tbody class='tree-browser-result-template js-tree-browser-result-template' aria-hidden='true'> <tr class='js-navigation-item tree-browser-result'><td> <a class='css-truncate-target js-navigation-open js-tree-finder-path'></a> </td> </tr> </tbody> </table>";
		
		query = $('.tree-finder-input2:last').val();

    	//condition because we don't want to replace table if the query is the same and enter is pressed
    	if (current !== query && query !== '') {
    		console.log(current);
    		console.log(query);
    		$('.flash-messages').remove();
    		
    		//add logo if not already present
    		if (document.getElementById('logo')===null || (!(document.getElementById('logo').offsetWidth >0) && !(document.getElementById.offsetHeight >0))) {
      			$('.tree-finder.clearfix:last').after("<div  width='100%' align='right' class='logodiv'><a href=http://sourcegraph.com><img id='logo' src="+logo2.src+" style='opacity:0.6;'></a></div>");
    		}
    		
    		(treefinderarray[treefinderarray.length-1]).innerHTML=table2;
		    if (getDefs !== undefined) {
		    	getDefs.abort();
		    }
		    if (getText !== undefined) {
				getText.abort();
			}
			try{$('.nomatch').remove();}catch(err){}

			ajaxCall();

    	}  


    	current=query;

		/*amplitude.logEvent('SEARCH');*/

	}

	//Press esc to hide
	else if (e.keyCode === 27) {
		$('div').remove("#toRemove");
		$('.repository-content').show();
	}

};


//removes loading div when searching for code
function removeDefLoadingDiv(){
	if ($('#seeDefs').hasClass('selected')){
		document.getElementById("loadingDiv").style.display='none'
	}
}

//removes loading div when searching for text
function removeTextLoadingDiv(){
	if ($('#seeText').hasClass('selected')){
		document.getElementById("loadingDiv").style.display='none'
	}
}

//Get request to Sourcegraph API based on current user/repo/branch
function ajaxCall() {
	document.getElementById("loadingDiv").style.display='block'
	getDefs = $.ajax ({
		method: "GET",
		url: "https://sourcegraph.com/.api/.defs?RepoRevs=github.com%2F"+user+"%2F"+repo+"@"+commitID+"&Nonlocal=true&Query="+query
		//url: "https://sourcegraph.com/.ui/github.com/"+user+"/"+repo+"@"+branch+"/.search/tokens?q="+query+"&PerPage=1000&Page=1"
	}).done(removeDefLoadingDiv, showDefResults);
	getDefs.fail(function(jqXHR, textStatus, errorThrown) {
		console.log (textStatus);
		console.log (errorThrown);
		if (textStatus!=='abort'){
			removeDefLoadingDiv();
			if (errorThrown == "Unauthorized"){
				nomatch ="<div class='nomatch'><p style='text-align:center;font-size:16px'><b> 401 (Unauthorized)</b></br></p><p style='text-align:center;font-size:12px'> You must be signed in on <a href='https://sourcegraph.com'>sourcegraph.com</a> to search private code.</p></div>";
			}
			if (errorThrown == "Not Found"){
				nomatch ="<div class='nomatch'><p style='text-align:center;font-size:16px'><b> 404 (Not Found)</b></br>This repository has not been analyzed and built by Sourcegraph yet.</br> Currently, Go repositories are supported.  More language support will be rolled out soon, stay tuned.</p></div>";
				if (repoIsGo===true){
					nomatch ="<div class='nomatch' id='404nomatch'><p style='text-align:center;font-size:16px'><b> This repository has not been analyzed by Sourcegraph yet.</br> Click the link below to activate search on this repository: <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"' target='none'>sourcegraph.com/github.com/"+user+"/"+repo+"</a> </b></p></div>"
				}
			}
			removeDefLoadingDiv();
			$('.tree-finder.clearfix:last').after(nomatch);
		}
	});
	/*
	getText = $.ajax ({
		method: "GET",
		url: "https://sourcegraph.com/.ui/github.com/"+user+"/"+repo+"@"+branch+"/.search/text?q="+query+"&PerPage=10&Page=1"
	}).done(showTextResults, removeTextLoadingDiv);
	getText.fail(function(jqXHR, textStatus, errorThrown) {
		console.log (textStatus);
		console.log (errorThrown);
		if (textStatus!=='abort'){
			removeTextLoadingDiv();
			if (errorThrown == "Unauthorized"){
				nomatch ="<div class='nomatch'><p style='text-align:center;font-size:16px'><b> 401 (Unauthorized)</b></br></p><p style='text-align:center;font-size:12px'> You must be signed in on <a href='https://sourcegraph.com'>sourcegraph.com</a> to search private code.</p></div>";	
			}
			if (errorThrown == "Not Found"){
				nomatch ="<div class='nomatch' id='404nomatch'><p style='text-align:center;font-size:16px'><b> 404 (Not Found)</b></br><p style='text-align:center;font-size:12px'>This repository has not been analyzed by Sourcegraph yet.</br> Currently, Go repositories are supported.  More language support will be rolled out soon, stay tuned.</p></div>";	
				if (repoIsGo===true){
					nomatch ="<div class='nomatch' id='404nomatch'><p style='text-align:center;font-size:16px'><b> This repository has not been analyzed by Sourcegraph yet.</br> Click the link below to activate search on this repository: <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"' target='none'>sourcegraph.com/github.com/"+user+"/"+repo+"</a> </b></p></div>"
				}
				$('.tree-finder.clearfix:last').after(nomatch);
				try{$('#nodefmatch').remove()}catch(err){}
			}
		}
	});*/
}


//Iterate thru JSON object array and generate results table
function showDefResults(dataArray) {
	if ($('#seeText').hasClass('selected')){
		$('#tree-finder-results2').attr("style", "display:none");
    }
	document.getElementById('codeCounter').style.display='block';
	console.log(dataArray)

	if (!dataArray.Defs && !(document.getElementById('404nomatch'))) {		
		nomatch = "<div class='nomatch' id='nodefmatch'><p style='text-align:center;font-size:16px'><b> No matching definitions found. </br></b></p></div>";
		$('.tree-finder.clearfix:last').after(nomatch);
		if ($('#seeText').hasClass('selected')) {
			try{document.getElementById('nodefmatch').style.display='none'}catch(err){}
		}
		$('.tree-browser:last').attr("style","border-top:none;border-bottom:none;");
		document.getElementById('codeCounter').innerHTML = "0";

		return;
	}

	document.getElementById('codeCounter').innerHTML = dataArray.Total;

	for(var i =0; i<dataArray.Defs.length;i++) {
		var eachRes = dataArray.Defs[i];
		var repWideQualified = eachRes.FmtStrings.Type.RepositoryWideQualified;
		if (repWideQualified === undefined) {
			repWideQualified = ''; 
		}
		var strToReturn = "<span style=color:#4078C0>" + eachRes.FmtStrings.Name.ScopeQualified + "</span>" + eachRes.FmtStrings.Type.ScopeQualified;
		console.log(strToReturn)
		var hrefurl = "https://sourcegraph.com/"+eachRes.Repo+"/.GoPackage/"+eachRes.Repo+"/.def/"+eachRes.Path;

		if (i !== dataArray.Defs.length-1) { 
			$('.tree-browser:last tbody:last').after("<tbody class='js-tree-finder-results'><tr id='searchrow' class='js-navigation-item tree-browser-result' style='border-bottom: 1px solid rgb(238, 238, 238);'><td class='icon' width='21px'><svg aria-hidden='true' class='octicon octicon-chevron-right' height='16' role='img' version='1.1' viewBox='0 0 8 16' width='8'><path d='M7.5 8L2.5 13l-1.5-1.5 3.75-3.5L1 4.5l1.5-1.5 5 5z'></path></svg></td><td class='icon'><svg aria-hidden='true' class='octicon octicon-file-text' height='16' role='img' version='1.1' viewBox='0 0 12 16' width='12'><path d='M6 5H2v-1h4v1zM2 8h7v-1H2v1z m0 2h7v-1H2v1z m0 2h7v-1H2v1z m10-7.5v9.5c0 0.55-0.45 1-1 1H1c-0.55 0-1-0.45-1-1V2c0-0.55 0.45-1 1-1h7.5l3.5 3.5z m-1 0.5L8 2H1v12h10V5z'></path></svg></td></td><td><a href="+hrefurl+" target='blank'><span class='res'>"+eachRes.Kind+ " "+ strToReturn + "</span></a></td></tr></tbody>");
		}
		else {
			$('.tree-browser:last tbody:last').after("<tbody class='js-tree-finder-results'><tr id='searchrow' class='js-navigation-item tree-browser-result'><td id='icon' style='width:21px;padding-left:10px'><svg aria-hidden='true' class='octicon octicon-chevron-right' height='16' role='img' version='1.1' viewBox='0 0 8 16' width='8'><path d='M7.5 8L2.5 13l-1.5-1.5 3.75-3.5L1 4.5l1.5-1.5 5 5z'></path></svg></td><td class='icon'><svg aria-hidden='true' class='octicon octicon-file-text' height='16' role='img' version='1.1' viewBox='0 0 12 16' width='12'><path d='M6 5H2v-1h4v1zM2 8h7v-1H2v1z m0 2h7v-1H2v1z m0 2h7v-1H2v1z m10-7.5v9.5c0 0.55-0.45 1-1 1H1c-0.55 0-1-0.45-1-1V2c0-0.55 0.45-1 1-1h7.5l3.5 3.5z m-1 0.5L8 2H1v12h10V5z'></path></svg></td></td><td><a href="+hrefurl+" target='blank'><span class='res'>"+eachRes.Kind + " " + strToReturn + "</span></a></td></tr></tbody>");
		}

	}
}	



/* --------------------------------------------Text search --------------------------------------------------------------*/
//https://sourcegraph.com/.ui/github.com/attfarhan/mux@master/.search/text?q=route&PerPage=10&Page=1
/*
//show text results
function showTextResults(dataArray){

	document.getElementById('textCounter').style.display='block';
	document.getElementById('textCounter').innerHTML = dataArray.Total;
			
	
	var codelist = "<div class='code-list' id=codelist style='margin-top:10px;'></div>"; 

	prevFile =''; 
	$('.tree-finder.clearfix:last').append(codelist);
	
	if (dataArray.Results.length===0) {
			nomatch = "<div class='nomatch' ><p style='text-align:center;font-size:16px'><b> No matching text found. </br></b></p></div>";
			$('#codelist').append(nomatch);
	}

	if ($('#seeText').hasClass('selected')){
    	document.getElementById('codelist').style.display = 'block';
		$('#tree-finder-results2').attr("style", "display:none");
    }

	for (var i = 0; i < dataArray.Results.length; i++) {
		var filetypesplit = (dataArray.Results[i].File).split(".");
		var filename = dataArray.Results[i].File; 
		var filetype = filetypesplit[filetypesplit.length-1];
		var startLine = dataArray.Results[i].StartLine;
		var endLine = dataArray.Results[i].EndLine;
		var lineNumber = startLine;
		var content = dataArray.Results[i].Contents;
		var match = query;
		var regexp = new RegExp (match, 'g');
		var toEnter = content.replace(regexp, "<span class='match'>"+query+"</span>");
		var contentArray = toEnter.split("\n");		

		if (filename!==prevFile){
			$('.code-list').append("<div class='code-list-item code-list-item-public repo-specific'> <span class='language'>" +filetype+ "</span> <p class='title'><a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"#L"+lineNumber+"' title='"+filename+"' target='none'>"+filename+"</a><br><span class='text-small text-muted match-count'> Results in "+filename+"</span></p><div class='file-box blob-wrapper'><table><tbody class='textres'></tbody>"); for(var j = 0; j < contentArray.length; j++) {$('.textres:last').append(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"#L"+lineNumber+"' target='none'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[j]+" </td> </tr>")
			//href=/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"'
			lineNumber++;
			}
		}
		else {
			$('tr:last').after("<tr class='divider'> <td class='blob-num'>…</td> <td class='blob-code'></td> </tr>");
			for(var k = 0; k < contentArray.length; k++) {
				$('tr:last').after(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"#L"+lineNumber+"' target='none'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[k]+" </td> </tr>")
				lineNumber++;
			}
		}
		prevFile = filename;	
	}
	var countScrolls=1;
	document.addEventListener('scroll', getInfiniteResults, true);
	if (! ($("body").height() > $(window).height()) && $('#seeText').hasClass('selected')) {
		countScrolls++;		
		subsequentAjaxCalls(countScrolls);
	}
}



function getInfiniteResults(){
	if (document.body.scrollHeight === document.body.scrollTop + window.innerHeight && $('#seeText').hasClass('selected')){
		countScrolls++;		
		subsequentAjaxCalls(countScrolls);
	}

}

//show text results 
function showtext(){
	if (document.getElementById('codelist')!==null){
		document.getElementById('codelist').style.display = 'block';
		$('#tree-finder-results2').attr("style", "display:none");
		try{$('#nodefmatch').attr("style","display:none")}catch(err){}
		document.getElementById('seeText').className += ' selected';
		document.getElementById('seeDefs').className = 'menu-item';
	}
	else{
		document.getElementById('seeText').className += ' selected';
		document.getElementById('seeDefs').className = 'menu-item';
	}
}

//show definition results
function showdefs(){
	if (document.getElementById('codelist')!==null){
		try{document.getElementById('nodefmatch').style.display='block'}catch(err){}
		document.getElementById('codelist').style.display = 'none';
		document.getElementById('tree-finder-results2').style.display = '';
		document.getElementById('seeDefs').className += ' selected';
		document.getElementById('seeText').className = 'menu-item';
	}
	else{
		document.getElementById('seeDefs').className += ' selected';
		document.getElementById('seeText').className = 'menu-item';
	}

}


function subsequentAjaxCalls(numScrolls) {
	document.removeEventListener('scroll', getInfiniteResults, true);

	if (!document.getElementById('load-more-results')){
		$('.tree-finder.clearfix:last').append("<p style='text-align:center;font-size:16px' id='load-more-results'><b> Loading more results... </b></p>")
	}
	getInfiniteText = $.ajax ({
		method: "GET", 
		url: "https://sourcegraph.com/.ui/github.com/"+user+"/"+repo+"@"+branch+"/.search/text?q="+query+"&PerPage=10&Page="+numScrolls
	}).done(removeLoading, infiniteTextResults)
	getInfiniteText.fail( function() {
		removeLoading();
		if (!document.getElementById('all-results')){
			$('.code-list').append("<p style='text-align:center;font-size:16px' id='all-results'><b> All results shown. </b></p>");
		}
		document.removeEventListener('scroll', getInfiniteResults)
	})	
}


function removeLoading(){
	$('#load-more-results').remove();
}


function infiniteTextResults(dataArray){
	if ($('#seeText').hasClass('selected')){
    	document.getElementById('codelist').style.display = 'block';
		$('#tree-finder-results2').attr("style", "display:none");
    }

	for (var i = 0; i < dataArray.Results.length; i++) {
		var filename = dataArray.Results[i].File; 
		if (prevFile !== filename) {
			$('.tree-finder.clearfix:last').append("<div class='code-list' id='codelist></div>");
		}
		var filetypesplit = (dataArray.Results[i].File).split(".");
		var filetype = filetypesplit[filetypesplit.length-1];
		var startLine = dataArray.Results[i].StartLine;
		var endLine = dataArray.Results[i].EndLine;
		var lineNumber = startLine;
		var content = dataArray.Results[i].Contents;
		var match = query;
		var regexp = new RegExp (match, 'g');
		var toEnter = content.replace(regexp, "<span class='match'>"+query+"</span>");
		var contentArray = toEnter.split("\n");



		if (filename!==prevFile){
			$('.code-list').append("<div class='code-list-item code-list-item-public repo-specific'> <span class='language'>" +filetype+ "</span> <p class='title'><a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"#L"+lineNumber+"' title='"+filename+"' target='none'>"+filename+"</a><br><span class='text-small text-muted match-count'> Results in "+filename+"</span></p><div class='file-box blob-wrapper'><table><tbody class='textres'></tbody>"); for(var j = 0; j < contentArray.length; j++) {$('.textres:last').append(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"#L"+lineNumber+"' target='none'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[j]+" </td> </tr>")
			lineNumber++;
			}
		}
		else {
			$('tr:last').after("<tr class='divider'> <td class='blob-num'>…</td> <td class='blob-code'></td> </tr>");
			for(var k = 0; k < contentArray.length; k++) {
				$('tr:last').after(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"#L"+lineNumber+"' target='none'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[k]+" </td> </tr>")
				lineNumber++;
			}
		}
		prevFile = filename;	
	}
	document.addEventListener('scroll', getInfiniteResults, true);

}

/*
//amplitude tracking
document.addEventListener('DOMContentLoaded', function(){
	(function(e,t){var n=e.amplitude||{_q:[],_iq:{}};var r=t.createElement("script");r.type="text/javascript";
	r.async=true;r.src="https://d24n15hnbwhuhn.cloudfront.net/libs/amplitude-2.9.0-min.gz.js";
	r.onload=function(){e.amplitude.runQueuedFunctions()};var i=t.getElementsByTagName("script")[0];
	i.parentNode.insertBefore(r,i);var s=function(){this._q=[];return this};function a(e){
	s.prototype[e]=function(){this._q.push([e].concat(Array.prototype.slice.call(arguments,0)));
	return this}}var o=["add","append","clearAll","set","setOnce","unset"];for(var c=0;c<o.length;c++){
	a(o[c])}n.Identify=s;var u=["init","logEvent","logRevenue","setUserId","setUserProperties","setOptOut","setVersionName","setDomain","setDeviceId","setGlobalUserProperties","identify","clearUserProperties"];
	function l(e){function t(t){e[t]=function(){e._q.push([t].concat(Array.prototype.slice.call(arguments,0)));
	}}for(var n=0;n<u.length;n++){t(u[n])}}l(n);n.getInstance=function(e){e=(!e||e.length===0?"$default_instance":e).toLowerCase();
	if(!n._iq.hasOwnProperty(e)){n._iq[e]={_q:[]};l(n._iq[e])}return n._iq[e]};e.amplitude=n;
	})(window, document);
	amplitude.init("f7491eae081c8c94baf15838b0166c63");
})	
*/
