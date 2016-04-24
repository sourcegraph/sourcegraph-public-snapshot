(function getAuthToken(){
	var xhr = new XMLHttpRequest();
	xhr.open('GET', 'https://sourcegraph.com', true);
	xhr.withCredentials = true;
	xhr.onload = function(){
		if(document.URL.split('/')[2]==='sourcegraph.com') {
			var reg = new RegExp (/(authorization\\":\\")(.*)(?=\\",\\"cache)/)
			var token =reg.exec(xhr.response)[2]
			chrome.storage.local.set({'value':token})
			chrome.storage.local.get(function(e){
					return e.value;
			})
		}
	}
	xhr.send()
})();

