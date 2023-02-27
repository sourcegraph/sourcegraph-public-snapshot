package foobar

import java.nio.channels.FileChannel

fun Mat.put(indices: IntArray, data: UShortArray)  = this.put(indices, data.asShortArray())

/***
 *  Example use:
 *
 *  val (b, g, r) = mat.at<UByte>(50, 50).v3c
 *  mat.at<UByte>(50, 50).val = T3(245u, 113u, 34u)
 *
 */
@Suppress("UNCHECKED_CAST")
inline fun <reified T> Mat.at(row: Int, col: Int) : Atable<T> =
    when (T::class) {
        UShort::class -> AtableUShort(this, row, col) as Atable<T>
        else -> throw RuntimeException("Unsupported class type")
    }


/**
 * Implementation of [DataAccess] which handles access and interactions with file and data
 * under scoped storage via the MediaStore API.
 */
@RequiresApi(Build.VERSION_CODES.Q)
internal class MediaStoreData(context: Context, filePath: String, accessFlag: FileAccessFlags) :
	DataAccess(filePath) {

	private data class DataItem(
		val id: Long,
		val mediaType: String
	)

	companion object {

		private val PROJECTION = arrayOf(
			MediaStore.Files.FileColumns._ID
		)

		private const val SELECTION_BY_PATH = "${MediaStore.Files.FileColumns.DISPLAY_NAME} = ? " +
			" AND ${MediaStore.Files.FileColumns.RELATIVE_PATH} = ?"

		private fun getSelectionByPathArguments(path: String): Array<String> {
			return arrayOf(getMediaStoreDisplayName(path), getMediaStoreRelativePath(path))
		}
	}
	override val fileChannel: FileChannel

	init {
		val contentResolver = context.contentResolver
		val dataItems = queryByPath(context, filePath)


		id = dataItem.id
		uri = dataItem.uri
	}
}
