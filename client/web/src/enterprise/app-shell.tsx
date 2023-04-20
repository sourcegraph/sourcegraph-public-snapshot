console.log('app-shell.tsx loaded')

const redirectToServer = (url) => {
    fetch(url)
      .then(res => {
        if (res.ok) {
          window.location.href = url
        } else {
          setTimeout(redirectToServer, 1000)
        }
      })
      .catch(err => {
        setTimeout(redirectToServer, 1000)
      })
  }

  redirectToServer('http://localhost:3080')
