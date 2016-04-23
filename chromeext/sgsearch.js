//This file adds code and text search into GitHub
(function() {

	var user, urlsplit, repo, branch, path, query,
	searchPage, nomatch, countScrolls,
	getDefs, getText, prevFile, commitID,
	repoIsGo = false, token, current = '';

	//Branding and icons
	var logo2 = document.createElement('img');
	logo2.src = chrome.extension.getURL("assets/src4.png");
	var search = document.createElement('img');
	search.src = chrome.extension.getURL("assets/search.png");
	var t = document.createElement('img')
	t.src = chrome.extension.getURL ("assets/t.png")

	function checkFile(document){
		if (document.getElementsByClassName('file').length!==0 || document.getElementsByClassName('repohead-details-container').length !== 0) {
			main();
		}
		else {
			setTimeout(function(){checkFile(document)}, 200)
		}
	}

	checkFile(document)


	function main(){
		//Repository properties
		urlsplit = document.URL.split("/");
		user = urlsplit[3];
		repo = urlsplit[4];
		branch = 'master';
		if (urlsplit[6] !== null && (urlsplit[5]==='tree' || urlsplit[5]==='blob')) {
			branch = urlsplit[6];
		}
		path = urlsplit[7];
		if (urlsplit.length > 8){
			for (var i = 8; i < urlsplit.length; i++){
				path = path + "/" + urlsplit[i];
			}
		}

		//Check language if public repo through GitHub API
		if (((document).getElementsByClassName('entry-title')).length !== 0 && document.getElementsByClassName('vis-private').length ===0){
			checkLanguageAjax(user, repo);
		}
		//Check language if private repo
		if (document.getElementsByClassName('vis-private').length !==0) {
			checkLanguageNotAjax();
		}

		//Get CommitID file page
		if (document.getElementsByClassName('file').length !== 0) {
			commitID =document.getElementsByClassName('js-permalink-shortcut')[0].href.split("/")[6]
		}
		//Get CommitID if repo page
		else if (document.getElementsByClassName('entry-title').length !== 0) {
			if (document.getElementsByClassName('commit-tease-sha')[0]){
				commitID=document.getElementsByClassName('commit-tease-sha')[0].href.split("/")[6];
			}
			else{
				//Set commitID on repo page in cases where we call this before body the element is loaded.
				//Often occurs on a newly forked repo.
				function getCommitID(){
					if (document.getElementsByClassName('commit-tease-sha')[0]) {
						return document.getElementsByClassName('commit-tease-sha')[0].href.split("/")[6];
					}
					else {
						commitID = setTimeout(function(){getCommitID()});
					}
				}
			}
		}

		//Search page elements when first pressing Shift+T
		searchPage = "<div class='column one-fourth2 codesearch-aside' id='toRemove'> <nav class='menu' data-pjax=''> <a role='button' id='seeDefs' class='menu-item'><svg aria-hidden='true' class='octicon octicon-code' height='16' role='img' version='1.1' viewBox='0 0 14 16' width='14'><path d='M9.5 3l-1.5 1.5 3.5 3.5L8 11.5l1.5 1.5 4.5-5L9.5 3zM4.5 3L0 8l4.5 5 1.5-1.5L2.5 8l3.5-3.5L4.5 3z'></path></svg>Code<span class='counter' id='codeCounter'></span></a><a role ='button' id='seeText' class='menu-item'><img id='t' src="+t.src+">Text<span class='counter' id='textCounter'></span></a> </nav></div><div class='column three-fourths2 codesearch-results' id='toRemove' style='float:right'><div class='repository-content' id='toRemove'> <div class='breadcrumb flex-table'> <div class='flex-table-item'> <span class='bold'><a href=/"+user+"/"+repo+">"+repo+"</a></span> / </div> <input type='text' name='query' autocomplete='off' spellcheck='false' autofocus='' id='tree-finder-field2' data-results='tree-finder-results' style='width:99%' class='tree-finder-input2' role='search' js-tree-finder-field js-navigation-enable flex-table-item-primary'><div class='flex-table-item'><button id='sg-search-submit-button' class='btn btn-sm empty-icon right js-show-file-finder' type='submit' tabindex='3'>Search</button></div></div><div id='loadingDiv' style='display:none'>Searching...</div ><div class='tree-finder clearfix' data-pjax=''><div class='flash-messages js-notice'> <div class='flash' ><form accept-charset='UTF-8' action='/sourcegraph/go-git/dismiss-tree-finder-help' class='flash-close js-notice-dismiss' data-form-nonce='9e84d03d8bcc6640b285af494d66a530ef543a51' data-remote='true' method='post'><div style='margin:0;padding:0;display:inline'><input name='utf8' type='hidden' value='✓'><input name='authenticity_token' type='hidden' value='mP8EUglfiCcfAl1tEEOFKkAhyNAQG/mxzCkwmqqhKapITZjnk06XW6lB6kmSxo6NLU6sI+cwDHdqrUlZiewlBA=='></div> </form> Type in a query and press <kbd>enter</kbd> to view results.  Press <kbd>esc</kbd> to exit. </br>  Powered by <a href='https://sourcegraph.com'>Sourcegraph</a>. </div> </div> <table id='tree-finder-results2' class='tree-browser css-truncate' cellpadding='0' cellspacing='0' width='100%' style='border-bottom:1px solid #;cacaca;width:100%'> <tbody class='tree-browser-result-template js-tree-browser-result-template'> <tr class='js-navigation-item tree-browser-result'><td> <a class='css-truncate-target js-navigation-open js-tree-finder-path'></a></td> </tr> </tbody> </table></div>";

		//Listen to GitHub pjax event when navigating b/t pages
		var pageScript = document.createElement('script');
		pageScript.innerHTML = '$(document).on("pjax:success", function () { var evt = new Event("PJAX_PUSH_STATE_0923"); document.dispatchEvent(evt); });';
		document.querySelector('body').appendChild(pageScript);

		//Check if language is Go
		document.addEventListener('PJAX_PUSH_STATE_0923', function() {
			if (document.getElementsByClassName('vis-private').length !== 0) {
				debounce(checkLanguageNotAjax());
			}
			else if (repoIsGo || document.getElementsByClassName('vis-private').length !==0) {
				addSearchButton();

			}
		})

		//Event listener for handling case when back button is used
		window.addEventListener('onpopstate', handleBackButton);

		//Event listener for keyboard shortcut
		document.addEventListener('keydown', keyboardevents);

	}


	//Make sure search events still work if user is using back/forward
	function handleBackButton(){
		try{document.getElementById('sg-search-button-container').addEventListener("click", clickSearchButton)}catch(err){}
		try{document.getElementById('sg-search-submit-button').addEventListener("click", clickSubmitButton)}catch(err){}
		try{document.getElementById('seeText').addEventListener("click", showtext)}catch(err){}
		try{document.getElementById('seeDefs').addEventListener("click", showdefs)}catch(err){}
	}


	//Check language not using GitHub API - good for navigating within the repo so we don't hit rate limit.
	function checkLanguageNotAjax(){
			//If it is a file page, we can get the language by looking at the file extension
			var fileElem = document.querySelector('.file .blob-wrapper')
			var repoElem = document.getElementsByClassName('repository-lang-stats')
			if (fileElem){
				var finalPath = document.getElementsByClassName('final-path')[0].innerText.split('.')
				lang = finalPath[finalPath.length-1]
				if (lang.toLowerCase() === "go"){
					addSearchButton();
					repoIsGo=true;
					return;
				}
			}
			else if (repoElem.length !== 0) {
				var langList = repoElem[0].innerText.split(' ')
				if (langList.indexOf("Go")>-1) {
					addSearchButton();
					repoIsGo=true;
					return;
				}
			}


		}

	//Checks if the language of the repository is Go using GitHub API
	function checkLanguageAjax(user, repo){
		checkLang = $.ajax ({
			method: "GET",
			url: "https://api.github.com/repos/"+user+"/"+repo+"/languages"
		}).done(function(e){
			if (e["Go"]) {
				addSearchButton();
				repoIsGo = true;
				return;
			}
		});
		checkLang.fail(checkLanguageNotAjax)
	}


	//Insert search button
	function addSearchButton (){
		var buttonHeader = document.querySelector('ul.pagehead-actions');
		var sgButton = buttonHeader.querySelector('#sg-search-button-container');
		if (!sgButton) {
			sgButton = document.createElement('li');
			sgButton.id = 'sg-search-button-container';
			buttonHeader.insertBefore(sgButton, buttonHeader.firstChild);
		}
		sgButton.innerHTML = "<a id='sg-search-button' class='btn btn-sm minibutton sg-button tooltipped tooltipped-s' aria-label='Find code definitions in this repository.\nKeyboard shortcut: shift-T'><img id='searchlogo' src="+search.src+" style='vertical-align:text-top' height='14' width='14'> Search code</a>";
		document.getElementById('sg-search-button-container').addEventListener("click", clickSearchButton);
	}

	//Handler when search button is clicked
	function clickSearchButton(){
		countScrolls=1;

		amplitude.logEvent('ClickSearchCodeButton')

		//Store value of current page
		if ($('.container.new-discussion-timeline').not(':has(#toRemove)')) {
			original = $('.container.new-discussion-timeline').children().html();
		}

		if (document.getElementById('toRemove')) {
			$('div').remove("#toRemove");
		}


		//Hide current page, show search bar
		$('.container.new-discussion-timeline').children().hide();
		$('.container.new-discussion-timeline').append(searchPage);

		//Abort old call if they reset search
		if (getDefs !== undefined) {
			getDefs.abort();
		}

		//Delay before focusing on search bar so T doesn't show up
		setTimeout(function(){
			$('.tree-finder-input2:last').focus();
		}, 1);

		current='';
		document.getElementById('sg-search-button-container').addEventListener("click", clickSearchButton);
		document.getElementById('sg-search-submit-button').addEventListener("click", clickSubmitButton);
		document.getElementById('seeText').addEventListener("click", showtext);
		document.getElementById('seeDefs').addEventListener("click", showdefs);
		$('#seeDefs:last').addClass(' selected');

	}


	//Handler for clicking submit button
	function clickSubmitButton(){
		var treefinderarray = document.getElementsByClassName('tree-finder');
		amplitude.logEvent('Search');

		//Table that replaces existing one during a search (does not include search bar)
		var newSearchPage = "<table id='tree-finder-results2' class='tree-browser css-truncate' cellpadding='0' cellspacing='0' style='border-bottom:1px solid #;cacaca'> <tbody class='tree-browser-result-template js-tree-browser-result-template' aria-hidden='true'> <tr class='js-navigation-item tree-browser-result'><td> <a class='css-truncate-target js-navigation-open js-tree-finder-path'></a> </td> </tr> </tbody> </table>";

		query = $('.tree-finder-input2:last').val();

		//Condition because we don't want to request again if the query is the same
		if (current !== query) {
			$('.flash-messages').remove();

			//Add logo if not already present
			if (document.getElementById('logo')===null || (!(document.getElementById('logo').offsetWidth >0) && !(document.getElementById.offsetHeight >0))) {
				$('.tree-finder.clearfix:last').after("<div  width='100%' align='right' class='logodiv'><a href=http://sourcegraph.com><img id='logo' src="+logo2.src+" style='opacity:0.6;'></a></div>");
			}

			(treefinderarray[treefinderarray.length-1]).innerHTML=newSearchPage;

			//Abort old calls if they search again before old request is finished
			if (getDefs !== undefined) {
				getDefs.abort();
			}
			if (getText !== undefined) {
				getText.abort();
			}

			try{$('.nomatch').remove();}catch(err){}

			if (document.getElementsByClassName('vis-private').length !==0){
				getAuthToken();
			}

			else {
				setRev();
			}
		}
		$('.tree-finder-input2:last').focus();
		current=query;

	}


	//Events for key presses: get search screen when Shift+t, submit when Enter, go back to previous page when esc
	function keyboardevents (e) {
		var treefinderarray = document.getElementsByClassName('tree-finder');
		//Shift+T
		if (e.which===84 && e.shiftKey && (e.target.tagName.toLowerCase()) !=='input' && (e.target.tagName.toLowerCase())!=='textarea') {
			if (repoIsGo || document.getElementsByClassName('vis-private').length !==0){
				amplitude.logEvent('KbdShortcut')

				countScrolls=1;

				if ($('.container.new-discussion-timeline').not(':has(#toRemove)')) {
					original = $('.container.new-discussion-timeline').children().html();
				}

				if (document.getElementById('toRemove')) {
					$('div').remove("#toRemove");
				}

				//Hide current page, show search bar
				$('.container.new-discussion-timeline').children().hide();
				$('.container.new-discussion-timeline').append(searchPage);

				if (getDefs !== undefined) {
					getDefs.abort();
				}
				if (getText !== undefined) {
					getText.abort();
				}

				//Delay before focusing on search bar so T doesn't show up
				setTimeout(function(){
					$('.tree-finder-input2:last').focus();
				}, 1);

				//Set default to def search
				$('#seeDefs:last').addClass(' selected');
				current='';

				try{document.getElementById('sg-search-button-container').addEventListener("click", clickSearchButton);}catch(err){};
				try{document.getElementById('sg-search-submit-button').addEventListener("click", clickSubmitButton)}catch(err){};
				document.getElementById('seeText').addEventListener("click", showtext);
				document.getElementById('seeDefs').addEventListener("click", showdefs);
			}
		}


		//Enter
		else if (e.which===13 && (e.target.tagName.toLowerCase())==='input' && e.target.id.toLowerCase() === 'tree-finder-field2') {
			amplitude.logEvent('Search');

			//To prevent GitHub event where pressing enter while hovering over file takes you to that file
			e.stopImmediatePropagation();
			countScrolls=1;

			//Table that replaces existing one during a search (does not include search bar)
			var newSearchPage= "<table id='tree-finder-results2' class='tree-browser css-truncate' cellpadding='0' cellspacing='0' style='border-bottom:1px solid #;cacaca'> <tbody class='tree-browser-result-template js-tree-browser-result-template' aria-hidden='true'> <tr class='js-navigation-item tree-browser-result'><td> <a class='css-truncate-target js-navigation-open js-tree-finder-path'></a> </td> </tr> </tbody> </table>";

			query = $('.tree-finder-input2:last').val();

			//Condition because we don't want to request again if the query is the same
			if (current !== query && query !== '') {
				$('.flash-messages').remove();

				if (document.getElementById('logo')===null || (!(document.getElementById('logo').offsetWidth >0) && !(document.getElementById.offsetHeight >0))) {
					$('.tree-finder.clearfix:last').after("<div  width='100%' align='right' class='logodiv'><a href=http://sourcegraph.com><img id='logo' src="+logo2.src+" style='opacity:0.6;'></a></div>");
				}
				(treefinderarray[treefinderarray.length-1]).innerHTML=newSearchPage;

				//Abort old calls if they search again before old request is finished
				if (getDefs !== undefined) {
					getDefs.abort();
				}
				if (getText !== undefined) {
					getText.abort();
				}
				try{$('.nomatch').remove();}catch(err){}


				if (document.getElementsByClassName('vis-private').length !==0){
					getAuthToken();
				}
				else {
					setRev();
				}

			}
			current=query;
		}

		//Esc
		else if (e.keyCode === 27) {
			$('div').remove("#toRemove");
			$('.repository-content').show();
		}

	}


	//Get authentication token from Sourcegraph to ensure user is logged in to Sourcegraph for private code
	function getAuthToken(){
		try{document.getElementById("loadingDiv").style.display='block'}catch(err){}
		if (document.getElementsByClassName('vis-private').length !==0){
			$.ajax ({
				method:"GET",
				url: "https://sourcegraph.com"
			}).done(authHandler)
		}
	}
	function authHandler(data) {
		var doc = (new DOMParser()).parseFromString(data,"text/xml");
		token = ("x-oauth-basic:"+doc.getElementsByTagName("head")[0].getAttribute('data-current-user-oauth2-access-token'));
		setRev(token)
	}

	function setRev(token) {
		var version = $.ajax ({
			dataType: "json",
			method: "GET",
			url: "https://sourcegraph.com/.api/repos/github.com/"+user+"/"+repo+"/-/srclib-data-version?",
			headers: {
				authorization:'Basic ' + token
			}
		}).done(function(result){
			ajaxCall(result.CommitID, token);
		})
		version.fail(function(jqXHR, textStatus, errorThrown) {
			nomatch = "<p>No matches</p>"
			if (textStatus!=='abort'){
				removeTextLoadingDiv();
				if (version.status === 401){
					nomatch ="<div class='nomatch'><p style='text-align:center;font-size:16px'><b> 401 (Unauthorized)</b></br></p><p style='text-align:center;font-size:12px'> You must be signed in on <a href='https://sourcegraph.com/login?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext'>sourcegraph.com</a> to search private code.</p></div>";
				}
				else if (version.status === 404 || version.status === 500 || version.status === 400){
					nomatch ="<div class='nomatch' id='404nomatch'><p style='text-align:center;font-size:16px'><b> This repository has not been analyzed by Sourcegraph yet.</br> Click the link below and refresh to activate search on this repository: <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext' target='blank'>sourcegraph.com/github.com/"+user+"/"+repo+"</a> </b></p></div>";
				}
				removeDefLoadingDiv();
				$('.tree-finder.clearfix:last').after(nomatch);
			}
		})
	}

	//Get def and search results.  Takes in token when used on private code.
	//Known issue: If a build is ongoing, we won't get any def results and if a build has failed, the api will return an empty object
	//Problematic because we'll show a message "No definition matches" for both, I haven't yet found a way to distinguish the build
	//failure or ongoing case and the actual no def case.  There is only a fail case for getText because getDefs seems to never return
	//an error when the repo is not built, but instead an empty list.
	function ajaxCall(rev, token){
		try{document.getElementById("loadingDiv").style.display='block'}catch(err){}
		getDefs = $.ajax ({
			method: "GET",
			url: "https://sourcegraph.com/.api/defs?RepoRevs=github.com%2F"+user+"%2F"+repo+"@"+rev+"&Nonlocal=true&Query="+query+"&PerPage=100&Page=1",
			headers: {
				'authorization': 'Basic ' + window.btoa(token)
			}
		}).done(removeDefLoadingDiv, showDefResults);
		getText = $.ajax ({
			method: "GET",
			url: "https://sourcegraph.com/.api/repos/github.com/"+user+"/"+repo+"@master==="+rev+"/-/tree-search?Query="+query+"&QueryType=fixed&N=10&ContextLines=2&Offset=0",
			headers: {
				'authorization': 'Basic ' + window.btoa(token)
			}
		}).done(removeTextLoadingDiv, showTextResults);
		getText.fail(function(jqXHR, textStatus, errorThrown) {
			nomatch = "<p>No matches</p>"
			if (textStatus!=='abort'){
				removeTextLoadingDiv();
				if (getText.status === 401){
					nomatch ="<div class='nomatch'><p style='text-align:center;font-size:16px'><b> 401 (Unauthorized)</b></br></p><p style='text-align:center;font-size:12px'> You must be signed in on <a href='https://sourcegraph.com/login?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext'>sourcegraph.com</a> to search private code.</p></div>";
				}
				else if (getText.status === 404 || getText.status === 500 || getText.status === 400){
					nomatch ="<div class='nomatch' id='404nomatch'><p style='text-align:center;font-size:16px'><b> This repository has not been analyzed by Sourcegraph yet.</br> Click the link below and refresh to activate search on this repository: <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext' target='blank'>sourcegraph.com/github.com/"+user+"/"+repo+"</a> </b></p></div>";
				}
				if (document.getElementsByClassName('nomatch').length ===0){
					$('.tree-finder.clearfix:last').after(nomatch);
				}

			}
			document.getElementById('textCounter').style.display='block';
			document.getElementById('textCounter').innerHTML = "0";
			document.getElementById('codeCounter').innerHTML = "0";

		});
	}

	//Removes loading div when searching for code
	function removeDefLoadingDiv(){
		if ($('#seeDefs').hasClass('selected')){
			document.getElementById("loadingDiv").style.display='none'
		}
	}

	//Removes loading div when searching for text
	function removeTextLoadingDiv(){
		if ($('#seeText').hasClass('selected')){
			document.getElementById("loadingDiv").style.display='none'
		}
	}

	//Generate def results table
	function showDefResults(dataArray) {
		document.addEventListener('click', function(e){
			if (e.target.className === 'res sg-res' || e.target.parentNode.className === 'res sg-res') {
				debounce(amplitude.logEvent('ViewCodeSearchResult'), 250)
			}
		})
		if ($('#seeText').hasClass('selected')){
			$('#tree-finder-results2').attr("style", "display:none");
		}
		document.getElementById('codeCounter').style.display='block';
		//console.log(dataArray)
		if (dataArray.Defs){
			document.getElementById('codeCounter').innerHTML = dataArray.Defs.length;

			for(var i =0; i<dataArray.Defs.length; i++) {
				var eachRes = dataArray.Defs[i];
				var repWideQualified = eachRes.FmtStrings.Type.RepositoryWideQualified;
				if (repWideQualified === undefined) {
					repWideQualified = '';
				}
				var strToReturn = "<span style=color:#4078C0>" + eachRes.FmtStrings.Name.ScopeQualified + "</span>" + eachRes.FmtStrings.Type.ScopeQualified;
				var hrefurl = "https://sourcegraph.com/"+eachRes.Repo+"/-/def/GoPackage/"+eachRes.Unit+"/-/"+eachRes.Path+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext";
				if (eachRes.Kind === "package"){
					hrefurl = "https://sourcegraph.com/"+eachRes.Repo+"@master==="+eachRes.CommitID+"/-/tree?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext"
				}

				if (i !== dataArray.Defs.length-1) {
					$('.tree-browser:last tbody:last').after("<tbody class='js-tree-finder-results'><tr id='searchrow' class='js-navigation-item tree-browser-result' style='border-bottom: 1px solid rgb(238, 238, 238);'><td class='icon' width='21px'><svg aria-hidden='true' class='octicon octicon-chevron-right' height='16' role='img' version='1.1' viewBox='0 0 8 16' width='8'><path d='M7.5 8L2.5 13l-1.5-1.5 3.75-3.5L1 4.5l1.5-1.5 5 5z'></path></svg></td><td class='icon'><svg aria-hidden='true' class='octicon octicon-file-text' height='16' role='img' version='1.1' viewBox='0 0 12 16' width='12'><path d='M6 5H2v-1h4v1zM2 8h7v-1H2v1z m0 2h7v-1H2v1z m0 2h7v-1H2v1z m10-7.5v9.5c0 0.55-0.45 1-1 1H1c-0.55 0-1-0.45-1-1V2c0-0.55 0.45-1 1-1h7.5l3.5 3.5z m-1 0.5L8 2H1v12h10V5z'></path></svg></td></td><td><a href="+hrefurl+" target='blank' ><span class='res sg-res' class='sgres'>"+eachRes.Kind+ " "+ strToReturn + "</span></a></td></tr></tbody>");
				}
				else {
					$('.tree-browser:last tbody:last').after("<tbody class='js-tree-finder-results'><tr id='searchrow' class='js-navigation-item tree-browser-result'><td id='icon' style='width:21px;padding-left:10px'><svg aria-hidden='true' class='octicon octicon-chevron-right' height='16' role='img' version='1.1' viewBox='0 0 8 16' width='8'><path d='M7.5 8L2.5 13l-1.5-1.5 3.75-3.5L1 4.5l1.5-1.5 5 5z'></path></svg></td><td class='icon'><svg aria-hidden='true' class='octicon octicon-file-text' height='16' role='img' version='1.1' viewBox='0 0 12 16' width='12'><path d='M6 5H2v-1h4v1zM2 8h7v-1H2v1z m0 2h7v-1H2v1z m0 2h7v-1H2v1z m10-7.5v9.5c0 0.55-0.45 1-1 1H1c-0.55 0-1-0.45-1-1V2c0-0.55 0.45-1 1-1h7.5l3.5 3.5z m-1 0.5L8 2H1v12h10V5z'></path></svg></td></td><td><a href="+hrefurl+" target='blank'><span class='res sg-res'>"+eachRes.Kind + " " + strToReturn + "</span></a></td></tr></tbody>");
				}

			}
		}
		else if (!dataArray.Defs && !(document.getElementById('404nomatch'))) {
			nomatch = "<div class='nomatch' id='nodefmatch'><p style='text-align:center;font-size:16px'><b> No matching definitions found. </br></b></p></div>";

			$('.tree-finder.clearfix:last').after(nomatch);

			if ($('#seeText').hasClass('selected')) {
				try{document.getElementById('nodefmatch').style.display='none'}catch(err){}
			}
			$('.tree-browser:last').attr("style","border-top:none;border-bottom:none;");
			document.getElementById('codeCounter').innerHTML = "0";

			return;
		}

	}

	/* --------------------------------------------Text search --------------------------------------------------------------*/

	//Generate def results table
	function showTextResults(dataArray){
		document.addEventListener('click', function(e){
			if(e.target.className === 'sgtextres') {
				amplitude.logEvent('ViewTextSearchResult')
			}
		})

		document.getElementById('textCounter').style.display='block';
		document.getElementById('textCounter').innerHTML = dataArray.Total;


		var codelist = "<div class='code-list' id=codelist style='margin-top:10px;'></div>";

		prevFile ='';
		$('.tree-finder.clearfix:last').append(codelist);

		if (!dataArray.SearchResults) {
			nomatch = "<div class='nomatch' ><p style='text-align:center;font-size:16px'><b> No matching text found. </br></b></p></div>";
			$('#codelist').append(nomatch);
			document.getElementById('textCounter').innerHTML = 0;
		}

		if ($('#seeText').hasClass('selected')){
			document.getElementById('codelist').style.display = 'block';
			$('#tree-finder-results2').attr("style", "display:none");
		}
		if (dataArray.SearchResults){


			for (var i = 0; i < dataArray.SearchResults.length; i++) {
				var filetypesplit = (dataArray.SearchResults[i].File).split(".");
				var filename = dataArray.SearchResults[i].File;
				var filetype = filetypesplit[filetypesplit.length-1];
				var startLine = dataArray.SearchResults[i].StartLine;
				var endLine = dataArray.SearchResults[i].EndLine;
				var lineNumber = startLine;
				var content = escapeHtml(window.atob(dataArray.SearchResults[i].Match));
				var match = query;
				var regexp = new RegExp (match, 'g');
				var toEnter = content.replace(regexp, "<span class='match'>"+query+"</span>");
				var contentArray = toEnter.split("\n");

				if (filename!==prevFile){
					$('.code-list').append("<div class='code-list-item code-list-item-public repo-specific'><span class='language'>"+filetype+"</span> <p class='title'><a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L"+lineNumber+"' title='"+filename+"' target='blank'>"+filename+"</a><br><span class='text-small text-muted match-count'> Results in "+filename+"</span></p><div class='file-box blob-wrapper'><table><tbody class='textres'></tbody>");
					for(var j = 0; j < contentArray.length; j++){
						try{$('.textres:last').append(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/-/blob/"+filename+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L"+lineNumber+"' target='blank' class='sgtextres'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[j]+" </td> </tr>")}catch(err){}
						lineNumber++;
					}
				}
				else {
					$('tr:last').after("<tr class='divider'><td class='blob-num'>…</td> <td class='blob-code'></td></tr>");
					for(var k = 0; k < contentArray.length; k++) {
						$('tr:last').after("<tr><td class='blob-num'><a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/-/blob/"+filename+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L"+lineNumber+"' target='blank' class='sgtextres'>"+lineNumber+"</a></td><td class='blob-code blob-code-inner'>"+contentArray[k]+"</td> </tr>");
						lineNumber++;
					}
				}
				prevFile = filename;
			}
		}
		var countScrolls=0;

		//listen for infinite scroll
		document.addEventListener('scroll', getInfiniteResults, true);
		if (! ($("body").height() > $(window).height()) && $('#seeText').hasClass('selected')) {
			subsequentAjaxCalls(countScrolls);
			countScrolls++;

		}
	}


	//Trigger infinite scroll
	function getInfiniteResults(){
		if ($(window).scrollTop() + $(window).height() === $(document).height() && $('#seeText').hasClass('selected')){
			subsequentAjaxCalls(countScrolls);
			countScrolls++;
		}
	}

	//Handler for clicking 'Text' on left nav menu
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

	//Handler for clicking 'Code' on left nav menu
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

	//Get subsequent text results for infinite scroll
	function subsequentAjaxCalls(numScrolls) {
		document.removeEventListener('scroll', getInfiniteResults, true);

		if (!document.getElementById('load-more-results')){
			$('.tree-finder.clearfix:last').append("<p style='text-align:center;font-size:16px' id='load-more-results'><b> Loading more results... </b></p>")
		}
		getInfiniteText = $.ajax ({
			method: "GET",
			url: "https://sourcegraph.com/.api/repos/github.com/"+user+"/"+repo+"@"+branch+"==="+commitID+"/-/tree-search?Query="+query+"&QueryType=fixed&N=10&ContextLines=2&Offset="+numScrolls*10,
			headers: {
				'authorization': 'Basic ' + window.btoa(token)
			}
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

	//show results on infinite scroll
	function infiniteTextResults(dataArray){
		if ($('#seeText').hasClass('selected')){
			document.getElementById('codelist').style.display = 'block';
			$('#tree-finder-results2').attr("style", "display:none");
		}

		for (var i = 0; i < dataArray.SearchResults.length; i++) {
			var filename = dataArray.SearchResults[i].File;
			if (prevFile !== filename) {
				$('.tree-finder.clearfix:last').append("<div class='code-list' id='codelist></div>");
			}
			var filetypesplit = (dataArray.SearchResults[i].File).split(".");
			var filetype = filetypesplit[filetypesplit.length-1];
			var startLine = dataArray.SearchResults[i].StartLine;
			var endLine = dataArray.SearchResults[i].EndLine;
			var lineNumber = startLine;
			var content = escapeHtml(window.atob(dataArray.SearchResults[i].Match));
			var match = query;
			var regexp = new RegExp (match, 'g');
			var toEnter = content.replace(regexp, "<span class='match'>"+query+"</span>");
			var contentArray = toEnter.split("\n");

			if (filename!==prevFile){
				$('.code-list').append("<div class='code-list-item code-list-item-public repo-specific'> <span class='language'>" +filetype+ "</span> <p class='title'><a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/-/blob/"+filename+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L"+lineNumber+"' title='"+filename+"' target='blank'>"+filename+"</a><br><span class='text-small text-muted match-count'> Results in "+filename+"</span></p><div class='file-box blob-wrapper'><table><tbody class='textres'></tbody>"); for(var j = 0; j < contentArray.length; j++) {$('.textres:last').append(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/.tree/"+filename+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L"+lineNumber+"' target='blank'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[j]+" </td> </tr>")
				lineNumber++;
			}
		}
		else {
			$('tr:last').after("<tr class='divider'> <td class='blob-num'>…</td> <td class='blob-code'></td> </tr>");
			for(var k = 0; k < contentArray.length; k++) {
				$('tr:last').after(" <tr> <td class='blob-num'> <a href='https://sourcegraph.com/github.com/"+user+"/"+repo+"@"+branch+"/-/blob/"+filename+"?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L"+lineNumber+"' target='blank'>"+lineNumber+"</a> </td> <td class='blob-code blob-code-inner'> "+contentArray[k]+" </td> </tr>")
				lineNumber++;
			}
		}
		prevFile = filename;
	}

		document.addEventListener('scroll', getInfiniteResults, true);

	}

})();

