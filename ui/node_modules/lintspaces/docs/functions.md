## Functions

An instance of the _Lintspaces validator_ has the following methods

### ```validate(path)```

This function runs the check for a given file based on the validator settings.

* **Parameter ```path```** is the path to the file as ```String```.
* **Throws** an error when given ```path``` is not a valid path.
* **Throws** an error when given ```path``` is not a file.
* **Returns** ```undefined```.

### ```getProcessedFiles()```

This returns the amount of processed through the validator.

* **Returns** the amount as ```Number```.

### ```getInvalidFiles()```

This returns all invalid lines and messages from processed files.

* **Returns** each key in the returned ```Object``` represents a path of a
processed invalid file. Each value is an other object containing the validation
result. For more informations about the returned format see [Usage](#usage).

### ```getInvalidLines(path)```

This returns all invalid lines and messages from the file of the given
```path```. This is just a shorter version of ```getInvalidFiles()[path]```.

* **Parameter ```path```** is the file path
* **Returns** each key in the returned ```Object``` represents a line from the
file of the given path, each value an exeption message of the current line. For
more informations about the returned format see [Usage](#usage).
