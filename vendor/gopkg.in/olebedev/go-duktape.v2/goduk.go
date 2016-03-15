// +build ignore

// duktape project duktape.go
package goduk

/*
#cgo CFLAGS: -std=c99 -O2 -Os -fomit-frame-pointer -fstrict-aliasing -DDUK_OPT_NO_ES6_OBJECT_SETPROTOTYPEOF -DDUK_OPT_NO_ES6_OBJECT_PROTO_PROPERTY -DDUK_OPT_NO_ES6_PROXY -DDUK_OPT_NO_AUGMENT_ERRORS -DDUK_OPT_NO_TRACEBACKS

int _get_output_format() { return 1; }

#include "duktape.h"

// wrappers for variadic functions

void __duk_error_raw(duk_context *ctx, duk_errcode_t err_code, const char *filename, duk_int_t line, const char *text) {
	duk_error_raw(ctx, err_code, filename, line, text);
}

duk_idx_t __duk_push_error_object_raw(duk_context *ctx, duk_errcode_t err_code, const char *filename, duk_int_t line, const char *text) {
	return duk_push_error_object_raw(ctx, err_code, filename, line, text);
}

void __duk_log(duk_context *ctx, duk_int_t level, const char *text) {
	duk_log(ctx, level, text);
}
*/
import "C"
import "unsafe"
import "io/ioutil"

//TODO
// revoir les fonction buffer

type Context struct {
	ctx *C.struct_duk_context
}

/* ----------------- */
/* Context managment */
/* ----------------- */

// duk_context *duk_create_heap_default(void);
func CreateHeapDefault() *Context {
	ctx := new(Context)
	ctx.ctx = (*C.struct_duk_context)(C.duk_create_heap(nil, nil, nil, nil, nil))

	return ctx
}

// void duk_destroy_heap(duk_context *ctx);
func (this Context) Destroy() {
	C.duk_destroy_heap(unsafe.Pointer(this.ctx))

	funcs_ctxDel(unsafe.Pointer(this.ctx))

	this.ctx = nil
}

/* -------------- */
/* Error handling */
/* -------------- */

// void duk_throw(duk_context *ctx);
func (this Context) Throw() {
	C.duk_throw(unsafe.Pointer(this.ctx))
}

// void duk_fatal(duk_context *ctx, duk_errcode_t err_code, const char *err_msg);
func (this Context) Fatal(err_code int, err_msg string) {
	cS_err_msg := C.CString(err_msg)
	defer C.free(unsafe.Pointer(cS_err_msg))

	C.duk_fatal(unsafe.Pointer(this.ctx), C.duk_errcode_t(err_code), cS_err_msg)
}

// void __duk_error_raw(duk_context *ctx, duk_errcode_t err_code, const char *filename, duk_int_t line, const char *text)
func (this Context) Error(err_code int, filename string, line int, err_msg string) {
	cS_filename := C.CString(filename)
	cS_err_msg := C.CString(err_msg)
	defer C.free(unsafe.Pointer(cS_filename))
	defer C.free(unsafe.Pointer(cS_err_msg))

	C.__duk_error_raw(unsafe.Pointer(this.ctx), C.duk_errcode_t(err_code), cS_filename, C.duk_int_t(line), cS_err_msg)
}

/* ----------------------------- */
/* Other state related functions */
/* ----------------------------- */

// duk_int_t duk_is_strict_call(duk_context *ctx);
func (this Context) IsStrictCall() int {
	return int(C.duk_is_strict_call(unsafe.Pointer(this.ctx)))
}

// duk_int_t duk_is_constructor_call(duk_context *ctx);
func (this Context) IsConstructorCall() int {
	return int(C.duk_is_constructor_call(unsafe.Pointer(this.ctx)))
}

/* ---------------- */
/* Stack management */
/* ---------------- */

