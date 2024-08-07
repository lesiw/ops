package main

import (
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"lesiw.io/cmdio"
	"lesiw.io/cmdio/sys"
	"lesiw.io/defers"
	"lesiw.io/flag"
)

var (
	flags    = flag.NewSet(os.Stderr, "op OPERATION")
	_        = flags.Bool("l", "list ops")
	install  = flags.Bool("install-completions", "install completion scripts")
	printver = flags.Bool("V,version", "print version")

	//go:embed version.txt
	versionfile string
	version     string
)

func init() {
	version = strings.TrimRight(versionfile, "\n")
}

func main() {
	cmdio.Trace = io.Discard
	if err := run(); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err)
		}
		defers.Exit(1)
	}
	defers.Exit(0)
}

func run() error {
	if err := flags.Parse(os.Args[1:]...); err != nil {
		return errors.New("")
	}
	if *printver {
		fmt.Println(version)
		return nil
	} else if *install {
		return installComp()
	}

	workdir, err := toRootDir()
	if err != nil {
		return err
	}
	if err := toOpsDir(workdir); err != nil {
		return err
	}

	bindir, err := cacheDir("bin")
	if err != nil {
		return fmt.Errorf("failed to create bincache: %w", err)
	}

	mtime, err := newestMtime(".")
	if err != nil {
		return fmt.Errorf("failed to determine src mtime: %w", err)
	}

	dirid, err := id()
	if err != nil {
		return err
	}

	opsbin := filepath.Join(bindir, "ops-"+base64.URLEncoding.
		WithPadding(base64.NoPadding).EncodeToString(dirid[:]))
	stat, err := os.Stat(opsbin)
	if os.IsNotExist(err) || stat.ModTime().Before(mtime) {
		if err := buildBin(opsbin); err != nil {
			return err
		}
	}

	if err := os.Chdir(workdir); err != nil {
		return fmt.Errorf("failed to chdir to '%s': %w", workdir, err)
	}

	err = sys.Run(append([]string{opsbin}, os.Args[1:]...)...)
	if err != nil {
		return errors.New("")
	}
	return nil
}

func id() (id uuid.UUID, err error) {
	var rawid []byte
	rawid, err = os.ReadFile(".uuid")
	var pe *fs.PathError
	if err == nil {
		uuidstring := strings.TrimSpace(string(rawid))
		if id, err = uuid.Parse(uuidstring); err != nil {
			err = fmt.Errorf("failed to parse .uuid: %s", err)
			return
		}
		return
	}
	if !errors.As(err, &pe) {
		err = fmt.Errorf("failed to read .uuid: %s", err)
		return
	}
	id = uuid.New()
	err = os.WriteFile(".uuid", []byte(id.String()+"\n"), 0644)
	if err != nil {
		err = fmt.Errorf("failed to write .uuid: %s", err)
		return
	}
	return
}

func cacheDir(path ...string) (cache string, err error) {
	if cache, err = os.UserCacheDir(); err != nil {
		return "", fmt.Errorf("failed to get user cache directory: %s", err)
	}
	cache = filepath.Join(cache, "ops", filepath.Join(path...))
	if err = os.MkdirAll(cache, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %s", err)
	}
	return
}

func toRootDir() (string, error) {
	for {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		for _, dir := range []string{".git", "ops"} {
			fileinfo, err := os.Stat(dir)
			if err == nil && fileinfo.IsDir() {
				return cwd, nil
			}
		}
		reachedRoot := (cwd == "/" || cwd == (filepath.VolumeName(cwd)+"\\"))
		if reachedRoot || os.Chdir("..") != nil {
			return "", fmt.Errorf("No .git or ops directory was found.")
		}
	}
}

func toOpsDir(root string) error {
	if stat, err := os.Stat("ops"); os.IsNotExist(err) || !stat.IsDir() {
		return fmt.Errorf("no 'ops' directory found: %w", err)
	}
	opsdir := filepath.Join(root, "ops")
	if err := os.Chdir(opsdir); err != nil {
		return fmt.Errorf("failed to chdir to '%s': %w", opsdir, err)
	}
	return nil
}

func newestMtime(dir string) (mtime time.Time, err error) {
	err = filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			t := info.ModTime()
			if mtime.IsZero() || t.After(mtime) {
				mtime = t
			}
			return nil
		},
	)
	return
}

func buildBin(path string) error {
	if err := sys.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("'go mod tidy' failed: %w", err)
	}
	if err := sys.Run("go", "build", "-o", path, "."); err != nil {
		return fmt.Errorf("'go build' failed: %w", err)
	}
	return nil
}
