window.initGRPCForm = function(services, svcDescs, mtdDescs, invokeURI, metadataURI, debug, headers) {

    var descriptionsShown = false;
    var requestForm = $("#grpc-request-form");

    function formServiceSelected(callback) {
        var svcName = $("#grpc-service").val();
        var svcDesc = svcDescs[svcName];
        var methods = services[svcName];
        var svcDescEnd = "";
        if (svcDesc) {
            svcDescEnd = '   // ... ' + (methods.length - 1) + ' more methods ...\n}';
        }
        $("#grpc-service-description").text(svcDesc);
        $("#grpc-service-description-end").text(svcDescEnd);

        var methodList = $("#grpc-method");
        methodList.empty();
        for (var i = 0; i < methods.length; i++) {
            m = methods[i];
            methodList.append($("<option>", {value: m, text: m}));
        }
        $("#grpc-method option:first-of-type").select();
        // implicit selection of first element does not
        // generate a change event, so we have to do this
        formMethodSelected(callback);
    }

    function formMethodSelected(callback) {
        var service = $("#grpc-service").val();
        var method = $("#grpc-method").val();
        var fullMethod = service + "." + method;

        var mtdDesc = mtdDescs[fullMethod];
        $("#grpc-method-description").text(mtdDesc);

        requestForm.empty();
        // disable the invoke button until we get the schema
        resetInvoke(false);

        $.ajax(metadataURI + "?method=" + fullMethod)
            .done(function(data) {
                buildRequestForm(data);
                callback?.();
            })
            .fail(function(data, status) {
                alert("Unexpected error: " + status);
                if (debug) {
                    console.trace(data);
                }
            });
    }

    function resetInvoke(enabled) {
        // Reset disables the response tab and navigates to
        // the request tab. Based on the given argument, it
        // can enable or disable the "invoke" buttons.
        var t = $("#grpc-request-response");
        if (t.tabs("option", "active") === 2) {
            t.tabs("option", "active", 0);
        }
        t.tabs("disable", 2);
        $(".grpc-invoke").prop("disabled", !enabled);
    }

    /*
     * Request Form
     *
     * The request form, requestForm inside this closure, has several pieces of
     * data attached to it (using jQuery.data(...)):
     *  1. 'schema': This is the schema response from the server for this RPC
     *     method. It contains definitions for all messages and enums and also
     *     identifies the "root" message: the type of the actual RPC request.
     *  2. 'request': This is the request object. As the form is edited, this
     *     object is edited, too. This is the object that will be sent to the
     *     server when the RPC is invoked. It can also be edited as raw JSON
     *     via the "raw request" tab in the UI.
     *  3. 'root': This is the root for a hierarchy of input controls. See below
     *     for more info on inputs. Inputs represent a tree, corresponding to
     *     the same structure as the request object. The root input corresponds
     *     to the main request object. Various inputs underneath in the
     *     hierarchy correspond to data elements within the request object.
     *
     * Request Input Hierarchy
     *
     * A request "input" is an object that provides the following interface:
     *  * parent: Inputs are a hierarchy. The parent property returns the
     *    input's parent in the hierarchy. The root input will have a null
     *    parent.
     *  * children: Returns an object of children. This object may be an array,
     *    if children are identified by array index, or an object, if children
     *    are identified by a property name. Each child is also an input.
     *  * value: This should be some descendant of the 'root' value for for the
     *    form.
     *  * onChange(ch, path, val): Notifies an input that one of its descendants
     *    has changed. The supplied ch is the child input from which the event
     *    came, val is the new value, and path is the path from the given ch
     *    input to the leaf input that actually changed. Note that the path is
     *    reversed, so that the first element represents the path element for
     *    the leaf and the last element represents which immediate child of ch
     *    leads to the leaf.
     *  * onAdd(ch, path, val): Notifies an input that a new entry has been
     *    added to one of its descendants. This is used when an element is added
     *    to an array or map value. The arguments are the same as for onChange.
     *  * onDelete(ch, path): Notifies an input that an entry has been removed
     *    from one of its descendants. This is used when an element is removed
     *    from an array or map value. The arguments are the same as for onChange
     *    except that no val is provided since it is unnecessary.
     *  * setValue(val): Sets the input's current value and, if it has changed,
     *    notifies ancestors of the change by calling onChange.
     *
     * Sync'ing
     *
     * There are three representations of the request to keep in sync: (1) the
     * raw JSON, (2) the request object, and (3) the form which contains a
     * hierarchy of inputs. There are only two ways that the request can be
     * modified: by editing the raw JSON or interacting with inputs on the form.
     *
     * If the raw JSON is edited, it is re-parsed to re-create the request
     * object and then the entire input hierarchy is re-created from that
     * object. This keeps everything in sync with changes to JSON.
     *
     * In the other direction, as inputs are edited, their paths are used to
     * locate elements in the request object, for surgical changes to keep the
     * object in sync. The request object is then re-serialized to JSON, to make
     * sure the raw JSON view is also in sync.
     */

    // Builds a new request form for the given RPC schema.
    function buildRequestForm(schema) {
        // build form
        requestForm.data("schema", schema);
        var requestObj = getInitialMessageValue(schema.requestType);
        if (schema.requestStream) {
            requestObj = [requestObj];
        }

        try {
            rebuildRequestForm(requestObj, true);
        } catch (e) {
            var msg = e.message;
            if (isUnset(msg)) {
                msg = e.toString();
            }
            alert(msg);
            if (debug) {
                console.trace(e);
            }
            // disable invoke button
            resetInvoke(true);
            return;
        }

        // set raw request text
        updateJSONRequest(requestObj);

        // enable the invoke button
        resetInvoke(true);

        // update the message list
        updateMessageList(schema.messageTypes);
    }

    function updateMessageList(messageTypes) {
        // TODO(jh): ideally, this would include all message types known by the
        // server, not just the ones for the current RPC method's schema.

        // start with well-known types
        var messageNamesUnique = {
            "google.protobuf.Int32Value": true,
            "google.protobuf.UInt32Value": true,
            "google.protobuf.Int64Value": true,
            "google.protobuf.UInt64Value": true,
            "google.protobuf.DoubleValue": true,
            "google.protobuf.FloatValue": true,
            "google.protobuf.StringValue": true,
            "google.protobuf.BytesValue": true,
            "google.protobuf.BoolValue": true,
            "google.protobuf.Timestamp": true,
            "google.protobuf.Duration": true,
            "google.protobuf.Value": true,
            "google.protobuf.Struct": true,
            "google.protobuf.ListValue": true,
        };
        for (var p in messageTypes) {
            // add all message types from current schema
            // (but not Any: JSON format for Any doesn't allow it hold itself
            if (messageTypes.hasOwnProperty(p) && p !== "google.protobuf.Any") {
                messageNamesUnique[p] = true;
            }
        }

        // now that we have a set of unique names, we can create sorted array
        var messageNames = [];
        for (p in messageNamesUnique) {
            if (messageNamesUnique.hasOwnProperty(p)) {
                messageNames.push(p);
            }
        }
        messageNames.sort();

        var list = $("#grpc-message-types");
        list.empty();
        for (var i = 0; i < messageNames.length; i++) {
            var opt = $('<option>');
            opt.attr('value', messageNames[i]);
            opt.text(messageNames[i]);
            list.append(opt);
        }
    }

    // Rebuilds the request form so it represents the given request value. If
    // allowMissing is true, the given newRequest can omit required fields, and
    // they will be populated with defaults as the form is built. If it is
    // false, an error will be thrown if the field is required but not present
    // in the given request value. If there are any errors in the request values
    // (wrong type, invalid, etc), an error is thrown.
    function rebuildRequestForm(newRequest, allowMissing) {
        requestForm.data("request", newRequest);
        var schema = requestForm.data("schema");
        // synthesize field definition that describes request type
        var fldDef = {
            name: "",
            type: schema.requestType,
            oneOfFields: [],
            isMessage: true,
            isEnum: false,
            isArray: schema.requestStream,
            isMap: false,
            isRequired: false,
            defaultVal: null,
            description: "",
        };
        requestForm.empty();

        // put it in table with label
        var div = $("<div>");
        div.addClass("input_container one-of-2 one-of-3 one-of-4 one-of-5");
        div.attr('id', 'root');
        requestForm.append(div);

        var table = $('<table>');
        div.append(table);

        var requestType = schema.requestType;
        var row = $('<tr>');
        table.append(row);
        var cell = $('<th>');
        cell.attr('colspan', '3');
        cell.text(requestType);
        if (schema.requestStream) {
            cell.prepend('<em>stream</em> ');
        }
        row.append(cell);

        requestForm.data("root", addElementToForm(schema, table, newRootInput(), 0, newRequest, allowMissing, fldDef));
    }

    var undefined = [][1];

    function isUnset(v) {
        return v === null || typeof v === 'undefined';
    }

    /*
     * The various add*ToForm functions add inputs to a table in the request
     * form. Each has a similar signature, accepting the following values:
     *
     *  * schema: the RPC schema, which contains definitions for all relevant
     *    messages and enums
     *  * container: the HTML element in which the input will be rendered. This
     *    will usually be a DIV or a TABLE.
     *  * parent: the parent input.
     *  * pathLen: the length of the path from the root of the request object to
     *    the value being added to the form. This is used for rendering/styling
     *    borders, so that groups of inputs are clear in the resulting HTML.
     *  * value: the value of the input to render and add to the form.
     *  * allowMissing: if true, if the given value has missing components (e.g.
     *    requiredFields that are not present), they will be assigned default
     *    values and construction of the form will proceed.
     *  * fld: the field definition for the value being rendered. This contains
     *    information about the value's formal type per the protocol.
     *
     *  Each such function returns the new child input. It is the function's
     *  responsibility to add its HTML elements to the given container. However,
     *  it is NOT the function's responsibility to add the returned input to the
     *  given parent (the caller will manage that).
     *
     *  The pathLen parameter is used to add CSS classes to containers, which
     *  can then be styled to provide visual indicators of groupings in the
     *  hierarchy. All messages, arrays, and maps are considered containers.
     *  The CSS classes added are in the form "one-of-3", "three-of-5", etc.
     *  This provides CSS authors a variety of ways to distinguish adjacent
     *  groups (e.g. an array inside of a message can use a different style,
     *  such as different border, then the message that contains it). The
     *  denominators used ("4" in class name "two-of-4") are 2 through 5. So
     *  the view could be rendered with alternating borders (using a denominator
     *  of 2) or with borders and backgrounds that rotate through 5 different
     *  styles (using a denominator of 5).
     */

    // Adds a request input to the given parent element for the field described
    // by fld. Returns the new request input. The path argument is the path to
    // the field from the root element (will be empty if fld represents the root
    // element). The value argument is the value in the request object. It will
    // be a composite value (object/map or array) for non-leaf inputs.
    //
    // This operation is recursive, since fld and value could represent a
    // composite input. This function will be called again, recursively, to
    // populate the children, and so on to the leaf inputs.
    //
    // This function does not handle oneofs. Those are only valid as direct
    // children of a message, and thus addMessageToForm handles them.
    function addElementToForm(schema, container, parent, pathLen, value, allowMissing, fld) {
        if (fld.isMap) {
            return addMapToForm(schema, container, parent, pathLen, value, allowMissing, fld);
        }

        if (fld.isArray) {
            return addArrayToForm(schema, container, parent, pathLen, value, allowMissing, "repeated " + fld.type, fld);
        }

        if (fld.type === "google.protobuf.ListValue") {
            // treat ListValue as if it were 'repeated Value'
            var elemType = {
                name: fld.name,
                protoName: fld.protoName,
                type: "google.protobuf.Value",
                oneOfFields: [],
                isMessage: true,
                isEnum: false,
                isArray: true,
                isMap: false,
                isRequired: false,
                defaultVal: null,
                description: fld.description,
            };
            return addArrayToForm(schema, container, parent, pathLen, value, allowMissing, fld.type, elemType);
        }

        if (fld.isMessage && !isWellKnown(fld.type)) {
            return addMessageToForm(schema, container, parent, pathLen, value, allowMissing, fld);
        }

        if (fld.type === 'google.protobuf.Any') {
            return addAnyToForm(schema, container, parent, pathLen, value, allowMissing, fld);
        }

        // Remaining cases are non-composite inputs. So we may need to create
        // a table row and cell (if container is the root table).
        container = maybeMakeInputContainer(container, pathLen);

        if (fld.isEnum) {
            return addEnumToForm(schema, container, parent, value, fld);
        }

        switch (fld.type) {
            case "int32": case "sint32": case "sfixed32": case "google.protobuf.Int32Value":
            return addIntToForm(container, parent, value, fld, MIN_INT32, MAX_INT32);

            case "uint32": case "fixed32": case "google.protobuf.UInt32Value":
            return addIntToForm(container, parent, value, fld, 0, MAX_UINT32);

            case "int64": case "sint64": case "sfixed64": case "google.protobuf.Int64Value":
            return addStringIntToForm(container, parent, value, fld, MIN_INT64, MAX_INT64);

            case "uint64": case "fixed64": case "google.protobuf.UInt64Value":
            return addStringIntToForm(container, parent, value, fld, "0", MAX_UINT64);

            case "double": case "float": case "google.protobuf.DoubleValue": case "google.protobuf.FloatValue":
            // TODO(jh): should 32-bit floats get extra range checks
            // (in case user is trying to use silly large magnitude
            // that will just get converted to "infinity" for 32-bit)?
            return addDoubleToForm(container, parent, value, fld);

            case "string": case "google.protobuf.StringValue":
            return addStringToForm(container, parent, value, fld);

            case "bytes": case "google.protobuf.BytesValue":
            return addBytesToForm(container, parent, value, fld);

            case "bool": case "google.protobuf.BoolValue":
            return addBoolToForm(container, parent, value, fld);

            case "google.protobuf.Timestamp":
                return addTimestampToForm(container, parent, value, fld);

            case "google.protobuf.Duration":
                return addDurationToForm(container, parent, value, fld);

            case "google.protobuf.Value": case "google.protobuf.Struct":
            return addJSONToForm(container, parent, value, fld, fld.type === "google.protobuf.Struct");

            default:
                throw new Error('Unable to handle non-message non-enum type ' + fld.type);
        }
    }

    // Adds a map value to the form.
    //
    // Maps are rendered as arrays of key+value pairs. This is also how they are
    // represented in the schema metadata (since that is how protobuf represents
    // maps: as repeated messages, where each message has a 'key' and a 'value'
    // field).
    function addMapToForm(schema, container, parent, pathLen, value, allowMissing, fld) {
        var mapEntryFields = schema.messageTypes[fld.type];
        var mapType = "map<" + mapEntryFields[0].type + "," + mapEntryFields[1].type + ">";

        if (isUnset(value)) {
            value = {};
        } else if (typeof value !== 'object' || value instanceof Array) {
            throw typeError(mapType, value, "object");
        }

        // Map inputs are different than others: their children and value
        // fields are not used. Instead the map uses an array input as its
        // source of children (created below), which is assigned to the
        // map input's "asArray" property. The map input further tracks
        // each array element's "key" value in its "childKeys" property.
        // When it receives notices of change from the array input, it must
        // transform the notifications so they refer to properties of the
        // map value, not the underlying array.
        var input = new Input(parent);
        input.childKeys = [];

        var arrayVal = [];
        for (k in value) {
            if (value.hasOwnProperty(k)) {
                var entry = {};
                entry[mapEntryFields[0].name] = k;
                entry[mapEntryFields[1].name] = value[k];
                arrayVal.push(entry);
                input.childKeys.push(k);
            }
        }

        input.asArray = addArrayToForm(schema, container, input, pathLen, arrayVal, allowMissing, mapType, fld);

        // adapt value form array to map by overriding change handlers
        input.doOnChange = function(child, revPath, val, nextFunc) {
            // translate child index to map key
            var index = + revPath[revPath.length-1];
            var key = this.childKeys[index];
            if (revPath.length === 2 && revPath[0] === 'key') {
                // key was changed! check that all keys are unique
                var newKey = val;

                // TODO: look for dupes for all keys (it is possible that
                // this change is fixing a duplicate key issue by changing
                // the other key, instead of the key that generated original
                // error...)
                for (var i = 0; i < this.childKeys.length; i++) {
                    if (i !== index && this.childKeys[i] === newKey) {
                        throw new Error('key "' + newKey + '" is a duplicate');
                    }
                }

                var entryVal = this.asArray.children[index]['value'].value;
                this.parent.onAdd(this, [newKey], entryVal);
                if (!isUnset(key)) {
                    this.parent.onDelete(this, [key]);
                }
                this.childKeys[index] = newKey;
                return;
            }

            if (isUnset(key)) {
                // not yet in model, skip
                return;
            }

            // otherwise, replace last two path elements
            // (array index and 'value') with the map key
            revPath.splice(revPath.length-1, 1);
            revPath[revPath.length-1] = key;

            this.parent[nextFunc].call(this.parent, this, revPath, val);
        };
        input.onAdd = function(child, revPath, val) {
            if (revPath.length <= 1) {
                // We are adding an entry to the map. We use an undefined
                // key to means it's not in main request model yet
                // (we leave it that way and don't call parent to let user
                // edit key field first, in case default value collides
                // with existing key).
                this.childKeys.push(undefined);
            } else {
                this.doOnChange(child, revPath, val, 'onAdd');
            }
        };
        input.onChange = function(child, revPath, val) {
            this.doOnChange(child, revPath, val, 'onChange');
        };
        input.onDelete = function(child, revPath) {
            // translate child index to map key
            var index = revPath[revPath.length-1];
            var key = this.childKeys[index];
            if (isUnset(key)) {
                // never added to model, so nothing to delete
                return;
            }

            if (revPath.length === 1) {
                // deleting entire entry
                this.parent.onDelete(this, [key]);
                this.childKeys.splice(index, 1);
                return;
            }

            // otherwise, replace last two path elements
            // (array index and 'value') with the map key
            revPath.splice(revPath.length-1, 1);
            revPath[revPath.length-1] = key;
            this.parent.onDelete(this, revPath);
        };

        return input;
    }

    // Adds an array value to the form.
    //
    // Arrays result in a table of inputs, where each row is an element in the
    // array value. The table includes buttons for adding and removing rows.
    function addArrayToForm(schema, container, parent, pathLen, value, allowMissing, typeName, fld) {
        if (isUnset(value)) {
            value = [];
        } else if (!(value instanceof Array)) {
            throw typeError(typeName, value, "array");
        }

        var table = makeTableContainer(container, pathLen);
        table.addClass('grpc-request-table');
        var children = [];
        var input = new Input(parent, children, value);

        var elementFld = {
            name: fld.name,
            protoName: fld.protoName,
            type: fld.type,
            oneOfFields: [],
            isMessage: fld.isMessage,
            isEnum: fld.isEnum,
            isArray: false,
            isMap: false,
            isMapEntry: fld.isMap,
            isRequired: false
        };

        for (var i = 0; i < value.length; i++) {
            var elementVal = value[i];

            var row = newArrayRow(input);
            table.append(row);
            cell = $('<td>');
            row.append(cell);

            var child = addElementToForm(schema, cell, input, pathLen+1, elementVal, allowMissing, elementFld);
            child.row = row;
            children.push(child);
        }

        row = $('<tr>');
        table.append(row);
        var cell = $('<td>');
        cell.addClass('array_button');
        var button = $('<button>');
        button.addClass('add');
        button.text('+');
        cell.append(button);
        row.append(cell);
        cell = $('<td>');
        cell.html('<span class="add-row-label">Add item</span>');
        row.append(cell);

        button.click(function() {
            // add a new element to the array
            var tr = $(this).closest('tr');
            var table = tr.closest('table');
            var elementVal = getInitialValue(schema, elementFld);

            var newRow = newArrayRow(input);
            table.append(newRow);
            var cell = $('<td>');
            newRow.append(cell);

            var child = addElementToForm(schema, cell, input, pathLen+1, elementVal, true, elementFld);
            child.row = newRow;
            children.push(child);

            tr.before(newRow);

            input.onAdd(child, [], elementVal);

            if (fld.isMap) {
                // put cursor in the key field (it likely needs to be
                // edited immediately to make sure keys are all unique)

                // TODO(jh): This probably belongs in the map input's onAdd
                // callback. But that doesn't have convenient access to the
                // actual TR element (which is the context for finding the
                // first input)
                var inputs = $('input,textarea', newRow);
                if (inputs.length > 0) {
                    var inp = $(inputs[0]);
                    // TODO(jh): this is kind of annoying if the element is
                    // already in the viewport -- it would be nice to detect
                    // that and skip the scrolling for those cases
                    $([document.documentElement, document.body]).scrollTop(inp.offset().top-30);
                    inp.focus();
                }
            }
        });

        return input;
    }

    function newArrayRow(input) {
        var row = $('<tr>');
        row.addClass('array_element');
        var cell = $('<td>');
        cell.addClass('array_button');
        var button = $('<button>');
        button.addClass('delete');
        button.text('X');
        cell.append(button);
        row.append(cell);
        button.click(function() {
            var children = input.children;
            var removed;
            $(this).closest('tr').remove();
            for (var idx = 0; idx < children.length; idx++) {
                if (children[idx].row === row) {
                    removed = children[idx];
                    break;
                }
            }
            input.onDelete(removed, []);
            children.splice(idx, 1);
        });
        return row;
    }

    // Adds a message value to the form.
    //
    // Message values are rendered as a table, where each row is a field.
    //
    // Other than fields in a one-of, all non-repeated fields have a cell with
    // an indicator of whether the field is present. If the field is required,
    // it is always present, and this cell contains an asterisk "*". Otherwise,
    // this cell contains a checkbox, for marking the field as present (checked)
    // or absent (unchecked).
    //
    // Messages, like repeated fields (maps and arrays), can be rendered with a
    // border, to visually group their child fields.
    function addMessageToForm(schema, container, parent, pathLen, value, allowMissing, fld) {
        if (typeof value !== 'object' || value instanceof Array) {
            throw typeError(fld.type, value, "object");
        }

        var fields = schema.messageTypes[fld.type];

        value = canonicalizeFields(value, fld.type, fields);

        // create table of child inputs, one for each field

        var table = makeTableContainer(container, pathLen);
        var children = {};
        var input = new Input(parent, children, value);

        for (var i = 0; i < fields.length; i++) {
            var currField = fields[i];

            var row = $('<tr>');
            row.addClass('message_field');
            table.append(row);
            var cell = makeFieldLabelCell(currField, schema);
            row.append(cell);

            if (isOneOf(currField)) {
                cell = $('<td>');
                cell.attr('colspan', 2);
                row.append(cell);
                var oneOfChildren = addOneOfFieldsToForm(schema, cell, input, pathLen, value, allowMissing, currField);
                $.extend(children, oneOfChildren);
                continue
            }

            var child = null;
            var fldVal = value[currField.name];

            if (currField.isArray || currField.isMap) {
                cell = $('<td>');
                cell.attr('colspan', 2);
                row.append(cell);
                if (isUnset(fldVal)) {
                    if (currField.isMap) {
                        fldVal = {};
                    } else {
                        fldVal = [];
                    }
                    value[currField.name] = fldVal;
                }
                child = addElementToForm(schema, cell, input, pathLen+1, fldVal, allowMissing, currField);
                children[currField.name] = child;

            } else {
                var required = fld.isMapEntry || currField.isRequired;
                var checkbox = undefined;
                if (!fld.isMapEntry) {
                    cell = $('<td>');
                    cell.addClass('toggle_presence');
                    if (currField.isRequired) {
                        cell.html('<span class="required">*</span>');
                    } else {
                        checkbox = $('<input>');
                        checkbox.attr('type', 'checkbox');
                        if (!isUnset(fldVal)) {
                            checkbox.prop('checked', true);
                        }
                        cell.append(checkbox);
                    }
                    row.append(cell);
                }

                cell = $('<td>');
                if (isUnset(fldVal) && required) {
                    if (allowMissing) {
                        fldVal = getInitialValue(schema, currField);
                        value[currField.name] = fldVal;
                    } else {
                        throw new Error("value for required field " + currField.name + " is missing");
                    }
                }
                if (!currField.isRequired && currField.isMessage) {
                    child = addOptionalMessage(schema, cell, input, pathLen+1, fldVal, allowMissing, currField);
                } else {
                    child = addElementToForm(schema, cell, input, pathLen+1, fldVal, allowMissing, currField);
                }
                child.fld = currField;
                children[currField.name] = child;
                row.append(cell);

                if (!isUnset(checkbox)) {
                    (function(fld, cell) {
                        checkbox.change(function() {
                            var checkbox = $(this);
                            cell.empty();
                            var fldVal;
                            if (checkbox.prop('checked')) {
                                fldVal = getInitialValue(schema, fld);
                            } else {
                                fldVal = undefined;
                            }
                            var child;
                            if (fld.isMessage) {
                                child = addOptionalMessage(schema, cell, input, pathLen+1, fldVal, true, fld);
                            } else {
                                child = addElementToForm(schema, cell, input, pathLen+1, fldVal, true, fld);
                            }
                            child.fld = fld;
                            for (var i in children) {
                                if (children.hasOwnProperty(i) && children[i].fld === fld) {
                                    children[i] = child;
                                    break;
                                }
                            }

                            input.onChange(child, [], fldVal);
                        })
                    })(currField, cell);
                }
            }
        }

        if (fields.length === 0) {
            row = $('<tr>');
            table.append(row);
            cell = $('<td>');
            cell.addClass('empty_message');
            cell.attr('colspan', 3);
            row.append(cell);
            cell.text('No fields');
        } else {
            // make sure this input's value tracks changes below
            input.onChange = function(child, revPath, value) {
                pushChildPath(child, this.children, revPath);
                if (revPath.length === 1) {
                    // immediate child was changed; update our local value
                    this.value[revPath[0]] = value;
                }
                this.parent.onChange(this, revPath, value);
            }
        }

        return input;
    }

    // used to provide a unique name for all radio button groups
    var radioSeq = 0;

    function addOneOfFieldsToForm(schema, container, parent, pathLen, value, allowMissing, oneof) {
        var div = $('<div>');
        div.addClass('oneof');
        container.append(div);

        var table = $('<table>');
        var children = {};
        var foundValue = false;
        var selected = {};

        radio_name = 'grpc_form_' + radioSeq;
        radioSeq++;

        var clearPrevious = function() {
            if (!isUnset(selected.field)) {
                // clear old selection
                var otherCell;
                var children = parent.children;
                for (var j in children) {
                    if (children.hasOwnProperty(j) && children[j].hasOwnProperty("fld") && children[j].fld.name === selected.field) {
                        otherCell = children[j].cell;
                        break;
                    }
                }
                var otherFld;
                for (var k = 0; k < oneof.oneOfFields.length; k++) {
                    if (oneof.oneOfFields[k].name === selected.field) {
                        otherFld = oneof.oneOfFields[k];
                        break;
                    }
                }
                otherCell.empty();
                var otherChild;
                if (otherFld.isMessage) {
                    otherChild = addOptionalMessage(schema, otherCell, parent, pathLen+1, undefined, false, otherFld);
                } else {
                    otherChild = addElementToForm(schema, otherCell, parent, pathLen+1, undefined, false, otherFld);
                }
                otherChild.fld = otherFld;
                otherChild.cell = otherCell;

                children[j] = otherChild;

                parent.onChange(otherChild, [], undefined);
            }
        };

        for (var i = 0; i < oneof.oneOfFields.length; i++) {
            var fld = oneof.oneOfFields[i];
            var fldVal = undefined;
            var isPresent = false;
            if (!foundValue) {
                fldVal = value[fld.name];
                if (!isUnset(fldVal)) {
                    isPresent = true;
                    foundValue = true;
                }
            }

            var row = $('<tr>');
            row.addClass('message_field');
            table.append(row);
            row.append(makeFieldLabelCell(fld, schema));

            var cell = $('<td>');
            var checkbox = $('<input>');
            checkbox.attr('type', 'radio');
            checkbox.attr('name', radio_name);
            checkbox.attr('value', fld.name);
            if (isPresent) {
                checkbox.prop('checked', true);
                selected.field = fld.name;
            }
            cell.append(checkbox);
            row.append(cell);

            cell = $('<td>');
            row.append(cell);

            var child;
            if (fld.isMessage) {
                child = addOptionalMessage(schema, cell, parent, pathLen+1, fldVal, allowMissing, fld);
            } else {
                child = addElementToForm(schema, cell, parent, pathLen+1, fldVal, allowMissing, fld);
            }
            children[fld.name] = child;
            child.fld = fld;
            child.cell = cell;

            (function(fld, cell) {
                checkbox.change(function() {
                    clearPrevious();
                    selected.field = fld.name;
                    var fldVal = getInitialValue(schema, fld);
                    var checkbox = $(this);
                    cell.empty();
                    var child;
                    if (fld.isMessage) {
                        child = addOptionalMessage(schema, cell, parent, pathLen+1, fldVal, true, fld);
                    } else {
                        child = addElementToForm(schema, cell, parent, pathLen+1, fldVal, true, fld);
                    }
                    child.fld = fld;
                    child.cell = cell;

                    var children = parent.children;
                    for (var i in children) {
                        if (children.hasOwnProperty(i) && children[i].fld === fld) {
                            children[i] = child;
                            break;
                        }
                    }
                    parent.onChange(child, [], fldVal);
                });
            })(fld, cell);
        }

        row = $('<tr>');
        row.addClass('message_field');
        table.append(row);
        cell = $('<td>');
        cell.addClass('oneof_none');
        cell.text('None');
        row.append(cell);
        cell = $('<td>');
        checkbox = $('<input>');
        checkbox.attr('type', 'radio');
        checkbox.attr('name', radio_name);
        checkbox.attr('value', '');
        if (!foundValue) {
            checkbox.prop('checked', true);
        }
        checkbox.change(clearPrevious);
        cell.append(checkbox);
        row.append(cell);
        cell = $('<td>');
        row.append(cell);

        div.append(table);
        return children;
    }

    function canonicalizeFields(value, type, fields) {
        // newValue will be a copy of value, but with all fields populated using their
        // original (proto) name, not their JSON name
        var newValue = {};

        // process fields identified by JSON name
        for (var i = 0; i < fields.length; i++) {
            var currField = fields[i];
            if (isOneOf(currField)) {
                for (var j = 0; j < currField.oneOfFields.length; j++) {
                    var field = currField.oneOfFields[j];
                    if (value.hasOwnProperty(field.name)) {
                        newValue[field.name] = value[field.name];
                        delete value[field.name];
                    }
                }
            } else {
                if (value.hasOwnProperty(currField.name)) {
                    newValue[currField.name] = value[currField.name];
                    delete value[currField.name];
                }
            }
        }
        // now process any remaining that are identified by original proto name
        for (i = 0; i < fields.length; i++) {
            currField = fields[i];
            if (isOneOf(currField)) {
                for (j = 0; j < currField.oneOfFields.length; j++) {
                    field = currField.oneOfFields[j];
                    if (value.hasOwnProperty(field.protoName)) {
                        if (newValue.hasOwnProperty(field.name)) {
                            throw new Error("value for type " + type + " has redundant values: " + field.name + " and " + field.protoName);
                        }
                        newValue[field.name] = value[field.protoName];
                        delete value[field.protoName];
                    }
                }
            } else {
                if (value.hasOwnProperty(currField.protoName)) {
                    if (newValue.hasOwnProperty(currField.name)) {
                        throw new Error("value for type " + type + " has redundant values: " + currField.name + " and " + currField.protoName);
                    }
                    newValue[currField.name] = value[currField.protoName];
                    delete value[currField.protoName];
                }
            }
        }

        // if there are any remaining fields, they are not valid field names
        for (var p in value) {
            if (value.hasOwnProperty(p)) {
                throw new Error("value for type " + type + " has unrecognized field: " + p)
            }
        }

        // copy canonicalized fields back into value
        return $.extend(value, newValue)
    }

    function isOneOf(fld) {
        return fld.type === "oneof" && !fld.isMessage;
    }

    function makeTableContainer(container, pathLen) {
        if (pathLen === 0) {
            return container;
        }
        var div = $('<div>');
        // add classes
        div.addClass('input_container');
        for (var i = 2; i <= 5; i++) {
            var idx = (pathLen % i) + 1;
            div.addClass(numberName(idx) + "-of-" + i);
        }
        container.append(div);
        var table = $('<table>');
        div.append(table);
        return table;
    }

    function maybeMakeInputContainer(container, pathLen) {
        if (pathLen === 0) {
            // container will be the root table, so add a row and cell and
            // let that be input's container
            var row = $('<tr>');
            container.append(row);
            var cell = $('<td>');
            row.append(cell);
            return cell;
        }
        return container;
    }

    function makeFieldLabelCell(fld, schema) {
        var cell = $('<td>');
        cell.addClass('name');

        if (fld.isMap) {
            var mapEntryFields = schema.messageTypes[fld.type];
            cell.text('<' + mapEntryFields[0].type + ',' + mapEntryFields[1].type + '>');
            cell.prepend('<em>map</em>');
        } else {
            cell.text(fld.type);
            if (fld.isArray) {
                cell.prepend('<em>repeated</em> ');
            }
        }

        var labelName = $('<strong>');
        labelName.text(fld.protoName);
        cell.prepend($('<br>'));
        if (fld.description) {
            labelName.prop('title', fld.description);
            labelName.tooltip({
                classes: {
                    "ui-tooltip": "grpc-field-description"
                }
            });
        }
        cell.prepend(labelName);

        return cell;
    }

    function numberName(num) {
        switch (num) {
            case 1:
                return "one";
            case 2:
                return "two";
            case 3:
                return "three";
            case 4:
                return "four";
            case 5:
                return "five";
        }
    }

    function addEnumToForm(schema, container, parent, value, fld) {
        var enumVals = schema.enumTypes[fld.type];

        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        } else if (typeof value !== 'string' && typeof value !== 'number') {
            throw typeError(fld.type, value, "string or number");
        }

        var sel = $('<select>');
        if (disabled) {
            sel.prop('disabled', true);
        }

        var isNumber = typeof value === 'number';
        var found = false;
        for (var i = 0; i < enumVals.length; i++) {
            var opt = $('<option>');
            opt.attr('value', enumVals[i].name);
            opt.text(enumVals[i].name);
            if (isNumber && enumVals[i].num === value || !isNumber && enumVals[i].name === value) {
                found = true;
                opt.prop('selected', true);
            }
            sel.append(opt);
        }
        if (!found) {
            throw new Error('Field ' + fld.name + ' of type enum ' + fld.type + ' has no such value: ' + value);
        }

        container.append(sel);

        var input = new Input(parent, [], value);
        sel.change(function() {
            var v = sel.val();
            if (fld.type === "google.protobuf.NullValue" && v === "NULL") {
                v = null;
            }
            input.setValue(v);
        });
        return input;
    }

    function addIntToForm(container, parent, value, fld, min, max) {
        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        } else {
            if (typeof value !== 'number') {
                throw typeError(fld.type, value, "number");
            }
            if (!isInt(value)) {
                throw new Error("value for type " + fld.type + " is not an integer: " + value);
            }
            if (value < min) {
                throw new Error("value for type " + fld.type + " is less than allowed min: " + value + " < " + min);
            }
            if (value > max) {
                throw new Error("value for type " + fld.type + " is greater than allowed max: " + value + " > " + max);
            }
        }

        var inp = $('<input>');
        inp.attr('type', 'text');
        inp.attr('size', 28);
        inp.attr('value', "" + value);
        if (disabled) {
            inp.prop('disabled', true);
        }
        container.append(inp);

        var input = new Input(parent, [], value);
        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                var num = + $(inp).val();
                if (!isInt(num)) {
                    throw new Error("integer value required");
                }
                if (num < min) {
                    throw new Error("value cannot be less than " + min);
                }
                if (num > max) {
                    throw new Error("value cannot be greater than " + max);
                }
                input.setValue(num);
            });
        });
        return input;
    }

    function isInt(num) {
        // In JS, % operator returns fractional remainder when first argument
        // is not a whole number. So this detects _all_ integers (even those
        // way outside the range of 32 or even 64 bit int). We do a range
        // check elsewhere (e.g. to verify if it's in range of 32-bits).
        return (num % 1) === 0;
    }

    function addStringIntToForm(container, parent, value, fld, min, max) {
        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        } else {
            if (typeof value !== 'string') {
                if (typeof value === 'number') {
                    value = value.toString();
                } else {
                    throw typeError(fld.type, value, "string");
                }
            }
            // parse string/make sure it's a number
            if (!isStringInt(value)) {
                throw new Error("value for type " + fld.type + " is not a valid number: " + JSON.stringify(value));
            }
            if (compareStringInts(value, min) < 0) {
                throw new Error("value for type " + fld.type + " is less than allowed min: " + value + " < " + min);
            }
            if (compareStringInts(value, max) > 0) {
                throw new Error("value for type " + fld.type + " is greater than allowed max: " + value + " > " + max);
            }
        }

        var inp = $('<input>');
        inp.attr('type', 'text');
        inp.attr('size', 28);
        inp.attr('value', "" + value);
        if (disabled) {
            inp.prop('disabled', true);
        }
        container.append(inp);

        var input = new Input(parent, [], value);
        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                var val = $(inp).val();
                if (!isStringInt(val)) {
                    throw new Error("integer value required");
                }
                if (compareStringInts(val, min) < 0) {
                    throw new Error("value cannot be less than " + min);
                }
                if (compareStringInts(val, max) > 0) {
                    throw new Error("value cannot be greater than " + max);
                }
                input.setValue(val);
            });
        });
        return input;
    }

    function isStringInt(str) {
        if (str === "") {
            return false;
        }
        // can start with minus sign
        if (str.charAt(0) === "-") {
            str = str.slice(1);
        }
        for (var i = 0; i < str.length; i++) {
            var ch = str.charAt(i);
            if (ch < '0' || ch > '9') {
                return false;
            }
        }
        return true;
    }

    function compareStringInts(a, b) {
        var aNeg = false;
        var bNeg = false;
        if (a.charAt(0) === '-') {
            aNeg = true;
            a = a.slice(1);
        }
        if (b.charAt(0) === '-') {
            bNeg = true;
            b = b.slice(1);
        }
        if (aNeg !== bNeg) {
            if (aNeg) {
                return -1;
            } else {
                return 1;
            }
        }
        var sgn = aNeg ? -1 : 1;
        var padLen = a.length - b.length;
        if (padLen > 0) {
            // a longer than b: pad b
            b = "0".repeat(padLen) + b;
        } else if (padLen < 0) {
            // a shorter than b: pad a
            a = "0".repeat(-padLen) + a;
        }
        // now we can safely use lexical compare
        if (a < b) {
            return -sgn;
        } else if (a > b) {
            return sgn;
        }
        return 0;
    }

    function addDoubleToForm(container, parent, value, fld) {
        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        } else if (typeof value !== 'number') {
            switch (value) {
                case "Infinity":
                    value = Infinity;
                    break;
                case "-Infinity":
                    value = -Infinity;
                    break;
                case "NaN":
                    value = NaN;
                    break;
                default:
                    throw typeError(fld.type, value, "number");
            }
        }

        var inp = $('<input>');
        inp.attr('type', 'text');
        inp.attr('size', 28);
        inp.attr('value', "" + value);
        if (disabled) {
            inp.prop('disabled', true);
        }
        container.append(inp);

        var input = new Input(parent, [], value);
        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                var val = $(inp).val();
                var num = + val;
                if (isNaN(num)) {
                    if (typeof val === "string") {
                        val = val.trim().toLowerCase();
                    }
                    // accept a few alternate spellings and don't care about case...
                    switch (val) {
                        case "inf": case "+inf": case "infinity": case "+infinity": case "infinite": case "+infinite":
                        num = "Infinity";
                        break;
                        case "-inf": case "-infinity": case "-infinite":
                        num = "-Infinity";
                        break;
                        case "nan":
                            num = "NaN";
                            break;
                        default:
                            throw new Error("numeric value required");
                    }
                }
                input.setValue(num);
            });
        });
        return input;
    }

    function addBoolToForm(container, parent, value, fld) {
        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        }
        if (typeof value !== 'boolean') {
            throw typeError(fld.type, value, "boolean");
        }

        var input = new Input(parent, [], value);

        var name = 'grpc_form_' + radioSeq;
        radioSeq++;
        var div = $('<div>');
        div.addClass('bool-input');
        for (var i = 0; i < 2; i++) {
            var checked = (i === 0) === value;
            var v = i === 0 ? 'true' : 'false';
            var inp = $('<input>');
            var id = name + '_' + v;
            inp.attr('type', 'radio');
            inp.attr('name', name);
            inp.attr('value', v);
            inp.attr('id', id);
            if (checked) {
                inp.prop('checked', true);
            }
            if (disabled) {
                inp.prop('disabled', true);
            }
            var lbl = $('<label>');
            lbl.attr('for', id);
            lbl.text(v);
            if (disabled) {
                lbl.addClass('disabled');
            }
            div.append(inp);
            div.append(lbl);

            (function(inp, b) {
                inp.click(function() {
                    input.setValue(b);
                });
            })(inp, i === 0);
        }
        container.append(div);

        return input;
    }

    function addStringToForm(container, parent, value, fld) {
        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        }
        if (typeof value !== 'string') {
            throw typeError(fld.type, value, "string");
        }

        var inp = $('<textarea>');
        inp.attr('cols', 40);
        inp.attr('rows', 1);
        inp.text(value);
        if (disabled) {
            inp.prop('disabled', true);
        }
        container.append(inp);

        var input = new Input(parent, [], value);
        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                var str = $(inp).val();
                input.setValue(str);
            });
        });
        return input;
    }

    function addBytesToForm(container, parent, value, fld) {
        var disabled = false;
        if (isUnset(value)) {
            value = fld.defaultVal;
            disabled = true;
        }
        if (typeof value !== 'string') {
            throw typeError(fld.type, value, "string");
        }
        if (!isBase64(value)) {
            throw new Error("value for type " + fld.type + " is not a valid base64-encoded string: " + JSON.stringify(value));
        }

        var box = $('<div>');
        box.addClass('grpc-bytes-container')
        var inp = $('<textarea>');
        inp.attr('cols', 40);
        inp.attr('rows', 1);
        inp.text(value);
        if (disabled) {
            inp.prop('disabled', true);
        }
        var lbl = $('<label>')
        lbl.addClass('grpc-file-button');
        lbl.text('Choose File');
        if (disabled) {
            lbl.addClass('disabled');
        }
        var fileInput = $('<input>');
        fileInput.attr('type', 'file');
        fileInput.attr('style', 'display:none');
        if (disabled) {
            fileInput.prop('disabled', true);
        }
        lbl.append(fileInput);
        box.append(inp);
        box.append(lbl);
        container.append(box);

        fileInput.on('change', function() {
            var reader = new FileReader();
            reader.addEventListener("load", function () {
                inp.text(btoa(this.result));
            }, false);
            reader.readAsBinaryString(fileInput[0].files[0]);
        })

        var input = new Input(parent, [], value);
        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                var str = $(inp).val();
                if (!isBase64(str)) {
                    throw new Error("value for type " + fld.type + " is not a valid base64-encoded string: " + JSON.stringify(value));
                }
                input.setValue(str);
            });
        });
        return input;
    }

    var base64alphabet_common = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789=";
    var base64alphabet_standard = "+/";
    var base64alphabet_websafe = "-_";

    function isBase64(str) {
        // TODO: verify padding chars only at end and right number of them
        var isStandard = false;
        var isWebsafe = false;
        for (var i = 0; i < str.length; i++) {
            var ch = str.charAt(i);
            if (base64alphabet_common.indexOf(ch) !== -1) {
                continue;
            }
            // not a common char, check dialect-specific chars
            if (base64alphabet_standard.indexOf(ch) !== -1) {
                if (isWebsafe) {
                    // string is mixed mode? reject
                    return false;
                }
                isStandard = true;
            } else if (base64alphabet_websafe.indexOf(ch) !== -1) {
                if (isStandard) {
                    // string is mixed mode? reject
                    return false;
                }
                isWebsafe = true;
            } else {
                // bad input!
                return false;
            }
        }
        return true;
    }

    var dateFormat = /^\d+-\d+-\d+T\d+:\d\d:\d\d(\.\d+)?Z(-?\d+:?\d\d)?$/;
    var timeFormat = /^\d+:\d\d(:\d\d(\.\d+)?)?(\s*(am|pm))?\s*((Z|GMT)?\s*(-|\+)?(\d+:?\d\d)?)?$/;

    function addTimestampToForm(container, parent, value, fld) {
        if (typeof value !== 'string') {
            throw typeError(fld.type, value, "string");
        }
        if (!dateFormat.test(value)) {
            throw new Error("value for type " + fld.type + " is not a valid timestamp string: " + JSON.stringify(value));
        }
        if (!isFinite(new Date(value).getDate())) {
            // value was invalid date string!
            throw new Error("value for type " + fld.type + " is not a valid timestamp string: " + JSON.stringify(value));
        }

        var parts = value.split('T', 2);

        var div = $('<div>');
        div.addClass('grpc-timestamp');
        var date = $('<input>');
        date.attr('type', 'text');
        date.attr('size', 12);
        date.attr('value', "" + parts[0]);
        div.append(date);

        var time = $('<input>');
        time.attr('type', 'text');
        time.attr('size', 20);
        time.attr('value', "" + parts[1]);
        div.append(time);
        container.append(div);

        var input = new Input(parent, [], value);
        var onTimeChange = function() {
            var timeStr = time.val();
            if (!timeFormat.test(timeStr)) {
                throw new Error('timestamp is not valid');
            }
            // make sure there is whitespace between time and any 'am' or 'pm' indicator
            timeStr = timeStr.replace('am', ' am').replace('pm', ' pm');

            // now test with known good day, for final verification that time is valid
            if (!isFinite(new Date('2018-10-31T' + timeStr).getDate())) {
                throw new Error('timestamp is not valid');
            }

            // now try to update value using date input
            var dateTime = new Date(date.val() + "T" + timeStr);
            if (isFinite(dateTime.getDate())) {
                input.setValue(dateTime.toISOString());
            }
        };
        var onDateChange = function() {
            // TODO(jh): Make this not suck. If a user types in garbage, this
            // will sadly prevent them from using the date picker pop-up until
            // they fix it. I can't find a reliable way to always validate the
            // input after it loses focus or the datepicker pop-up is dismissed
            // without this sort of spurious validation side effect :(

            // this will throw if date picker's text is bad
            $.datepicker.parseDate('yy-mm-dd', date.val());

            // now try to update value using time input
            var dateTime = new Date(date.val() + "T" + time.val());
            if (isFinite(dateTime.getDate())) {
                input.setValue(dateTime.toISOString());
            }
        };
        time.focus(function() {
            setValidation(this, onTimeChange);
        });
        date.focus(function() {
            setValidation(this, onDateChange);
        });
        var validateDate = function() {
            doValidation(date[0], onDateChange);
        };
        date.datepicker({
                            dateFormat: "yy-mm-dd",
                            showButtonPanel: true,
                            onClose: validateDate,
                            onSelect: validateDate,
                            nextText: '\u02C3',
                            prevText: '\u02C2',
                            beforeShow: function() {
                                $('#ui-datepicker-div').addClass('grpc-timestamp-picker');
                            },
                        });
        return input;
    }

    function addDurationToForm(container, parent, value, fld) {
        if (typeof value !== 'string') {
            throw typeError(fld.type, value, "string");
        }
        if (value.slice(-1) !== "s") {
            // duration must end in "s" for seconds
            throw new Error("value for type " + fld.type + " is not a valid duration string; " + JSON.stringify(value));
        }
        var secs = +(value.slice(0,-1));
        if (!isFinite(secs)) {
            throw new Error("value for type " + fld.type + " is not a valid duration string: " + JSON.stringify(value));
        }

        var div = $('<div>');
        var inp = $('<input>');
        inp.attr('type', 'text');
        inp.attr('size', 14);
        inp.attr('value', "" + secs);
        var unit = $('<select>');
        unit.append('<option value="0.000000001">nanoseconds</option>');
        unit.append('<option value="0.000001">microseconds</option>');
        unit.append('<option value="0.001">milliseconds</option>');
        unit.append('<option value="1" selected>seconds</option>');
        unit.append('<option value="60">minutes</option>');
        unit.append('<option value="3600">hours</option>');
        unit.append('<option value="86400">days</option>');
        unit.append('<option value="604800">weeks</option>');
        div.append(inp);
        div.append(unit);
        container.append(div);

        var input = new Input(parent, [], value);
        var onChange = function() {
            var num = + inp.val();
            if (isNaN(num)) {
                throw new Error("numeric value required");
            }
            num = num * unit.val();
            input.setValue(num.toFixed(9) + 's');
        };
        inp.focus(function() {
            setValidation(this, onChange);
        });
        unit.change(function() {
            try {
                onChange();
            } catch (ex) {
                // errors would be caused by the textbox, not
                // the unit, so ignore
            }
        });
        return input;
    }

    function addJSONToForm(container, parent, value, fld, mustBeObject) {
        if (mustBeObject && (typeof value !== 'object' || value instanceof Array)) {
            throw typeError(fld.type, value, "object");
        }

        var div = $('<div class="json-entry">');
        div.append('<h4>JSON</h4>');

        var inp = $('<textarea>');
        inp.attr('cols', 40);
        inp.attr('rows', 5);
        inp.val(JSON.stringify(value));
        div.append(inp);

        container.append(div);

        var input = new Input(parent, [], value);
        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                var o = JSON.parse($(inp).val());
                if (mustBeObject && (typeof o !== 'object' || o instanceof Array)) {
                    throw typeError(fld.type, o, "object");
                }
                input.setValue(o);
            });
        });
        return input;
    }

    function addAnyToForm(schema, container, parent, pathLen, value, allowMissing, fld) {
        if (typeof value !== 'object' || value instanceof Array) {
            throw typeError(fld.type, value, "object");
        }
        var typeUrl = value["@type"];
        if (isUnset(typeUrl) && allowMissing) {
            typeUrl = "type.googleapis.com/google.protobuf.StringValue";
            value["@type"] = typeUrl;
        }
        if (typeof typeUrl !== 'string') {
            throw new Error("value for type " + fld.type + " must have attribute '@type'");
        }

        var typeName = typeUrl.split("/").slice(-1)[0];
        if (typeName === 'google.protobuf.Any') {
            throw new Error('Any message cannot directly contain an Any message');
        }

        var table = makeTableContainer(container, pathLen);
        var children = {};
        var input = new Input(parent, children, value);

        var row = $('<tr>');
        row.addClass('message_field');
        table.append(row);
        var cell = makeFieldLabelCell({name:'@type', type:''}, schema);
        row.append(cell);

        cell = $('<td>');
        var inp = $('<input>');
        inp.attr('type', 'text');
        inp.attr('size', 42);
        inp.attr('list', 'grpc-message-types');
        inp.val(typeName);
        cell.append(inp);
        row.append(cell);

        var updateVal = function(typeName, allowMissing) {
            var wellKnown = isWellKnown(typeName);
            var knownType = wellKnown || schema.messageTypes.hasOwnProperty(typeName);

            var row;
            var cell;
            if (!knownType) {
                row = $('<tr class="unknown">');
                table.append(row);
                cell = $('<td>');
                row.append(cell);
                cell = $('<td>');
                row.append(cell);
                cell.append('<h4 class="unknown">unrecognized</h4>');
            }

            row = $('<tr>');
            row.addClass('message_field');
            table.append(row);
            cell = makeFieldLabelCell({name:'value', type:''}, schema);
            row.append(cell);
            cell = $('<td>');
            row.append(cell);

            if (wellKnown) {
                input.anyValue = value["value"];
                if (isUnset(input.anyValue)) {
                    input.anyValue = getInitialMessageValue(typeName);
                    value["value"] = input.anyValue;
                }
            } else {
                input.anyValue = $.extend({}, value);
                delete input.anyValue["@type"];
            }

            var valFld = {
                name: "value",
                type: typeName,
                oneOfFields: [],
                isMessage: true,
                isEnum: false,
                isArray: false,
                isMap: false,
                isRequired: false
            };
            if (knownType) {
                input.child = addElementToForm(schema, cell, input, pathLen, input.anyValue, allowMissing, valFld);
                if (!wellKnown && allowMissing) {
                    // if we added any required fields to anyValue, make sure we also
                    // add them to given value
                    $.extend(value, input.anyValue);
                }
            } else {
                input.child = addJSONToForm(cell, input, input.anyValue, valFld, true);
            }
        };

        inp.focus(function() {
            var inp = this;
            setValidation(inp, function() {
                // remove all but first row
                var rows = $('tr', table);
                for (var i = 1; i < rows.length; i++) {
                    $(rows[i]).remove();
                }

                // re-create value for new type
                for (var p in value) {
                    if (value.hasOwnProperty(p)) {
                        delete value[p];
                    }
                }
                var typeName = $(inp).val();
                value['@type'] = 'type.googleapis.com/' + typeName;
                updateVal(typeName, true);

                parent.onChange(input, [], value);
            });
        });

        input.onChange = function(child, revPath, value) {
            var typeName = $(inp).val();
            if (isWellKnown(typeName)) {
                revPath.push('value');
            } else if (revPath.length === 0) {
                // replace this value completely? don't forget @type
                value = $.extend({}, value);
                value['@type'] = 'type.googleapis.com/' + typeName;
            }
            this.parent.onChange(this, revPath, value);
        };
        input.onAdd = function(child, revPath, value) {
            var typeName = $(inp).val();
            if (isWellKnown(typeName)) {
                revPath.push('value');
            }
            this.parent.onAdd(this, revPath, value);
        };
        input.onDelete = function(child, revPath) {
            var typeName = $(inp).val();
            if (isWellKnown(typeName)) {
                revPath.push('value');
            }
            this.parent.onDelete(this, revPath);
        };

        // add value field
        updateVal(typeName, allowMissing);

        return input;
    }

    function addOptionalMessage(schema, container, parent, pathLen, value, allowMissing, fld) {
        var div = $('<div>');
        container.append(div);

        if (isUnset(value)) {
            div.append('<div class="null">unset</div>');
            return new Input(parent, [], undefined);
        }

        return addElementToForm(schema, container, parent, pathLen, value, allowMissing, fld);
    }

    function typeName(v) {
        if (v === null) {
            return "null";
        }
        var t = typeof v;
        if (t === 'object') {
            return v.constructor.name;
        }
        return t;
    }

    function typeError(type, val, expectedType) {
        var article = "a";
        if ("aeiou".indexOf(expectedType.charAt(0)) !== -1) {
            article = "an";
        }
        return new Error("value for type " + type + " should be " + article + " " + expectedType + "; instead got " + typeName(val))
    }

    function Input(parent, children, value) {
        this.parent = parent;
        this.children = children;
        this.value = value;

        this.setValue = function(value) {
            // TODO(jh): filter out no-op changes?
            // We can't do that as is because of how maps work: when we add a
            // row, we don't immediately check it (because it would be obnoxious
            // to get an error pop-up every time you add an entry to a map that
            // already has an entry with a default-value key). So we rely on a
            // no-op change bubbling up to actually add an entry to the request
            // value that has an empty/default key.
            this.value = value;
            this.parent.onChange(this, [], value);
        };

        this.onAdd = function(child, revPath, value) {
            pushChildPath(child, this.children, revPath);
            this.parent.onAdd(this, revPath, value);
        };

        this.onChange = function(child, revPath, value) {
            pushChildPath(child, this.children, revPath);
            this.parent.onChange(this, revPath, value);
        };

        this.onDelete = function(child, revPath) {
            pushChildPath(child, this.children, revPath);
            this.parent.onDelete(this, revPath);
        }
    }

    function pushChildPath(child, children, path) {
        for (var p in children) {
            if (children.hasOwnProperty(p)) {
                var ch = children[p];
                if (ch === child) {
                    path.push(p);
                    return;
                }
            }
        }
    }

    function newRootInput() {
        var root = new Input();
        root.onAdd = function(child, revPath, val) {
            onInputChange(revPath.reverse(), val);
        };
        root.onChange = function(child, revPath, val) {
            onInputChange(revPath.reverse(), val);
        };
        root.onDelete = function(child, revPath) {
            onInputDelete(revPath.reverse());
        };
        return root;
    }

    function getInitialValue(schema, fld) {
        if (fld.isMap) {
            return {}
        }
        if (fld.isArray) {
            return []
        }
        if (fld.isMessage) {
            return getInitialMessageValue(fld.type)
        }
        if (!isUnset(fld.defaultVal)) {
            return fld.defaultVal;
        }

        if (fld.isEnum) {
            var enumVals = schema.enumTypes[fld.type];
            return enumVals[0].name;
        }

        switch (fld.type) {
            case "int32":
            case "uint32":
            case "sint32":
            case "fixed32":
            case "sfixed32":
            case "double":
            case "float":
                return 0;
            case "int64":
            case "uint64":
            case "sint64":
            case "fixed64":
            case "sfixed64":
                return "0";
            case "string":
            case "bytes":
                return "";
            case "bool":
                return false;
            default:
                throw new Error('unknown non-message non-enum type: ' + fld.type);
        }
    }

    function getInitialMessageValue(messageType) {
        switch (messageType) {
            case "google.protobuf.Timestamp":
                return "1970-01-01T00:00:00Z";
            case "google.protobuf.Duration":
                return "0s";
            case "google.protobuf.Int32Value":
            case "google.protobuf.UInt32Value":
            case "google.protobuf.DoubleValue":
            case "google.protobuf.FloatValue":
                return 0;
            case "google.protobuf.Int64Value":
            case "google.protobuf.UInt64Value":
                return "0";
            case "google.protobuf.StringValue":
            case "google.protobuf.BytesValue":
                return "";
            case "google.protobuf.BoolValue":
                return false;
            case "google.protobuf.Value":
                return {};
            case "google.protobuf.ListValue":
                return [];
            default:
                return {};
        }
    }

    function isWellKnown(t) {
        switch (t) {
            case "google.protobuf.Int32Value":
            case "google.protobuf.UInt32Value":
            case "google.protobuf.Int64Value":
            case "google.protobuf.UInt64Value":
            case "google.protobuf.DoubleValue":
            case "google.protobuf.FloatValue":
            case "google.protobuf.StringValue":
            case "google.protobuf.BytesValue":
            case "google.protobuf.BoolValue":
            case "google.protobuf.Timestamp":
            case "google.protobuf.Duration":
            case "google.protobuf.Value":
            case "google.protobuf.Struct":
            case "google.protobuf.ListValue":
            case "google.protobuf.Any":
                return true;
            default:
                return false;
        }
    }

    /*
     * The following functions keep the request object and raw JSON form of the
     * request in sync with the form. When an input in the form changes, the
     * notifications will eventually bubble up to call the onInput* functions
     * below, which apply the mutation to the request object and then re-create
     * the raw JSON from the newly changed request object.
     */

    function onInputChange(path, value) {
        if (debug) {
            console.log("changing " + JSON.stringify(path) + " to " + JSON.stringify(value));
        }
        var req;
        if (path.length === 0) {
            // replacing value entirely
            requestForm.data("request", value);
            req = value;
        } else {
            req = requestForm.data("request");
            updateRequestObject(path, req, value);
        }
        updateJSONRequest(req);
    }

    function onInputDelete(path) {
        if (debug) {
            console.log("deleting " + JSON.stringify(path));
        }
        var req = requestForm.data("request");
        removeFromRequestObject(path, req);
        updateJSONRequest(req);
    }

    function updateRequestObject(path, req, value) {
        if (path.length === 1) {
            req[path[0]] = value;
        } else {
            updateRequestObject(path.slice(1), req[path[0]], value);
        }
    }

    function removeFromRequestObject(path, req) {
        if (path.length === 1) {
            if (req instanceof Array) {
                req.splice(path[0], 1);
            } else {
                delete req[path[0]];
            }
        } else {
            removeFromRequestObject(path.slice(1), req[path[0]]);
        }
    }

    var jsonRawTextArea = $("#grpc-request-raw-text");

    function updateJSONRequest(req) {
        jsonRawTextArea.val(JSON.stringify(req, null, 2));
    }

    function validateJSON() {
        var reqObj = JSON.parse($("#grpc-request-raw-text").val());
        rebuildRequestForm(reqObj, false);
    }

    jsonRawTextArea.focus(function() {
        setValidation(this, validateJSON);
    });

    var MAX_INT64 = "9223372036854775807";
    var MIN_INT64 = "-9223372036854775808";
    var MAX_UINT64 = "18446744073709551615";
    var MAX_INT32 = 2147483647;
    var MIN_INT32 = -2147483648;
    var MAX_UINT32 = 4294967295;

    // Adds a row to the request metadata table.
    function addMetadataRow(name = '', value = '') {
        tr = $('<tr class="metadataRow">');
        $("#grpc-request-metadata-form tr:last-of-type").before(tr);

        tr.append('<td><button class="delete">X</button></td>');
        $("button.delete", tr).click(function() {
            $(this).closest('tr').remove();
        });

        const td = $('<td>');
        const nameInput = $('<input class="name" size="24" />');
        const valueInput = $('<input class="value" size="40" />');
        const nameTd = $('<td />').append(nameInput);
        const valueTd = $('<td />').append(valueInput);

        if (name) {
            nameInput.val(name);
        }

        if (value) {
            valueInput.val(value);
        }

        tr.append(nameTd, valueTd);
    }

    // Invokes an RPC by sending the user-defined request data and metadata to
    // the server and then rendering the result to the "Response" tab.
    //
    // Some errors will result in an alert. But RPC errors returned by the
    // server will be rendered into the response tab. The response tab shows
    // all response headers and trailers, all response messages (usually zero or
    // one, but can be more than one for streaming calls), and any error details
    // if the final status of the RPC was not "Ok". For client-streaming calls,
    // it will also show if the server refused to accept all request messages
    // (This happens if the server terminates the call before the client has
    // finished uploading. It can even happen on a "successful" call, where the
    // server terminates the RPC but with no error, which still preempts sending
    // of any more request messages.)
    function invoke() {
        var service = $("#grpc-service").val();
        var method = $("#grpc-method").val();

        var timeoutStr = $("#grpc-request-timeout input").val();
        var timeout = Number(timeoutStr);
        timeout = (timeoutStr === "" || Number.isNaN(timeout)) ? undefined : timeout;

        const originalData = requestForm.data("request");
        let data = originalData;
        if (!(originalData instanceof Array)) {
            data = [data];
        }
        var metadata = [];
        var rows = $("#grpc-request-metadata-form tr");
        for (var i = 0; i < rows.length; i++) {
            var cells = $("input", rows[i]);
            if (cells.length === 0) {
                continue;
            }
            var name = $(cells[0]).val();
            var val = $(cells[1]).val();
            if (name !== "") {
                metadata.push({name: name, value: val });
            }
        }

        // ignore subsequent clicks until this RPC finishes
        $(".grpc-invoke").prop("disabled", true);

        if (originalData instanceof Array) {
            cloneData = $.extend(true, [], originalData)
        } else {
            cloneData = $.extend(true, {}, originalData)
        }
        const historyItem = {
            request: {
                timeout_seconds: timeout,
                metadata: $.extend([], metadata),
                data: cloneData
            },
            service: service,
            method: method,
            startTime: new Date().toISOString(),
        }

        const startTime = window.performance.now();

        $.ajax(
            {
                type: "POST",
                url: invokeURI + "/" + service + "." + method,
                contentType: "application/json",
                data: JSON.stringify({timeout_seconds: timeout, metadata: metadata, data: data}),
            })
            .done(function(responseData) {
                var durationMs = window.performance.now() - startTime;
                renderResponse(historyItem, durationMs, responseData);
            })
            .fail(function(failureData, status) {
                addHistory({
                    ...historyItem,
                    durationMS: window.performance.now() - startTime,
                    failureStatus: status,
                });
                alert("Unexpected error: " + status);
                if (debug) {
                    console.trace(failureData.responseText);
                }
            })
            .always(function() {
                $(".grpc-invoke").prop("disabled", false);
            });
    }

    function renderResponse(historyItem, durationMs, responseData) {
        if (responseData.headers instanceof Array && responseData.headers.length > 0) {
            var hdrs = $("#grpc-response-headers");
            hdrs.empty();
            for (var i = 0; i < responseData.headers.length; i++) {
                var hdrRow = $('<tr>');
                hdrs.append(hdrRow);
                var hdrCell = $('<td>');
                hdrCell.text(responseData.headers[i].name);
                hdrRow.append(hdrCell);
                hdrCell = $('<td>');
                hdrCell.text(responseData.headers[i].value);
                hdrRow.append(hdrCell);
            }
        } else {
            $("#grpc-response-headers").html('<tr><td class="none">None</td></tr>');
        }

        if (responseData.requests && responseData.requests.total !== responseData.requests.sent) {
            var stats = $("#grpc-response-req-stats");
            stats.show();
            stats.text('Only ' + responseData.requests.sent + ' of ' + responseData.requests.total + ' requests accepted');
        } else {
            $("#grpc-response-req-stats").hide();
        }

        // TODO(jh): better presentation of responses?
        // It would be really nice to show type information. It would also be nice
        // to render maps a little differently and also to omit unset one-of fields
        // (or even better, be able to render unset fields differently).
        // But we need response schema info to do that...
        renderMessages($("#grpc-response-data"), responseData.responses)

        if (responseData.error) {
            $("#grpc-response-error").show();
            $("#grpc-response-error-desc").text(responseData.error.name);
            $("#grpc-response-error-num").text('(' + responseData.error.code + ')');
            if (responseData.error.message !== responseData.error.name) {
                var msg = $("#grpc-response-error-msg");
                msg.show();
                msg.text(responseData.error.message);
            } else {
                $("#grpc-response-error-msg").hide();
            }
            if (renderMessages($("#grpc-response-error-details"), responseData.error.details)) {
                $("#grpc-response-error-details-container").show();
            } else {
                $("#grpc-response-error-details-container").hide();
            }
        } else {
            $("#grpc-response-error").hide();
        }

        addHistory({
            ...historyItem,
            durationMS: durationMs,
            responseData: historyResponseData(responseData),
        });

        // TODO(jh): "copy as grpcurl" button? This would provide a
        // command-line for grpcurl that does the same thing as clicking
        // the "invoke" button for the current request responseData and metadata.

        // TODO(jh): "paste as grpcurl" button? This would provide a
        // way to paste in a grpcurl command-line which would then select
        // the right method, and populate the request responseData and metadata.

        if (responseData.trailers instanceof Array && responseData.trailers.length > 0) {
            var tlrs = $("#grpc-response-trailers");
            tlrs.empty();
            for (i = 0; i < responseData.trailers.length; i++) {
                var tlrRow = $('<tr>');
                tlrs.append(tlrRow);
                var tlrCell = $('<td>');
                tlrCell.text(responseData.trailers[i].name);
                tlrRow.append(tlrCell);
                tlrCell = $('<td>');
                tlrCell.text(responseData.trailers[i].value);
                tlrRow.append(tlrCell);
            }
        } else {
            $("#grpc-response-trailers").html('<tr><td class="none">None</td></tr>');
        }

        var t = $("#grpc-request-response");
        t.tabs("enable", 2);
        t.tabs("option", "active", 2);
    }

    function renderMessages(enclosingDiv, messages) {
        enclosingDiv.empty();
        if (messages instanceof Array && messages.length > 0) {
            enclosingDiv.show();
            for (const msg of messages) {
                const container = $('<div>');
                if (msg.isError) {
                    container.html('<div class="error">Server error processing message #' + (i+1) + '</div>');
                } else {
                    const textArea = $('<textarea>');
                    textArea.val(JSON.stringify(msg.message, null, 2));
                    textArea.addClass('grpc-response-textarea');
                    container.append(textArea);
                }
                enclosingDiv.append(container);
            }
            return true;
        } else {
            enclosingDiv.hide();
            return false;
        }
    }

    function historyResponseData(responseData) {
        // extract the minimum info from responseData for showing summary of
        // operation in history UI
        let result = {};
        if (responseData.error) {
            result.error = {
                name: responseData.error.name,
                code: responseData.error.code,
            };
        }
        result.responseMsgCount = responseData.responses?.length ?? 0;
        return result;
    }

    // Renders the given response value to the given DIV element. The depth
    // parameter is used to add CSS styles to container elements (in the same
    // way that pathLen is used in the add*ToForm functions above).
    // TODO: Wire this back into results rendering if pretty results are still desirable.
    function populateResultContainer(div, depth, val) {
        if (val === null) {
            div.html('<span class="null">null</span>');
        } else if (typeof val === 'object') {
            var container = makeResultContainer(div, depth);
            depth++;
            var table = $('<table>');
            container.append(table);
            if (val instanceof Array) {
                for (var i = 0; i < val.length; i++) {
                    var element = $('<tr>');
                    table.append(element);
                    var td = $('<td>');
                    element.append(td);
                    populateResultContainer(td, depth, val[i]);
                }
                if (val.length === 0) {
                    table.append('<tr><td><span class="null">No elements</span></td></tr>');
                }
            } else {
                var count = 0;
                for (var p in val) {
                    if (val.hasOwnProperty(p)) {
                        count++;
                        var row = $('<tr>');
                        table.append(row);
                        var cell = $('<td>');
                        cell.addClass("name");
                        row.append(cell);
                        cell.text(p);
                        cell = $('<td>');
                        row.append(cell);
                        populateResultContainer(cell, depth, val[p]);
                    }
                }
                if (count === 0) {
                    table.append('<tr><td><span class="null">No fields</span></td></tr>');
                }
            }
        } else {
            div.text('' + val);
        }
    }

    function makeResultContainer(div, depth) {
        if (depth === 0) {
            return div;
        }
        var container = $('<div>');
        // add classes
        container.addClass('output_container');
        for (var i = 2; i <= 5; i++) {
            var idx = (depth % i) + 1;
            container.addClass(numberName(idx) + "-of-" + i);
        }
        div.append(container);
        return container;
    }

    var validation;

    // Returns true if the currently configured validation function is for
    // the given element
    function isValidating(element) {
        return !isUnset(validation) && validation.element === element;
    }

    // Sets the current validation to the given element function. The given
    // function should throw an error with a useful error message if the
    // element's value/state is not valid.
    //
    // If a different element is configured for validation, the given element
    // will be validated first. This also configures the given element so that
    // the function is checked when the element loses focus.
    //
    // When creating a new input, to enable validation, simply configure the
    // element's "on focus" event handler to call this function, supplying the
    // element and its validator function. If the element corresponds to an
    // input in the form, the function is responsible for propagating the change
    // by calling the parent input's onChange method.
    function setValidation(element, func) {
        if (isValidating(element)) {
            // no op
            return;
        }

        checkValidation();
        validation = { element: element, func: func };
        $(element).blur(onElementBlur);
    }

    function doValidation(element, func) {
        setValidation(element, func);
        $(element).blur();
    }

    // Checks the currently configured element using the configured validation
    // function. Returns true if the element is valid. When the function
    // returns, the element will have an "invalid-input" CSS class if and only
    // if it failed validation.
    function checkValidation() {
        if (!isUnset(validation)) {
            var err = '';
            try {
                validation.func(validation.element);
            } catch (ex) {
                if (debug) {
                    console.trace(ex);
                }
                if (ex.message) {
                    err = ex.message
                } else {
                    err = ex.toString();
                }
            }

            var el = $(validation.element);
            if (err) {
                alert(err);
                el.addClass('invalid-input');
                return false;
            } else {
                el.removeClass('invalid-input');
            }
        }
        return true;
    }

    // Allows the given event only if the form is in a currently valid state.
    // This runs any current validation if necessary. It also checks to see if
    // any elements are in the "invalid-input" state (by checking CSS class).
    // It returns true if all is good and the form is valid.
    function onlyIfValid(evt) {
        var isValid = checkValidation();
        if (isValid) {
            var badInputs = $(".invalid-input").filter(":visible");
            if (badInputs.length > 0) {
                var badInput = $(badInputs[0]);
                $([document.documentElement, document.body]).scrollTop(badInput.offset().top);
                badInput.focus();

                isValid = false;
            }
        }

        if (!isValid) {
            evt.preventDefault();
            evt.stopImmediatePropagation();
            return false;
        }
        return true;
    }

    function onElementBlur() {
        if (document.activeElement !== this && isValidating(this)) {
            checkValidation();
            validation = undefined;
        }
    }

    let history = [];
    let examples = [];
    // TODO: add ability to customize these max history values
    const maxHistory = 100;
    const maxHistorySize = 1024*1024; // 1mb
    const target = $(".target").text();

    const historyStorageKey = `grpcui-history-${window.location.host}-${target}`;
    const expandDescStorageKey = `grpcui-expand-description`;

    const loadExamples = () => {
        $.ajax({
            url: 'examples',
            type: 'GET',
            success: function(data) {
                examples = data;
                // only populate the example list if we have some
                if (examples && examples.length > 0) {
                    showExamplesUI();
                }
            },
            error: function(e) {
                console.log("Failed to load examples: " + e);
            }
        });
    }

    const showExamplesUI = () => {
        let examplesList = $("#grpc-request-examples");
        examples.forEach(example => {
            let exampleItem = $('<li class="grpc-request-example">');
            exampleItem.addClass("grpc-request-example");
            exampleItem.text(example.name);
            if (example.description) {
                exampleItem.prop('title', example.description);
                exampleItem.tooltip();
            }
            examplesList.append(exampleItem);
        })
        examplesList.selectable({
            stop: function() {
                $(".ui-selected", this).each(function() {
                    const index = $("li", examplesList).index(this);
                    loadRequest(examples[index])
                });
            }
        });
        $("#grpc-request-examples-container").show();
    }

    const loadHistory = () => {
        const json = localStorage.getItem(historyStorageKey);
        if (json) {
            history = (JSON.parse(json));
        }
        updateHistoryUI();
    }

    const onHistoryChange = () => {
        let data = '';
        while (true) {
            data = JSON.stringify(history);
            if (data.length <= maxHistorySize) {
                break;
            }
            // purge oldest item from history to see if that makes enough room
            history = history.slice(0, history.length - 1);
        }

        try {
            localStorage.setItem(historyStorageKey, data);
        } catch (e) {
            // Likely no room in local storage quota. This can still happen, despite
            // above code to limit the size, because there could be other keys in
            // storage OR we could be in incognito mode, which doesn't allow writing
            // to local storage...
            //
            // Log the error and keep going.
            console.trace(e);
        }
        updateHistoryUI();
    }

    const clearHistory = () => {
        if (confirm('Are you sure you wish to delete all history? This action is permanent and cannot be undone.')) {
            history = [];
            onHistoryChange();
        }
    }

    const download = (filename, text) => {
        let element = document.createElement('a');
        element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
        element.setAttribute('download', filename);

        element.style.display = 'none';
        document.body.appendChild(element);

        element.click();

        document.body.removeChild(element);
    }

    const saveHistory = () => {
        let savedHistory = [];
        for (let i = 0; i < history.length; i++) {
            let item = $.extend({}, history[i]); // make a copy before mutating
            item.name = "Example #" + (i+1) + " @ " + item.startTime;
            item.description = "";
            delete item.startTime;
            delete item.durationMS;
            delete item.responseData;
            let req = item.request;
            if (req.hasOwnProperty('timeout')) {
                req = $.extend({}, req); // make a copy before mutating
                item.request = req;
                // if old history item has 'timeout' property, convert it to 'timeout_seconds'
                let timeout = Number(req.timeout);
                timeout = (timeoutStr === "" || Number.isNaN(timeout)) ? undefined : timeout;
                req.timeout_seconds = timeout;
                delete req.timeout;
            }
            savedHistory.push(item);
        }

        download('history.json', JSON.stringify(savedHistory, null, 2));
    }

    const addHistory = (item) => {
        history = history.slice(0, maxHistory - 1);
        history.unshift(item);
        onHistoryChange();
    }

    const updateHistoryUI = () => {
        const list = $('#grpc-history-list');
        list.empty();
        const accordion = $('<div>');
        list.append(accordion);
        for (let i = 0; i < history.length; i++) {
            const item = history[i];
            const id = `grpc-history-item-${i}`;
            const dataString = JSON.stringify(item.request.data, null, 4);
            const valid = services[item.service] && services[item.service].includes(item.method);
            let result = '';
            let messages = '&nbsp;';
            let err = false;
            if (item.failureStatus) {
                result = `Failure: ${item.failureStatus}`;
                err = true;
            } else if (item.responseData?.error) {
                // for errors, show if any response messages were sent
                let numMsgs = item.responseData.responseMsgCount ?? 0;
                if (numMsgs === 1) {
                    messages = `1 response message`; // singular
                } else if (numMsgs > 1) {
                    messages = `${numMsgs} response messages`;
                }
                result = item.responseData.error.name ?? 'FAILED';
                err = true;
            } else {
                // on success, only show number of response messages if not one (e.g. a stream)
                messages = (item.responseData?.responseMsgCount ?? 1) !== 1 ? `${item.responseData.responseMsgCount } response messages` : '';
                result = 'OK';
            }
            accordion.append(`<div class="history-item-header" id="${id}">
                <span class="history-item-delete"><button class="delete" id="delete-${id}">X</button></span>
                <span class="history-item-load">
                    <button class="load" ${valid ? '' : 'disabled'} id="load-${id}">Load</button>
                </span>
                <span class="history-item-time">${new Date(item.startTime).toLocaleString()}</span>
                <span class="history-item-duration">${item.durationMS.toFixed(2)}ms</span>
                <span class="history-item-result ${err ? 'error' : ''}">${result}</span>
                <span class="history-item-method ${valid ? '' : 'invalid-history'}"
                      ${valid ? '' : 'title="Service or method no longer available"'}
                >
                    ${item.service}.${item.method}
                </span>
                <span class="history-item-messages">${messages}</span>
            </div>`);
            accordion.append(`<div class="history-item-panel">
                <div class="history-detail-request">
                    <div class="history-detail-heading">Request</div>
                    <span><pre class="request-json">${dataString.slice(0, 250)}${dataString.length > 250 ? '...' : ''}</pre></span>
                </div>
                ${item.request.metadata.length === 0 ? '' : `
                <div class="history-detail-metadata">
                    <div class="history-detail-heading">Metadata</div>
                    <table>
                        ${item.request.metadata.map((item) => `
                        <tr><th>${item.name}</th><td>${item.value}</td></tr>
                        `).join('\n')}
                    </table>
                </div>`}
            </div>`);
            $(`#delete-${id}`).click((evt) => {
                deleteHistoryItem(i);
                evt.preventDefault();
                evt.stopImmediatePropagation();
            });
            $(`#load-${id}`).click((evt) => {
                loadHistoryItem(i);
                // These prevent the accordion from opening or folding when clicking load...
                evt.preventDefault();
                evt.stopImmediatePropagation();
            });
        }
        accordion.accordion({
            animate: 200,
            active: false,
            collapsible: true,
            icons: false,
            header: ".history-item-header",
            heightStyle: "content",
        });
    }

    const loadRequest = (item) => {
        const t = $("#grpc-request-response");
        t.tabs("option", "active", 0);
        let timeout = "";
        if (item.request.timeout_seconds) {
            timeout = item.request.timeout_seconds + "";
        } else if (item.request.timeout) {
            // older versions stored string in 'timeout' attribute; so support
            // that in case someone loads an item from history from older
            // version of grpcui
            timeout = item.request.timeout;
        }
        $("#grpc-request-timeout input").val(timeout);
        $("#grpc-service").val(item.service);
        formServiceSelected(() => {
            $("#grpc-method").val(item.method);
            formMethodSelected(() => {
                jsonRawTextArea.val(JSON.stringify(item.request.data, null, 2));
                validateJSON();
                // remove all rows
                $("tr").remove('.metadataRow');
                for (let metadata of item.request.metadata) {
                    addMetadataRow(metadata.name, metadata.value);
                }
            });
        });
    }

    const clearExampleSelection = () => {
        $('#grpc-request-examples .ui-selected').removeClass('ui-selected')
    }

    const loadHistoryItem = (index) => {
        clearExampleSelection();
        loadRequest(history[index]);
    }

    const deleteHistoryItem = (index) => {
        history.splice(index, 1);
        onHistoryChange();
    }

    $("#grpc-request-timeout input").focus(function() {
        var inp = this;
        setValidation(inp, function() {
            var val = $(inp).val();
            if (val === "" || val === undefined) {
                return;
            }
            var num = Number(val);
            if (Number.isNaN(num)) {
                throw new Error("numeric value required");
            }
            if (num <= 0) {
                $(inp).val(undefined);
                throw new Error("timeout value must be greater than zero");
            }
        });
    });

    $("#grpc-request-response").tabs(
        {
            beforeActivate: function(e) {
                onlyIfValid(e);
            }
        });

    $("#grpc-service").change(() => formServiceSelected());
    $("#grpc-method").change(() => {
        clearExampleSelection();
        formMethodSelected();
    });

    $("#grpc-request-metadata-add-row").click(function() {
        addMetadataRow();
    });
    $(".grpc-invoke").click(function(e) {
        if (onlyIfValid(e)) {
            invoke();
        }
    });

    if (localStorage.getItem(expandDescStorageKey) === "true") {
        descriptionsShown = true;
        $("#grpc-descriptions-toggle").text("«")
    } else {
        $("#grpc-descriptions pre").hide();
    }
    $("#grpc-descriptions-toggle").click(() => {
        if (descriptionsShown) {
            $("#grpc-descriptions pre").hide();
            $("#grpc-descriptions-toggle").text("»")
        } else {
            $("#grpc-descriptions pre").show();
            $("#grpc-descriptions-toggle").text("«")
        }
        descriptionsShown = !descriptionsShown;
        localStorage.setItem(expandDescStorageKey, descriptionsShown+"");
    });

    $('#grpc-history-clear').click(() => clearHistory());
    $('#grpc-history-save').click(() => saveHistory());

    loadExamples();
    loadHistory();

    // TODO(jh): support populating the selected method and even request
    // data and metadata from URL hash fragment (and add a way for user to
    // get URL with hash fragment for currently selected method and data)

    // initialize methods drop-down based on selected service
    formServiceSelected();

    if (isUnset(headers) || headers.length === 0) {
        // add a single blank entry to request metadata table
        addMetadataRow();
    } else {
        for (let metadata of headers) {
            addMetadataRow(metadata.name, metadata.value);
        }
    }
};
