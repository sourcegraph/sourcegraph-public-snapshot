package com.sourcegraph.cody.auth.ui

import com.intellij.openapi.util.Pair
import com.intellij.ui.scale.DerivedScaleType
import com.intellij.ui.scale.UserScaleContext
import java.util.concurrent.atomic.AtomicReference
import java.util.function.Function

open class ScaleContextCache<D, S : UserScaleContext>(
    private val myDataProvider: Function<in S, out D>
) {

  private val myData = AtomicReference<Pair<Double, D>?>(null)

  /**
   * Returns the data object from the cache if it matches the `ctx`, otherwise provides the new data
   * via the provider and caches it.
   */
  fun getOrProvide(ctx: S): D? {
    var data = myData.get()
    val scale = ctx.getScale(DerivedScaleType.PIX_SCALE)
    if (data == null || scale.compareTo(data.first) != 0) {
      myData.set(Pair.create(scale, myDataProvider.apply(ctx)).also { data = it })
    }
    return data!!.second
  }

  /** Clears the cache. */
  fun clear() {
    myData.set(null)
  }
}
