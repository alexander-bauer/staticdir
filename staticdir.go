package staticdir

import (
	"io"
	"os"
	"path"
)

type Translator struct {
	Source, Target string

	// ExcludeDir and ExcludeFile are used for determining if a file
	// or directory should not be copied from the source to the target
	// directory.
	ExcludeDir  func(os.FileInfo) bool
	ExcludeFile func(os.FileInfo) bool

	// DirMode is the FileMode that copied directories are created
	// with.
	DirMode os.FileMode

	// CopyFunc is called when copying a source file to the target
	// directory, after it has already been checked with
	// ExcludeFile. It is passed the path to the source file, target
	// file, the source fileinfo, and CopyData, which can be anything.
	CopyFunc func(string, string, os.FileInfo, interface{}) error
	CopyData interface{}
}

func New(source, target string) *Translator {
	return &Translator{
		Source: path.Clean(source),
		Target: path.Clean(target),

		ExcludeDir:  ExcludeNone,
		ExcludeFile: ExcludeNone,

		DirMode:  0755,
		CopyFunc: ColdCopy,
	}
}

func (t *Translator) Translate() error {
	return nil
}

func (t *Translator) CopyDir(subpath string) error {
	children, err := GetChildren(path.Join(t.Source, subpath))
	if err != nil {
		return err
	}

	// Create the matching subdirectory.
	err = os.Mkdir(path.Join(t.Target, subpath), t.DirMode)
	if err != nil {
		return err
	}

	// Copy over every child in the source directory.
	for _, child := range children {
		// If the child is a directory, recursively call CopyDir on
		// it, giving the basename as the new part of the
		// subpath. Otherwise, call CopyFile.
		if child.IsDir() {
			t.CopyDir(path.Join(subpath, child.Name()))
		} else {
			t.CopyFile(path.Join(subpath, child.Name()), child)
		}
	}

	return nil
}

func (t *Translator) CopyFile(subpath string, fi os.FileInfo) error {
	if !t.ExcludeFile(fi) {
		return t.CopyFunc(path.Join(t.Source, subpath),
			path.Join(t.Target, subpath),
			fi, t.CopyData)
	}
	return nil
}

// GetChildren retrieves all fileinfos contained by a directory.
func GetChildren(path string) (fis []os.FileInfo, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}

	fis, err = f.Readdir(0)
	f.Close()
	return
}

func ExcludeNone(fi os.FileInfo) bool {
	return false
}

// ColdCopy simply copies a source file to a target file, discarding
// other parameters.
func ColdCopy(source, target string, fi os.FileInfo,
	data interface{}) error {

	// Begin by opening the in file and creating the out file.
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	// Then just copy it.
	_, err = io.Copy(out, in)
	return err
}
