// Create the overlay
const overlay = document.createElement('div')
overlay.className = 'cody-overlay'
overlay.id = 'cody-overlay'
// create button element and set its styles
const closeButton = document.createElement('button')
closeButton.innerHTML = 'x'
closeButton.className = 'cody-overlay-close-btn'
// add event listener to button element to remove overlay on click
closeButton.addEventListener('click', () => {
    overlay.remove()
})
document.body.appendChild(overlay)

chrome.runtime.onMessage.addListener(request => {
    if (request.type === 'cody') {
        const query = request.data.query
        if (query === 'auth' && !request.data.message) {
            overlay.innerHTML = 'You must be sign in to use Cody.'
            overlay.className = 'cody-overlay-shows'
            overlay.appendChild(closeButton)
            return
        }
        if (query === 'wait') {
            overlay.innerHTML = 'Cody is typing...'
            overlay.className = 'cody-overlay-shows'
            overlay.appendChild(closeButton)
        } else {
            const answer = request.data.message
            overlay.innerText = answer ? answer : 'Connection failed'
            overlay.className = 'cody-overlay-shows'
            overlay.appendChild(closeButton)
        }
    }
})