// duk_idx_t duk_normalize_index(duk_context *ctx, duk_idx_t index);
func (this Context) NormalizeIndex(index int) int {
	return int(C.duk_normalize_index(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_idx_t duk_require_normalize_index(duk_context *ctx, duk_idx_t index);
func (this Context) RequireNormalizeIndex(index int) int {
	return int(C.duk_require_normalize_index(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int_t duk_is_valid_index(duk_context *ctx, duk_idx_t index);
func (this Context) IsValidIndex(index int) int {
	return int(C.duk_is_valid_index(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// void duk_require_valid_index(duk_context *ctx, duk_idx_t index);
func (this Context) RequireValidIndex(index int) {
	C.duk_require_valid_index(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// duk_idx_t duk_get_top(duk_context *ctx);
func (this Context) GetTop() int {
	return int(C.duk_get_top(unsafe.Pointer(this.ctx)))
}

// void duk_set_top(duk_context *ctx, duk_idx_t index);
func (this Context) SetTop(index int) {
	C.duk_set_top(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// duk_idx_t duk_get_top_index(duk_context *ctx);
func (this Context) DukGetTopIndex() int {
	return int(C.duk_get_top_index(unsafe.Pointer(this.ctx)))
}

// duk_idx_t duk_require_top_index(duk_context *ctx);
func (this Context) RequireTopIndex() int {
	return int(C.duk_require_top_index(unsafe.Pointer(this.ctx)))
}

// duk_int_t duk_check_stack(duk_context *ctx, duk_idx_t extra);
func (this Context) CheckStack(extra int) int {
	return int(C.duk_check_stack(unsafe.Pointer(this.ctx), C.duk_idx_t(extra)))
}

// void duk_require_stack(duk_context *ctx, duk_idx_t extra);
func (this Context) RequireStack(extra int) {
	C.duk_require_stack(unsafe.Pointer(this.ctx), C.duk_idx_t(extra))
}

// duk_int_t duk_check_stack_top(duk_context *ctx, duk_idx_t top);
func (this Context) CheckStackTop(top int) int {
	return int(C.duk_check_stack_top(unsafe.Pointer(this.ctx), C.duk_idx_t(top)))
}

// void duk_require_stack_top(duk_context *ctx, duk_idx_t top);
func (this Context) RequireStackTop(top int) {
	C.duk_require_stack_top(unsafe.Pointer(this.ctx), C.duk_idx_t(top))
}

/* ---------------------------------------- */
/* Stack manipulation (other than push/pop) */
/* ---------------------------------------- */

// void duk_swap(duk_context *ctx, duk_idx_t index1, duk_idx_t index2);
func (this Context) Swap(index1, index2 int) {
	C.duk_swap(unsafe.Pointer(this.ctx), C.duk_idx_t(index1), C.duk_idx_t(index2))
}

// void duk_swap_top(duk_context *ctx, duk_idx_t index);
func (this Context) SwapTop(index int) {
	C.duk_swap_top(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_dup(duk_context *ctx, duk_idx_t from_index);
func (this Context) Dup(from_index int) {
	C.duk_dup(unsafe.Pointer(this.ctx), C.duk_idx_t(from_index))
}

// void duk_dup_top(duk_context *ctx);
func (this Context) DupTop() {
	C.duk_dup_top(unsafe.Pointer(this.ctx))
}

// void duk_insert(duk_context *ctx, duk_idx_t to_index);
func (this Context) Insert(to_index int) {
	C.duk_insert(unsafe.Pointer(this.ctx), C.duk_idx_t(to_index))
}

// void duk_replace(duk_context *ctx, duk_idx_t to_index);
func (this Context) Replace(to_index int) {
	C.duk_replace(unsafe.Pointer(this.ctx), C.duk_idx_t(to_index))
}

// void duk_copy(duk_context *ctx, duk_idx_t from_index, duk_idx_t to_index);
func (this Context) Copt(from_index, to_index int) {
	C.duk_copy(unsafe.Pointer(this.ctx), C.duk_idx_t(from_index), C.duk_idx_t(to_index))
}

// void duk_remove(duk_context *ctx, duk_idx_t index);
func (this Context) Remove(index int) {
	C.duk_remove(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_xcopymove_raw(duk_context *to_ctx, duk_context *from_ctx, duk_idx_t count, duk_int_t is_copy);
func (this Context) XCopyMoveRaw(from_ctx Context, count, is_copy int) {
	C.duk_xcopymove_raw(unsafe.Pointer(this.ctx), unsafe.Pointer(from_ctx.ctx), C.duk_idx_t(count), C.duk_bool_t(is_copy))
}

// #define duk_xmove_top(to_ctx,from_ctx,count) duk_xcopymove_raw((to_ctx), (from_ctx), (count), 0 /*is_copy*/)
func (this Context) XMoveTop(from_ctx Context, count int) {
	this.XCopyMoveRaw(from_ctx, count, 0)
}

// #define duk_xcopy_top(to_ctx,from_ctx,count) duk_xcopymove_raw((to_ctx), (from_ctx), (count), 1 /*is_copy*/)
func (this Context) XCopyTop(from_ctx Context, count int) {
	this.XCopyMoveRaw(from_ctx, count, 1)
}

/* --------------- */
/* Push operations */
/* --------------- */
// Push functions return the absolute (relative to bottom of frame) position of the pushed value for convenience.
// Note: duk_dup() is technically a push.

// void duk_push_undefined(duk_context *ctx);
func (this Context) PushUndefined() {
	C.duk_push_undefined(unsafe.Pointer(this.ctx))
}

// void duk_push_null(duk_context *ctx);
func (this Context) PushNull() {
	C.duk_push_null(unsafe.Pointer(this.ctx))
}

// void duk_push_boolean(duk_context *ctx, duk_bool_t val);
func (this Context) PushBoolean(val bool) {
	if val {
		C.duk_push_boolean(unsafe.Pointer(this.ctx), C.duk_bool_t(1))
	} else {
		C.duk_push_boolean(unsafe.Pointer(this.ctx), C.duk_bool_t(0))
	}
}

// void duk_push_true(duk_context *ctx);
func (this Context) PushTrue() {
	C.duk_push_true(unsafe.Pointer(this.ctx))
}

// void duk_push_false(duk_context *ctx);
func (this Context) PushFalse() {
	C.duk_push_false(unsafe.Pointer(this.ctx))
}

// void duk_push_number(duk_context *ctx, duk_double_t val);
func (this Context) PushNumber(val float64) {
	C.duk_push_number(unsafe.Pointer(this.ctx), C.duk_double_t(val))
}

// void duk_push_nan(duk_context *ctx);
func (this Context) PushNan() {
	C.duk_push_nan(unsafe.Pointer(this.ctx))
}

// void duk_push_int(duk_context *ctx, duk_int_t val);
func (this Context) PushInt(val int) {
	C.duk_push_int(unsafe.Pointer(this.ctx), C.duk_int_t(val))
}

// void duk_push_uint(duk_context *ctx, duk_uint_t val);
func (this Context) PushUInt(val uint) {
	C.duk_push_uint(unsafe.Pointer(this.ctx), C.duk_uint_t(val))
}

// const char *duk_push_string(duk_context *ctx, const char *str);
// const char *duk_push_lstring(duk_context *ctx, const char *str, duk_size_t len);
func (this Context) PushString(str string) {
	bstr := []byte(str)
	if len(bstr) <= 0 {
		bstr = make([]byte, 1)
		bstr[0] = 0
	}
	C.duk_push_lstring(unsafe.Pointer(this.ctx), (*C.char)(unsafe.Pointer(&bstr[0])), C.duk_size_t(len(bstr)))
}

// void duk_push_pointer(duk_context *ctx, void *p);
func (this Context) PushPointer(p uintptr) {
	C.duk_push_pointer(unsafe.Pointer(this.ctx), unsafe.Pointer(p))
}

//TODO implement (or not :p)
// const char *duk_push_sprintf(duk_context *ctx, const char *fmt, ...);
// const char *duk_push_vsprintf(duk_context *ctx, const char *fmt, va_list ap);

// const char *duk_push_string_file_raw(duk_context *ctx, const char *path, duk_uint_t flags);
func (this Context) PushStringFileRaw(path string, flags uint) error {
	data, e := ioutil.ReadFile(path)
	if e != nil {
		return e
	}

	this.PushString(string(data))
	return nil
}

// #define duk_push_string_file(ctx,path) duk_push_string_file_raw((ctx), (path), 0)
func (this Context) PushStringFile(path string) error {
	return this.PushStringFileRaw(path, 0)
}

// void duk_push_this(duk_context *ctx);
func (this Context) PushThis() {
	C.duk_push_this(unsafe.Pointer(this.ctx))
}

// void duk_push_current_function(duk_context *ctx);
func (this Context) PushCurrentFunction() {
	C.duk_push_current_function(unsafe.Pointer(this.ctx))
}

// void duk_push_current_thread(duk_context *ctx);
func (this Context) PushCurrentThread() {
	C.duk_push_current_thread(unsafe.Pointer(this.ctx))
}

// void duk_push_global_object(duk_context *ctx);
func (this Context) PushGlobalObject() {
	C.duk_push_global_object(unsafe.Pointer(this.ctx))
}

// void duk_push_heap_stash(duk_context *ctx);
func (this Context) PushHeapStach() {
	C.duk_push_heap_stash(unsafe.Pointer(this.ctx))
}

// void duk_push_global_stash(duk_context *ctx);
func (this Context) PushGlobalStash() {
	C.duk_push_global_stash(unsafe.Pointer(this.ctx))
}

// void duk_push_thread_stash(duk_context *ctx, duk_context *target_ctx);
func (this Context) PushThreadStash(target_ctx Context) {
	C.duk_push_thread_stash(unsafe.Pointer(this.ctx), unsafe.Pointer(target_ctx.ctx))
}

// duk_idx_t duk_push_object(duk_context *ctx);
func (this Context) PushObject() int {
	return int(C.duk_push_object(unsafe.Pointer(this.ctx)))
}

// duk_idx_t duk_push_array(duk_context *ctx);
func (this Context) PushArray() int {
	return int(C.duk_push_array(unsafe.Pointer(this.ctx)))
}

//TODO implement foreign functions handling
// duk_idx_t duk_push_c_function(duk_context *ctx, duk_c_function func, duk_idx_t nargs);
// duk_idx_t duk_push_c_lightfunc(duk_context *ctx, duk_c_function func, duk_idx_t nargs, duk_idx_t length, duk_int_t magic);

// duk_idx_t duk_push_thread_raw(duk_context *ctx, duk_uint_t flags);
func (this Context) PushThreadRaw(flags uint) int {
	return int(C.duk_push_thread_raw(unsafe.Pointer(this.ctx), C.duk_uint_t(flags)))
}

// #define duk_push_thread(ctx) duk_push_thread_raw((ctx), 0 /*flags*/)
func (this Context) PushThread() int {
	return int(C.duk_push_thread_raw(unsafe.Pointer(this.ctx), C.duk_uint_t(0)))
}

// #define duk_push_thread_new_globalenv(ctx) duk_push_thread_raw((ctx), DUK_THREAD_NEW_GLOBAL_ENV /*flags*/)
func (this Context) PushThreadNewGlobalEnv() int {
	return int(C.duk_push_thread_raw(unsafe.Pointer(this.ctx), C.duk_uint_t(THREAD_NEW_GLOBAL_ENV)))
}

// duk_idx_t __duk_push_error_object_raw(duk_context *ctx, duk_errcode_t err_code, const char *filename, duk_int_t line, const char *text)
func (this Context) PushErrorObjectRaw(err_code int, filename string, line int, text string) int {
	cS_filename := C.CString(filename)
	cS_text := C.CString(text)
	defer C.free(unsafe.Pointer(cS_filename))
	defer C.free(unsafe.Pointer(cS_text))

	return int(C.__duk_push_error_object_raw(unsafe.Pointer(this.ctx), C.duk_errcode_t(err_code), cS_filename, C.duk_int_t(line), cS_text))
}

// void *duk_push_buffer_raw(duk_context *ctx, duk_size_t size, duk_int_t dynamic);
func (this Context) pushBufferRaw(buf []byte, dynamic int) {
	ptr := C.duk_push_buffer_raw(unsafe.Pointer(this.ctx), C.duk_size_t(len(buf)), C.duk_bool_t(dynamic))
	if ptr != nil {
		C.memcpy(ptr, unsafe.Pointer(&buf[0]), C.size_t(len(buf)))
	}
}

// #define duk_push_buffer(ctx,size,dynamic) duk_push_buffer_raw((ctx), (size), (dynamic));
func (this Context) PushBuffer(buf []byte, dynamic int) {
	this.pushBufferRaw(buf, dynamic)
}

// #define duk_push_fixed_buffer(ctx,size) duk_push_buffer_raw((ctx), (size), 0 /*dynamic*/)
func (this Context) PushFixedBuffer(buf []byte) {
	this.pushBufferRaw(buf, 0)
}

// #define duk_push_dynamic_buffer(ctx,size) duk_push_buffer_raw((ctx), (size), 1 /*dynamic*/)
func (this Context) PushDynamicBuffer(buf []byte) {
	this.pushBufferRaw(buf, 1)
}

// duk_idx_t duk_push_heapptr(duk_context *ctx, void *ptr);
func (this Context) PushHeapPtr(ptr uintptr) int {
	return int(C.duk_push_heapptr(unsafe.Pointer(this.ctx), unsafe.Pointer(ptr)))
}

/* -------------- */
/* Pop operations */
/* -------------- */

// void duk_pop(duk_context *ctx);
func (this Context) Pop() {
	C.duk_pop(unsafe.Pointer(this.ctx))
}

// void duk_pop_n(duk_context *ctx, duk_idx_t count);
func (this Context) PopN(count int) {
	C.duk_pop_n(unsafe.Pointer(this.ctx), C.duk_idx_t(count))
}

// void duk_pop_2(duk_context *ctx);
func (this Context) Pop2() {
	C.duk_pop_2(unsafe.Pointer(this.ctx))
}

// void duk_pop_3(duk_context *ctx);
func (this Context) Pop3() {
	C.duk_pop_3(unsafe.Pointer(this.ctx))
}

/* ----------- */
/* Type checks */
/* ----------- */
// duk_is_none(), which would indicate whether index it outside of stack, is not needed;
// duk_is_valid_index() gives the same information.

// duk_int_t duk_get_type(duk_context *ctx, duk_idx_t index);
func (this Context) GetType(index int) int {
	return int(C.duk_get_type(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int_t duk_check_type(duk_context *ctx, duk_idx_t index, duk_int_t type);
func (this Context) CheckType(index, _type int) int {
	return int(C.duk_check_type(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_int_t(_type)))
}

// duk_uint_t duk_get_type_mask(duk_context *ctx, duk_idx_t index);
func (this Context) GetTypeMask(index int) uint {
	return uint(C.duk_get_type_mask(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int_t duk_check_type_mask(duk_context *ctx, duk_idx_t index, duk_uint_t mask);
func (this Context) CheckTypeMask(index int, mask uint) int {
	return int(C.duk_check_type_mask(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_uint_t(mask)))
}

// duk_int_t duk_is_undefined(duk_context *ctx, duk_idx_t index);
func (this Context) IsUndefined(index int) bool {
	if int(C.duk_is_undefined(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) == 0 {
		return false
	} else {
		return true
	}
}

// duk_int_t duk_is_null(duk_context *ctx, duk_idx_t index);
func (this Context) IsNull(index int) bool {
	if int(C.duk_is_null(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) == 0 {
		return false
	} else {
		return true
	}
}

// duk_int_t duk_is_null_or_undefined(duk_context *ctx, duk_idx_t index);
func (this Context) IsNullOrUndefined(index int) bool {
	if int(C.duk_is_null_or_undefined(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) == 0 {
		return false
	} else {
		return true
	}
}

// duk_int_t duk_is_boolean(duk_context *ctx, duk_idx_t index);
func (this Context) IsBoolean(index int) bool {
	return int(C.duk_is_boolean(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_number(duk_context *ctx, duk_idx_t index);
func (this Context) IsNumber(index int) bool {
	return int(C.duk_is_number(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_nan(duk_context *ctx, duk_idx_t index);
func (this Context) IsNan(index int) bool {
	return int(C.duk_is_nan(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_string(duk_context *ctx, duk_idx_t index);
func (this Context) IsString(index int) bool {
	return int(C.duk_is_string(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_object(duk_context *ctx, duk_idx_t index);
func (this Context) IsObject(index int) bool {
	return int(C.duk_is_object(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_buffer(duk_context *ctx, duk_idx_t index);
func (this Context) IsBuffer(index int) bool {
	return int(C.duk_is_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_pointer(duk_context *ctx, duk_idx_t index);
func (this Context) IsPointer(index int) bool {
	return int(C.duk_is_pointer(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_lightfunc(duk_context *ctx, duk_idx_t index);
func (this Context) IsLightFunc(index int) bool {
	return int(C.duk_is_lightfunc(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_array(duk_context *ctx, duk_idx_t index);
func (this Context) IsArray(index int) bool {
	return int(C.duk_is_array(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_function(duk_context *ctx, duk_idx_t index);
func (this Context) IsFunction(index int) bool {
	return int(C.duk_is_function(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_c_function(duk_context *ctx, duk_idx_t index);
func (this Context) IsCFunction(index int) bool {
	return int(C.duk_is_c_function(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_ecmascript_function(duk_context *ctx, duk_idx_t index);
func (this Context) IsEcmascriptFunction(index int) bool {
	return int(C.duk_is_ecmascript_function(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_bound_function(duk_context *ctx, duk_idx_t index);
func (this Context) IsBoundFunction(index int) bool {
	return int(C.duk_is_bound_function(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_thread(duk_context *ctx, duk_idx_t index);
func (this Context) IsThread(index int) bool {
	return int(C.duk_is_thread(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_callable(duk_context *ctx, duk_idx_t index);
func (this Context) IsCallable(index int) bool {
	return int(C.duk_is_callable(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_dynamic_buffer(duk_context *ctx, duk_idx_t index);
func (this Context) IsDynamicBuffer(index int) bool {
	return int(C.duk_is_dynamic_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_fixed_buffer(duk_context *ctx, duk_idx_t index);
func (this Context) IsFixedBuffer(index int) bool {
	return int(C.duk_is_dynamic_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_int_t duk_is_primitive(duk_context *ctx, duk_idx_t index);
func (this Context) IsPrimitive(index int) bool {
	return int(C.duk_is_primitive(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// #define duk_is_object_coercible(ctx,index) duk_check_type_mask((ctx), (index), DUK_TYPE_MASK_BOOLEAN | DUK_TYPE_MASK_NUMBER | DUK_TYPE_MASK_STRING | DUK_TYPE_MASK_OBJECT | DUK_TYPE_MASK_BUFFER | DUK_TYPE_MASK_POINTER | DUK_TYPE_MASK_LIGHTFUNC)
func (this Context) IsObjectCoercible(index int) bool {
	mask := TYPE_MASK_BOOLEAN | TYPE_MASK_NUMBER | TYPE_MASK_STRING | TYPE_MASK_OBJECT | TYPE_MASK_BUFFER | TYPE_MASK_POINTER | TYPE_MASK_LIGHTFUNC
	return int(C.duk_check_type_mask(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_uint_t(mask))) != 0
}

// duk_errcode_t duk_get_error_code(duk_context *ctx, duk_idx_t index);
func (this Context) GetErrorCode(index int) int {
	return int(C.duk_get_error_code(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// #define duk_is_error(ctx,index) (duk_get_error_code((ctx), (index)) != 0)
func (this Context) IsError(index int) bool {
	return C.duk_get_error_code(unsafe.Pointer(this.ctx), C.duk_idx_t(index)) != C.duk_errcode_t(0)
}

/* -------------- */
/* Get operations */
/* -------------- */
// no coercion, returns default value for invalid indices and invalid value types.
// duk_get_undefined() and duk_get_null() would be pointless and are not included.

// duk_int_t duk_get_boolean(duk_context *ctx, duk_idx_t index);
func (this Context) GetBoolean(index int) bool {
	return int(C.duk_get_boolean(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_double_t duk_get_number(duk_context *ctx, duk_idx_t index);
func (this Context) GetNumber(index int) float64 {
	return float64(C.duk_get_number(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int_t duk_get_int(duk_context *ctx, duk_idx_t index);
func (this Context) GetInt(index int) int {
	return int(C.duk_get_int(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_uint_t duk_get_uint(duk_context *ctx, duk_idx_t index);
func (this Context) GetUint(index int) uint {
	return uint(C.duk_get_uint(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// const char *duk_get_string(duk_context *ctx, duk_idx_t index);
// const char *duk_get_lstring(duk_context *ctx, duk_idx_t index, duk_size_t *out_len);
func (this Context) GetString(index int) string {
	var (
		ret     *C.char
		out_len C.duk_size_t
	)
	ret = C.duk_get_lstring(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_len)
	return C.GoStringN(ret, C.int(out_len))
}

// void *duk_get_buffer(duk_context *ctx, duk_idx_t index, duk_size_t *out_size);
func (this Context) GetBuffer(index int) []byte {
	var (
		ret      unsafe.Pointer
		out_size C.duk_size_t
	)
	ret = C.duk_get_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_size)
	return C.GoBytes(ret, C.int(out_size))
}

// void *duk_get_pointer(duk_context *ctx, duk_idx_t index);
func (this Context) GetPointer(index int) uintptr {
	return uintptr(C.duk_get_pointer(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

//TODO check implementation of foreign functions
// duk_c_function duk_get_c_function(duk_context *ctx, duk_idx_t index);

// duk_context *duk_get_context(duk_context *ctx, duk_idx_t index);
func (this Context) GetContext(index int) *Context {
	ctx := C.duk_get_context(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
	if ctx != nil {
		ret := new(Context)
		ret.ctx = (*C.struct_duk_context)(ctx)
		return ret
	} else {
		return nil
	}
}

// void *duk_get_heapptr(duk_context *ctx, duk_idx_t index);
func (this Context) GetHeapPtr(index int) uintptr {
	return uintptr(C.duk_get_heapptr(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_size_t duk_get_length(duk_context *ctx, duk_idx_t index);
func (this Context) GetLength(index int) uint {
	return uint(C.duk_get_length(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

/* ------------------ */
/* Require operations */
/* ------------------ */

// #define duk_require_type_mask(ctx,index,mask) ((void) duk_check_type_mask((ctx), (index), (mask) | DUK_TYPE_MASK_THROW))
func (this Context) RequireTypeMask(index, mask int) {
	C.duk_check_type_mask(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_uint_t(mask|TYPE_MASK_THROW))
}

// void duk_require_undefined(duk_context *ctx, duk_idx_t index);
func (this Context) RequireUndefined(index int) {
	C.duk_require_undefined(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_require_null(duk_context *ctx, duk_idx_t index);
func (this Context) RequireNull(index int) {
	C.duk_require_null(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// duk_int_t duk_require_boolean(duk_context *ctx, duk_idx_t index);
func (this Context) RequireBoolean(index int) bool {
	return int(C.duk_require_boolean(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_double_t duk_require_number(duk_context *ctx, duk_idx_t index);
func (this Context) RequireNumber(index int) float64 {
	return float64(C.duk_require_number(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int_t duk_require_int(duk_context *ctx, duk_idx_t index);
func (this Context) RequireInt(index int) int {
	return int(C.duk_require_int(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_uint_t duk_require_uint(duk_context *ctx, duk_idx_t index);
func (this Context) RequireUint(index int) uint {
	return uint(C.duk_require_uint(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// const char *duk_require_string(duk_context *ctx, duk_idx_t index);
// const char *duk_require_lstring(duk_context *ctx, duk_idx_t index, duk_size_t *out_len);
func (this Context) RequireString(index int) string {
	var (
		ret     *C.char
		out_len C.duk_size_t
	)
	ret = C.duk_require_lstring(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_len)
	return C.GoStringN(ret, C.int(out_len))
}

// void *duk_require_buffer(duk_context *ctx, duk_idx_t index, duk_size_t *out_size);
func (this Context) RequireBuffer(index int) []byte {
	var (
		ret      unsafe.Pointer
		out_size C.duk_size_t
	)
	ret = C.duk_require_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_size)
	return C.GoBytes(ret, C.int(out_size))
}

// void *duk_require_pointer(duk_context *ctx, duk_idx_t index);
func (this Context) RequirePointer(index int) uintptr {
	return uintptr(C.duk_require_pointer(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

//TODO foreign function handling
// duk_c_function duk_require_c_function(duk_context *ctx, duk_idx_t index);

// duk_context *duk_require_context(duk_context *ctx, duk_idx_t index);
func (this Context) RequireContext(index int) *Context {
	ctx := C.duk_require_context(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
	if ctx != nil {
		ret := new(Context)
		ret.ctx = (*C.struct_duk_context)(ctx)
		return ret
	} else {
		return nil
	}
}

// void *duk_require_heapptr(duk_context *ctx, duk_idx_t index);
func (this Context) RequireHeapPtr(index int) uintptr {
	return uintptr(C.duk_require_heapptr(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// #define duk_require_object_coercible(ctx,index) ((void) duk_check_type_mask((ctx), (index), DUK_TYPE_MASK_BOOLEAN | DUK_TYPE_MASK_NUMBER | DUK_TYPE_MASK_STRING | DUK_TYPE_MASK_OBJECT | DUK_TYPE_MASK_BUFFER | DUK_TYPE_MASK_POINTER | DUK_TYPE_MASK_LIGHTFUNC | DUK_TYPE_MASK_THROW))
func (this Context) RequireObjectCoercible(index int) {
	mask := TYPE_MASK_BOOLEAN | TYPE_MASK_NUMBER | TYPE_MASK_STRING | TYPE_MASK_OBJECT | TYPE_MASK_BUFFER | TYPE_MASK_POINTER | TYPE_MASK_LIGHTFUNC | TYPE_MASK_THROW
	C.duk_check_type_mask(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_uint_t(mask))
}

/* ------------------- */
/* Coercion operations */
/* ------------------- */
// in-place coercion, return coerced value where applicable. If index is invalid, throw error.
// Some coercions may throw an expected error (e.g. from a toString() or valueOf() call) or an internal error (e.g. from out of memory).

// void duk_to_undefined(duk_context *ctx, duk_idx_t index);
func (this Context) ToUndefined(index int) {
	C.duk_to_undefined(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_to_null(duk_context *ctx, duk_idx_t index);
func (this Context) ToNull(index int) {
	C.duk_to_null(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// duk_int_t duk_to_boolean(duk_context *ctx, duk_idx_t index);
func (this Context) ToBoolean(index int) bool {
	return int(C.duk_to_boolean(unsafe.Pointer(this.ctx), C.duk_idx_t(index))) != 0
}

// duk_double_t duk_to_number(duk_context *ctx, duk_idx_t index);
func (this Context) ToNumber(index int) float64 {
	return float64(C.duk_to_number(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int_t duk_to_int(duk_context *ctx, duk_idx_t index);
func (this Context) ToInt(index int) int {
	return int(C.duk_to_int(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_uint_t duk_to_uint(duk_context *ctx, duk_idx_t index);
func (this Context) ToUint(index int) uint {
	return uint(C.duk_to_uint(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_int32_t duk_to_int32(duk_context *ctx, duk_idx_t index);
func (this Context) ToInt32(index int) int32 {
	return int32(C.duk_to_int32(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_uint32_t duk_to_uint32(duk_context *ctx, duk_idx_t index);
func (this Context) ToUint32(index int) uint32 {
	return uint32(C.duk_to_uint32(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// duk_uint16_t duk_to_uint16(duk_context *ctx, duk_idx_t index);
func (this Context) ToUint16(index int) uint16 {
	return uint16(C.duk_to_uint16(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// const char *duk_to_string(duk_context *ctx, duk_idx_t index);
// const char *duk_to_lstring(duk_context *ctx, duk_idx_t index, duk_size_t *out_len);
func (this Context) ToString(index int) string {
	var (
		ret     *C.char
		out_len C.duk_size_t
	)
	ret = C.duk_to_lstring(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_len)

	return C.GoStringN(ret, C.int(out_len))
}

// void *duk_to_buffer_raw(duk_context *ctx, duk_idx_t index, duk_size_t *out_size, duk_uint_t flags);
func (this Context) toBufferRaw(index int, flags uint) []byte {
	var (
		ret      unsafe.Pointer
		out_size C.duk_size_t
	)
	ret = C.duk_to_buffer_raw(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_size, C.duk_uint_t(flags))

	if ret != nil {
		return C.GoBytes(ret, C.int(out_size))
	} else {
		return make([]byte, 0)
	}
}

// void *duk_to_pointer(duk_context *ctx, duk_idx_t index);
func (this Context) ToPointer(index int) uintptr {
	return uintptr(C.duk_to_pointer(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// void duk_to_object(duk_context *ctx, duk_idx_t index);
func (this Context) ToObject(index int) {
	C.duk_to_object(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_to_defaultvalue(duk_context *ctx, duk_idx_t index, duk_int_t hint);
func (this Context) ToDefaultValue(index, hint int) {
	C.duk_to_defaultvalue(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_int_t(hint))
}

// void duk_to_primitive(duk_context *ctx, duk_idx_t index, duk_int_t hint);
func (this Context) ToPrimitive(index, hint int) {
	C.duk_to_primitive(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_int_t(hint))
}

// #define duk_to_buffer(ctx,index,out_size) duk_to_buffer_raw((ctx), (index), (out_size), DUK_BUF_MODE_DONTCARE)
func (this Context) ToBuffer(index int) []byte {
	return this.toBufferRaw(index, uint(BUF_MODE_DONTCARE))
}

// #define duk_to_fixed_buffer(ctx,index,out_size) duk_to_buffer_raw((ctx), (index), (out_size), DUK_BUF_MODE_FIXED)
func (this Context) ToFixedBuffer(index int) []byte {
	return this.toBufferRaw(index, uint(BUF_MODE_FIXED))
}

// #define duk_to_dynamic_buffer(ctx,index,out_size) duk_to_buffer_raw((ctx), (index), (out_size), DUK_BUF_MODE_DYNAMIC)
func (this Context) ToDynamicBuffer(index int) []byte {
	return this.toBufferRaw(index, uint(BUF_MODE_DYNAMIC))
}

/* safe variants of a few coercion operations */

// const char *duk_safe_to_lstring(duk_context *ctx, duk_idx_t index, duk_size_t *out_len);
// #define duk_safe_to_string(ctx,index) duk_safe_to_lstring((ctx), (index), NULL)
func (this Context) SafeToString(index int) string {
	var (
		ret     *C.char
		out_len C.duk_size_t
	)
	ret = C.duk_safe_to_lstring(unsafe.Pointer(this.ctx), C.duk_idx_t(index), &out_len)

	return C.GoStringN(ret, C.int(out_len))
}

/* --------------- */
/* Misc conversion */
/* --------------- */

// const char *duk_base64_encode(duk_context *ctx, duk_idx_t index);
func (this Context) Base64Encode(index int) string {
	return C.GoString(C.duk_base64_encode(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// void duk_base64_decode(duk_context *ctx, duk_idx_t index);
func (this Context) Base64Decode(index int) {
	C.duk_base64_decode(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// const char *duk_hex_encode(duk_context *ctx, duk_idx_t index);
func (this Context) HexEncode(index int) string {
	return C.GoString(C.duk_hex_encode(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// void duk_hex_decode(duk_context *ctx, duk_idx_t index);
func (this Context) HexDecode(index int) {
	C.duk_hex_decode(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// const char *duk_json_encode(duk_context *ctx, duk_idx_t index);
func (this Context) JsonEncode(index int) string {
	return C.GoString(C.duk_json_encode(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// void duk_json_decode(duk_context *ctx, duk_idx_t index);
func (this Context) JsonDecode(index int) {
	C.duk_json_decode(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

/* ------ */
/* Buffer */
/* ------ */

// void *duk_resize_buffer(duk_context *ctx, duk_idx_t index, duk_size_t new_size);
func (this Context) ResizeBuffer(index int, new_size uint) {
	C.duk_resize_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_size_t(new_size))
}

func (this Context) SetBuffer(index int, buf []byte) {
	ptr := C.duk_resize_buffer(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_size_t(len(buf)))
	C.memcpy(ptr, unsafe.Pointer(&buf[0]), C.size_t(len(buf)))
}

/* --------------- */
/* Property access */
/* --------------- */
// The basic function assumes key is on stack. The _string variant takes
// a C string as a property name, while the _index variant takes an array
// index as a property name (e.g. 123 is equivalent to the key "123").

// duk_int_t duk_get_prop(duk_context *ctx, duk_idx_t obj_index);
func (this Context) GetProp(obj_index int) int {
	return int(C.duk_get_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index)))
}

// duk_int_t duk_get_prop_string(duk_context *ctx, duk_idx_t obj_index, const char *key);
func (this Context) GetPropString(obj_index int, key string) int {
	cS_key := C.CString(key)
	defer C.free(unsafe.Pointer(cS_key))

	return int(C.duk_get_prop_string(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), cS_key))
}

// duk_int_t duk_get_prop_index(duk_context *ctx, duk_idx_t obj_index, duk_uarridx_t arr_index);
func (this Context) GetPropIndex(obj_index int, arr_index uint) int {
	return int(C.duk_get_prop_index(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_uarridx_t(arr_index)))
}

// duk_int_t duk_put_prop(duk_context *ctx, duk_idx_t obj_index);
func (this Context) PutProp(obj_index int) int {
	return int(C.duk_put_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index)))
}

// duk_int_t duk_put_prop_string(duk_context *ctx, duk_idx_t obj_index, const char *key);
func (this Context) PutPropString(obj_index int, key string) int {
	cS_key := C.CString(key)
	defer C.free(unsafe.Pointer(cS_key))

	return int(C.duk_put_prop_string(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), cS_key))
}

// duk_int_t duk_put_prop_index(duk_context *ctx, duk_idx_t obj_index, duk_uarridx_t arr_index);
func (this Context) PutPropIndex(obj_index int, arr_index uint) int {
	return int(C.duk_put_prop_index(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_uarridx_t(arr_index)))
}

// duk_int_t duk_del_prop(duk_context *ctx, duk_idx_t obj_index);
func (this Context) DelProp(obj_index int) int {
	return int(C.duk_del_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index)))
}

// duk_int_t duk_del_prop_string(duk_context *ctx, duk_idx_t obj_index, const char *key);
func (this Context) DelPropString(obj_index int, key string) int {
	cS_key := C.CString(key)
	defer C.free(unsafe.Pointer(cS_key))

	return int(C.duk_del_prop_string(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), cS_key))
}

// duk_int_t duk_del_prop_index(duk_context *ctx, duk_idx_t obj_index, duk_uarridx_t arr_index);
func (this Context) DelPropIndex(obj_index int, arr_index uint) int {
	return int(C.duk_del_prop_index(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_uarridx_t(arr_index)))
}

// duk_int_t duk_has_prop(duk_context *ctx, duk_idx_t obj_index);
func (this Context) HasProp(obj_index int) int {
	return int(C.duk_has_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index)))
}

// duk_int_t duk_has_prop_string(duk_context *ctx, duk_idx_t obj_index, const char *key);
func (this Context) HasPropString(obj_index int, key string) int {
	cS_key := C.CString(key)
	defer C.free(unsafe.Pointer(cS_key))

	return int(C.duk_has_prop_string(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), cS_key))
}

// duk_int_t duk_has_prop_index(duk_context *ctx, duk_idx_t obj_index, duk_uarridx_t arr_index);
func (this Context) HasPropIndex(obj_index int, arr_index uint) int {
	return int(C.duk_has_prop_index(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_uarridx_t(arr_index)))
}

// void duk_def_prop(duk_context *ctx, duk_idx_t obj_index, duk_uint_t flags);
func (this Context) DefProp(obj_index int, flags uint) {
	C.duk_def_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_uint_t(flags))
}

// duk_int_t duk_get_global_string(duk_context *ctx, const char *key);
func (this Context) GetGlobalString(key string) int {
	cS_key := C.CString(key)
	defer C.free(unsafe.Pointer(cS_key))

	return int(C.duk_get_global_string(unsafe.Pointer(this.ctx), cS_key))
}

// duk_int_t duk_put_global_string(duk_context *ctx, const char *key);
func (this Context) PutGlobalString(key string) int {
	cS_key := C.CString(key)
	defer C.free(unsafe.Pointer(cS_key))

	return int(C.duk_put_global_string(unsafe.Pointer(this.ctx), cS_key))
}

/* ---------------- */
/* Object prototype */
/* ---------------- */

// void duk_get_prototype(duk_context *ctx, duk_idx_t index);
func (this Context) GetPrototype(index int) {
	C.duk_get_prototype(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_set_prototype(duk_context *ctx, duk_idx_t index);
func (this Context) SetPrototype(index int) {
	C.duk_set_prototype(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

/* ---------------- */
/* Object finalizer */
/* ---------------- */

// void duk_get_finalizer(duk_context *ctx, duk_idx_t index);
func (this Context) GetFinalizer(index int) {
	C.duk_get_finalizer(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

// void duk_set_finalizer(duk_context *ctx, duk_idx_t index);
func (this Context) SetFinalizer(index int) {
	C.duk_set_finalizer(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

/* ------------- */
/* Global object */
/* ------------- */

// void duk_set_global_object(duk_context *ctx);
func (this Context) SetGlobalObject() {
	C.duk_set_global_object(unsafe.Pointer(this.ctx))
}

/* ------------------------------ */
/* Duktape/C function magic value */
/* ------------------------------ */
/* DO NOT use these, as they are used for resolving duktape/c/go function call

// duk_int_t duk_get_magic(duk_context *ctx, duk_idx_t index);
func (this Context) GetMagic(index int) int {
	return int(C.duk_get_magic(unsafe.Pointer(this.ctx), C.duk_idx_t(index)))
}

// void duk_set_magic(duk_context *ctx, duk_idx_t index, duk_int_t magic);
func (this Context) SetMagic(index, magic int) {
	C.duk_set_magic(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_int_t(magic))
}

// duk_int_t duk_get_current_magic(duk_context *ctx);
func (this Context) GetCurrentMagic() int {
	return int(C.duk_get_current_magic(unsafe.Pointer(this.ctx)))
}
*/
/* -------------- */
/* Module helpers */
/* -------------- */
// put multiple function or constant properties

//TODO implement multiple C functions lists putting
// void duk_put_function_list(duk_context *ctx, duk_idx_t obj_index, const duk_function_list_entry *funcs);

// void duk_put_number_list(duk_context *ctx, duk_idx_t obj_index, const duk_number_list_entry *numbers);
func (this Context) PutNumberList(obj_index int, numbers map[string]float64) {
	obj_index = this.NormalizeIndex(obj_index)
	for key, val := range numbers {
		this.PushNumber(val)
		this.PutPropString(obj_index, key)
	}
}

/* ----------------- */
/* Object operations */
/* ----------------- */

// void duk_compact(duk_context *ctx, duk_idx_t obj_index);
func (this Context) Compact(obj_index int) {
	C.duk_compact(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index))
}

// void duk_enum(duk_context *ctx, duk_idx_t obj_index, duk_uint_t enum_flags);
func (this Context) Enum(obj_index int, enum_flags uint) {
	C.duk_enum(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_uint_t(enum_flags))
}

// duk_int_t duk_next(duk_context *ctx, duk_idx_t enum_index, duk_int_t get_value);
func (this Context) Next(enum_index int, get_value int) int {
	return int(C.duk_next(unsafe.Pointer(this.ctx), C.duk_idx_t(enum_index), C.duk_bool_t(get_value)))
}

/* ------------------- */
/* String manipulation */
/* ------------------- */

// void duk_concat(duk_context *ctx, duk_idx_t count);
func (this Context) Concat(count int) {
	C.duk_concat(unsafe.Pointer(this.ctx), C.duk_idx_t(count))
}

// void duk_join(duk_context *ctx, duk_idx_t count);
func (this Context) Join(count int) {
	C.duk_join(unsafe.Pointer(this.ctx), C.duk_idx_t(count))
}

//TODO implement if needed
// void duk_decode_string(duk_context *ctx, duk_idx_t index, duk_decode_char_function callback, void *udata);
// void duk_map_string(duk_context *ctx, duk_idx_t index, duk_map_char_function callback, void *udata);

// void duk_substring(duk_context *ctx, duk_idx_t index, duk_size_t start_char_offset, duk_size_t end_char_offset);
func (this Context) Substring(index, start_char_offset, end_char_offset int) {
	C.duk_substring(unsafe.Pointer(this.ctx), C.duk_idx_t(index), C.duk_size_t(start_char_offset), C.duk_size_t(end_char_offset))
}

// void duk_trim(duk_context *ctx, duk_idx_t index);
func (this Context) Trim(index int) {
	C.duk_trim(unsafe.Pointer(this.ctx), C.duk_idx_t(index))
}

//TODO implement if needed
// duk_codepoint_t duk_char_code_at(duk_context *ctx, duk_idx_t index, duk_size_t char_offset);

/* -------------------- */
/* Ecmascript operators */
/* -------------------- */

// duk_int_t duk_equals(duk_context *ctx, duk_idx_t index1, duk_idx_t index2);
func (this Context) Equals(index1, index2 int) int {
	return int(C.duk_equals(unsafe.Pointer(this.ctx), C.duk_idx_t(index1), C.duk_idx_t(index2)))
}

// duk_int_t duk_strict_equals(duk_context *ctx, duk_idx_t index1, duk_idx_t index2);
func (this Context) StrictEquals(index1, index2 int) int {
	return int(C.duk_strict_equals(unsafe.Pointer(this.ctx), C.duk_idx_t(index1), C.duk_idx_t(index2)))
}

/* ----------------------- */
/* Function (method) calls */
/* ----------------------- */

// void duk_call(duk_context *ctx, duk_idx_t nargs);
func (this Context) Call(nargs int) {
	C.duk_call(unsafe.Pointer(this.ctx), C.duk_idx_t(nargs))
}

// void duk_call_method(duk_context *ctx, duk_idx_t nargs);
func (this Context) CallMethod(nargs int) {
	C.duk_call_method(unsafe.Pointer(this.ctx), C.duk_idx_t(nargs))
}

// void duk_call_prop(duk_context *ctx, duk_idx_t obj_index, duk_idx_t nargs);
func (this Context) CallProp(obj_index, nargs int) {
	C.duk_call_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_idx_t(nargs))
}

// duk_int_t duk_pcall(duk_context *ctx, duk_idx_t nargs);
func (this Context) Pcall(nargs int) int {
	return int(C.duk_pcall(unsafe.Pointer(this.ctx), C.duk_idx_t(nargs)))
}

// duk_int_t duk_pcall_method(duk_context *ctx, duk_idx_t nargs);
func (this Context) PcallMethod(nargs int) int {
	return int(C.duk_pcall_method(unsafe.Pointer(this.ctx), C.duk_idx_t(nargs)))
}

// duk_int_t duk_pcall_prop(duk_context *ctx, duk_idx_t obj_index, duk_idx_t nargs);
func (this Context) PcallProp(obj_index, nargs int) int {
	return int(C.duk_pcall_prop(unsafe.Pointer(this.ctx), C.duk_idx_t(obj_index), C.duk_idx_t(nargs)))
}

// void duk_new(duk_context *ctx, duk_idx_t nargs);
func (this Context) New(nargs int) {
	C.duk_new(unsafe.Pointer(this.ctx), C.duk_idx_t(nargs))
}

//TODO implement calling a C function
// duk_int_t duk_safe_call(duk_context *ctx, duk_safe_call_function func, duk_idx_t nargs, duk_idx_t nrets);

/* ----------------- */
/* Thread management */
/* ----------------- */

/* -------------------------- */
/* Compilation and evaluation */
/* -------------------------- */

// duk_int_t duk_eval_raw(duk_context *ctx, const char *src_buffer, duk_size_t src_length, duk_uint_t flags);
func (this Context) EvalRaw(src_buffer string, flags uint) int {
	cS_src_buffer := C.CString(src_buffer)
	defer C.free(unsafe.Pointer(cS_src_buffer))

	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), cS_src_buffer, C.duk_size_t(len(src_buffer)), C.duk_uint_t(flags)))
}

// duk_int_t duk_compile_raw(duk_context *ctx, const char *src_buffer, duk_size_t src_length, duk_uint_t flags);
func (this Context) CompileRaw(src_buffer string, flags uint) int {
	cS_src_buffer := C.CString(src_buffer)
	defer C.free(unsafe.Pointer(cS_src_buffer))

	return int(C.duk_compile_raw(unsafe.Pointer(this.ctx), cS_src_buffer, C.duk_size_t(len(src_buffer)), C.duk_uint_t(flags)))
}

/* plain */
/* #define duk_eval(ctx)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 (void) duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL)) */
func (this Context) Eval(file_name string) {
	this.PushString(file_name)
	C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL))
}

/* #define duk_eval_noresult(ctx)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 (void) duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL | DUK_COMPILE_NORESULT)) */
func (this Context) EvalNoResult(file_name string) {
	this.PushString(file_name)
	C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL|COMPILE_NORESULT))
}

/* #define duk_peval(ctx)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL | DUK_COMPILE_SAFE)) */
func (this Context) Peval(file_name string) int {
	this.PushString(file_name)
	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL|COMPILE_SAFE)))
}

/* #define duk_peval_noresult(ctx)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL | DUK_COMPILE_SAFE | DUK_COMPILE_NORESULT)) */
func (this Context) PevalNoResult(file_name string) int {
	this.PushString(file_name)
	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL|COMPILE_SAFE|COMPILE_NORESULT)))
}

/* #define duk_compile(ctx,flags)  \
((void) duk_compile_raw((ctx), NULL, 0, (flags))) */
func (this Context) Compile(flags uint) {
	C.duk_compile_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(flags))
}

/* #define duk_pcompile(ctx,flags) int  \
(duk_compile_raw((ctx), NULL, 0, (flags) | DUK_COMPILE_SAFE)) */
func (this Context) Pcompile(flags uint) int {
	return int(C.duk_compile_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(flags|COMPILE_SAFE)))
}

/* string */
/* lstring */
/* #define duk_eval_lstring(ctx,buf,len)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 (void) duk_eval_raw((ctx), buf, len, DUK_COMPILE_EVAL | DUK_COMPILE_NOSOURCE)) */
func (this Context) EvalString(src string, file_name string) {
	this.PushString(file_name)

	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	C.duk_eval_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(COMPILE_EVAL|COMPILE_NOSOURCE))
}

/* #define duk_eval_lstring_noresult(ctx,buf,len)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 (void) duk_eval_raw((ctx), buf, len, DUK_COMPILE_EVAL | DUK_COMPILE_NOSOURCE | DUK_COMPILE_NORESULT)) */
func (this Context) EvalStringNoResult(src string, file_name string) {
	this.PushString(file_name)

	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	C.duk_eval_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(COMPILE_EVAL|COMPILE_NOSOURCE|COMPILE_NORESULT))
}

/* #define duk_peval_lstring(ctx,buf,len)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 duk_eval_raw((ctx), buf, len, DUK_COMPILE_EVAL | DUK_COMPILE_NOSOURCE | DUK_COMPILE_SAFE)) */
func (this Context) PevalString(src string, file_name string) int {
	this.PushString(file_name)

	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(COMPILE_EVAL|COMPILE_NOSOURCE|COMPILE_SAFE)))
}

/* #define duk_peval_lstring_noresult(ctx,buf,len)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 duk_eval_raw((ctx), buf, len, DUK_COMPILE_EVAL | DUK_COMPILE_SAFE | DUK_COMPILE_NOSOURCE | DUK_COMPILE_NORESULT)) */
func (this Context) PevalStringNoResult(src string, file_name string) int {
	this.PushString(file_name)

	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(COMPILE_EVAL|COMPILE_SAFE|COMPILE_NOSOURCE|COMPILE_NORESULT)))
}

/* #define duk_compile_lstring(ctx,flags,buf,len)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 (void) duk_compile_raw((ctx), buf, len, (flags) | DUK_COMPILE_NOSOURCE)) */
func (this Context) CompileString(flags uint, src string, file_name string) {
	this.PushString(file_name)

	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	C.duk_compile_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(flags|COMPILE_NOSOURCE))
}

/* #define duk_compile_lstring_filename(ctx,flags,buf,len)  \
((void) duk_compile_raw((ctx), buf, len, (flags) | DUK_COMPILE_NOSOURCE)) */
func (this Context) CompileStringFilename(flags uint, src string) int {
	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	return int(C.duk_compile_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(flags|COMPILE_NOSOURCE)))
}

/* #define duk_pcompile_lstring(ctx,flags,buf,len)  \
((void) duk_push_string((ctx), (const char *) (__FILE__)), \
 duk_compile_raw((ctx), buf, len, (flags) | DUK_COMPILE_SAFE | DUK_COMPILE_NOSOURCE)) */
func (this Context) PcompileString(flags uint, src string, file_name string) int {
	this.PushString(file_name)

	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))

	return int(C.duk_compile_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(flags|COMPILE_SAFE|COMPILE_NOSOURCE)))
}

/* #define duk_pcompile_lstring_filename(ctx,flags,buf,len)  \
(duk_compile_raw((ctx), buf, len, (flags) | DUK_COMPILE_SAFE | DUK_COMPILE_NOSOURCE)) */
func (this Context) PcompileStringFilename(flags uint, src string) int {
	cS_src := C.CString(src)
	defer C.free(unsafe.Pointer(cS_src))
	return int(C.duk_compile_raw(unsafe.Pointer(this.ctx), cS_src, C.duk_size_t(len(src)), C.duk_uint_t(flags|COMPILE_SAFE|COMPILE_NOSOURCE)))
}

/* file */
/* #define duk_eval_file(ctx,path)  \
((void) duk_push_string_file_raw((ctx), (path), 0), \
 (void) duk_push_string((ctx), (path)), \
 (void) duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL)) */
func (this Context) EvalFile(path string) {
	this.PushStringFileRaw(path, 0)
	this.PushString(path)
	C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL))
}

/* #define duk_eval_file_noresult(ctx,path)  \
((void) duk_push_string_file_raw((ctx), (path), 0), \
 (void) duk_push_string((ctx), (path)), \
 (void) duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL | DUK_COMPILE_NORESULT)) */
func (this Context) EvalFileNoResult(path string) {
	this.PushStringFileRaw(path, 0)
	this.PushString(path)
	C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL|COMPILE_NORESULT))
}

/* #define duk_peval_file(ctx,path)  \
((void) duk_push_string_file_raw((ctx), (path), DUK_STRING_PUSH_SAFE), \
 (void) duk_push_string((ctx), (path)), \
 duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL | DUK_COMPILE_SAFE)) */
func (this Context) PevalFile(path string) int {
	this.PushStringFileRaw(path, STRING_PUSH_SAFE)
	this.PushString(path)
	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL|COMPILE_SAFE)))
}

/* #define duk_peval_file_noresult(ctx,path)  \
((void) duk_push_string_file_raw((ctx), (path), DUK_STRING_PUSH_SAFE), \
 (void) duk_push_string((ctx), (path)), \
 duk_eval_raw((ctx), NULL, 0, DUK_COMPILE_EVAL | DUK_COMPILE_SAFE | DUK_COMPILE_NORESULT)) */
func (this Context) PevalFileNoResult(path string) int {
	this.PushStringFileRaw(path, STRING_PUSH_SAFE)
	this.PushString(path)
	return int(C.duk_eval_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(COMPILE_EVAL|COMPILE_SAFE|COMPILE_NORESULT)))
}

/* #define duk_compile_file(ctx,flags,path)  \
((void) duk_push_string_file_raw((ctx), (path), 0), \
 (void) duk_push_string((ctx), (path)), \
 (void) duk_compile_raw((ctx), NULL, 0, (flags))) */
func (this Context) CompileFile(flags uint, path string) {
	this.PushStringFileRaw(path, 0)
	this.PushString(path)
	C.duk_compile_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(flags))
}

/* #define duk_pcompile_file(ctx,flags,path)  \
((void) duk_push_string_file_raw((ctx), (path), DUK_STRING_PUSH_SAFE), \
 (void) duk_push_string((ctx), (path)), \
 duk_compile_raw((ctx), NULL, 0, (flags) | DUK_COMPILE_SAFE)) */
func (this Context) PcompileFile(flags uint, path string) int {
	this.PushStringFileRaw(path, STRING_PUSH_SAFE)
	this.PushString(path)
	return int(C.duk_compile_raw(unsafe.Pointer(this.ctx), nil, 0, C.duk_uint_t(flags|COMPILE_SAFE)))
}

/* ------- */
/* Logging */
/* ------- */

// void duk_log(duk_context *ctx, duk_int_t level, const char *fmt, ...);
// void duk_log_va(duk_context *ctx, duk_int_t level, const char *fmt, va_list ap);
func (this Context) Log(level int, text string) {
	cS_text := C.CString(text)
	defer C.free(unsafe.Pointer(cS_text))

	C.__duk_log(unsafe.Pointer(this.ctx), C.duk_int_t(level), cS_text)
}

/* --------- */
/* Debugging */
/* --------- */

// void duk_push_context_dump(duk_context *ctx);
func (this Context) PushContextDump() {
	C.duk_push_context_dump(unsafe.Pointer(this.ctx))
}

func (this Context) DumpContextStdout() {
	this.PushContextDump()
	println(this.SafeToString(-1))
	this.Pop()
}
