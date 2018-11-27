import { assert } from 'chai'
import { describe, it } from 'mocha'
import { TestScheduler } from 'rxjs/testing'
import { onMountRemovedFromDOM } from './overlay_mount'

describe('onMountRemovedFromDOM', () => {
    it('emits when the mount is removed from the DOM', () => {
        const scheduler = new TestScheduler((a, b) => assert.deepEqual(a, b))

        scheduler.run(({ cold, expectObservable }) => {
            const container = document.createElement('div')
            const mount = document.createElement('div')
            container.appendChild(mount)

            const mounts = cold('a', { a: mount })

            const removedFromDOM = mounts.pipe(onMountRemovedFromDOM())

            mounts.subscribe(() => {
                container.innerHTML = ''
            })

            expectObservable(removedFromDOM).toBe('a')
        })
    })
})
