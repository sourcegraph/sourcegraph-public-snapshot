package tar

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveAndExtract(t *testing.T) {
	tempDirSource, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tempDirSource)

	tempDirDestination, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tempDirDestination)

	fileContents := map[string]string{
		"0":         "Aenean maximus dolor id mi condimentum fringilla.",
		"1":         "Aliquam interdum feugiat auctor.",
		"2":         "Aliquam molestie pulvinar tellus, eget auctor sapien mattis non.",
		"3":         "Aliquam venenatis tortor eros, id sodales turpis blandit id.",
		"4":         "Cras tempus quam odio, sit amet tincidunt tortor pellentesque sit amet.",
		"5":         "Donec commodo, dui quis fringilla mollis, est elit venenatis sapien, eget laoreet quam ante eu odio.",
		"6":         "Donec malesuada accumsan gravida.",
		"7":         "Donec tincidunt lectus metus, at lobortis nisi maximus in.",
		"8":         "Donec tristique enim non turpis dignissim placerat.",
		"9":         "Duis bibendum eu eros eu faucibus.",
		"foo/0":     "Etiam a dignissim urna, quis porttitor nulla.",
		"foo/1":     "Etiam in finibus ligula, ut dictum tellus.",
		"foo/2":     "Fusce semper metus vel quam tempus, quis sollicitudin turpis condimentum.",
		"foo/3":     "In a vestibulum augue.",
		"foo/4":     "In convallis dui ut urna auctor maximus.",
		"foo/5":     "In est neque, pulvinar eget velit quis, tristique facilisis nibh.",
		"foo/6":     "In pharetra eros vitae tempus faucibus.",
		"foo/7":     "In vel turpis lectus.",
		"foo/8":     "Integer condimentum vel metus ac accumsan.",
		"foo/9":     "Integer et mauris faucibus, tempor purus a, bibendum magna.",
		"bar/0":     "Integer varius ultrices rhoncus.",
		"bar/1":     "Integer vel egestas felis, ac porta augue.",
		"bar/2":     "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"bar/3":     "Maecenas porta, enim sit amet blandit luctus, mi justo vestibulum quam, eu rhoncus lacus dui eget felis.",
		"bar/4":     "Maecenas sed neque tristique, volutpat mauris eu, fermentum leo.",
		"bar/5":     "Mauris eu libero augue.",
		"bar/6":     "Mauris hendrerit, ante fermentum facilisis congue, quam odio blandit sem, placerat ultricies mauris nibh eget felis.",
		"bar/7":     "Morbi blandit, felis vitae gravida imperdiet, lacus ipsum varius quam, quis ullamcorper felis nisi nec elit.",
		"bar/8":     "Morbi commodo at urna non laoreet.",
		"bar/9":     "Morbi dapibus malesuada dolor in consequat.",
		"foo/bar/0": "Morbi semper semper ex quis semper.",
		"foo/bar/1": "Morbi tincidunt tellus turpis, eget finibus magna congue in.",
		"foo/bar/2": "Morbi ut nisi nec purus sollicitudin feugiat.",
		"foo/bar/3": "Morbi vehicula sodales ante, eu sodales purus dapibus ac.",
		"foo/bar/4": "Nulla feugiat elementum ligula a imperdiet.",
		"foo/bar/5": "Nulla sed augue augue.",
		"foo/bar/6": "Nulla tempor sapien eu posuere dignissim.",
		"foo/bar/7": "Nullam blandit nisl non enim accumsan, ac vulputate elit fringilla.",
		"foo/bar/8": "Nullam cursus eros a ipsum laoreet commodo.",
		"foo/bar/9": "Nunc imperdiet lacus quis cursus sodales.",
		"baz/foo/0": "Pellentesque non magna luctus, sodales erat non, sollicitudin sapien.",
		"baz/foo/1": "Proin facilisis nisi est, id ornare enim congue ut.",
		"baz/foo/2": "Proin laoreet, tellus sed rhoncus ultricies, nisi odio egestas est, ac porttitor odio augue sit amet ante.",
		"baz/foo/3": "Quisque a metus libero.",
		"baz/foo/4": "Sed in lectus et quam malesuada dapibus eu id diam.",
		"baz/foo/5": "Sed leo est, pretium quis dignissim ac, placerat ac justo.",
		"baz/foo/6": "Sed ut nunc in purus pharetra consequat.",
		"baz/foo/7": "Sed vitae lacus felis.",
		"baz/foo/8": "Suspendisse a urna turpis.",
		"baz/foo/9": "Suspendisse non orci vel ex bibendum feugiat a a nibh.",
	}
	for filename, contents := range fileContents {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(tempDirSource, filename)), os.ModePerm); err != nil {
			t.Fatalf("unexpected error creating driectory: %s", err)
		}

		if err := ioutil.WriteFile(filepath.Join(tempDirSource, filename), []byte(contents), os.ModePerm); err != nil {
			t.Fatalf("unexpected error writing file: %s", err)
		}
	}

	if err := Extract(tempDirDestination, Archive(tempDirSource)); err != nil {
		t.Fatalf("unexpected error archiving and extracting: %s", err)
	}

	for filename, expected := range fileContents {
		actual, err := ioutil.ReadFile(filepath.Join(tempDirDestination, filename))
		if err != nil {
			t.Fatalf("Unexpected error reading file: %s", err)
		}

		if string(actual) != expected {
			t.Errorf("unexpected content for file %s. want=%q have=%q", filename, expected, actual)
		}
	}
}
