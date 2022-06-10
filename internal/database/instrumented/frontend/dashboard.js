window.addEventListener('DOMContentLoaded', async () => {
  const formatNanoseconds = ns => {
    for (const divisor of [
      { unit: 1_000_000_000, suffix: 's' },
      { unit: 1_000_000, suffix: 'ms' },
      { unit: 1_000, suffix: 'μs' },
      { unit: 1, suffix: 'ns' },
    ]) {
      if (ns >= divisor.unit) {
        return (ns / divisor.unit).toLocaleString(undefined, { maximumFractionDigits: 3 }) + divisor.suffix
      }
    }

    return '0'
  }

  let expanded = new Set()
  let cache = new Map()
  const renderList = async queries => {
    const container = document.getElementById('queries')
    container.innerHTML = ''
    queries.forEach(async query => {
      const isExpanded = expanded.has(query.id)

      const chevron = document.createElement('button')
      chevron.className = 'chevron'
      chevron.innerText = isExpanded ? '^' : '>'
      chevron.addEventListener('click', async event => {
        event.preventDefault()
        if (expanded.has(query.id)) {
          expanded.delete(query.id)
        } else {
          expanded.add(query.id)
        }
        await renderList(queries)
      })
      container.appendChild(chevron)

      const time = document.createElement('div')
      time.className = 'time'
      if (query.time_ns > 500_000_000) {
        time.className += ' very-slow'
      } else if (query.time_ns > 1_000_000) {
        time.className += ' slow'
      }
      time.innerText = formatNanoseconds(query.time_ns)
      container.appendChild(time)

      const hasPlan = document.createElement('div')
      hasPlan.className = 'has-plan'
      hasPlan.innerText = query.has_plan ? '🗺️' : ''
      container.appendChild(hasPlan)

      const hasError = document.createElement('div')
      hasError.className = 'has-error'
      hasError.innerText = query.has_err ? '' : ''
      container.appendChild(hasError)

      const queryString = document.createElement('div')
      queryString.className = 'queryString' + (isExpanded ? ' expanded' : '')
      if (isExpanded) {
        queryString.innerText = query.query
      } else {
        queryString.appendChild(document.createTextNode(query.query))
      }
      container.appendChild(queryString)

      if (expanded.has(query.id)) {
        let detail
        if (cache.has(query.id)) {
          detail = cache.get(query.id)
        } else {
          const response = await fetch(`/query?id=${query.id}`)
          detail = await response.json()
          cache.set(query.id, detail)
        }

        const plan = document.createElement('div')
        plan.className = 'plan'
        detail.plan.forEach(row => {
          const line = document.createElement('div')
          line.innerText = row
          plan.appendChild(line)
        })
        queryString.insertAdjacentElement('afterend', plan)
      }
    })
  }

  const refreshQueries = async () => {
    const response = await fetch('/queries')
    const queries = await response.json()

    await renderList(queries)
  }

  let interval = window.setInterval(refreshQueries, 5000)
  document.getElementById('toggle').addEventListener('click', event => {
    event.preventDefault()
    if (interval !== undefined) {
      window.clearInterval(interval)
      interval = undefined
      event.target.innerText = 'Resume'
    } else {
      interval = window.setInterval(refreshQueries, 5000)
      event.target.innerText = 'Pause'
    }
  })

  await refreshQueries()
})
