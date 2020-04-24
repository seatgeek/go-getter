package getter

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestSmb_ClientGet(t *testing.T) {
	smbTestsPreCheck(t)

	tests := []struct {
		name    string
		rawURL  string
		mode    Mode
		file    string
		mounted bool
		fail    bool
	}{
		{
			"smb scheme subdir with registered authentication in private share",
			"smb://user:password@samba/private/subdir",
			ModeDir,
			"file.txt",
			false,
			false,
		},
		{
			"smb scheme file with registered authentication with file in private share",
			"smb://user:password@samba/private/subdir/file.txt",
			ModeFile,
			"file.txt",
			false,
			false,
		},
		{
			"smb scheme file without authentication in public share",
			"smb://samba/public/subdir/file.txt",
			ModeFile,
			"file.txt",
			false,
			false,
		},
		{
			"// prefix path local mounted smb shared file",
			"//mnt/file.txt",
			ModeFile,
			"file.txt",
			true,
			false,
		},
		{
			"// prefix path local mounted smb shared directory",
			"//mnt/subdir",
			ModeDir,
			"file.txt",
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := tempDir(t)
			defer os.RemoveAll(dst)

			if tt.mode == ModeFile {
				dst = filepath.Join(dst, tt.file)
			}

			log.Printf("MOSS dst %s", dst)

			req := &Request{
				Dst:  dst,
				Src:  tt.rawURL,
				Mode: tt.mode,
			}

			result, err := DefaultClient.Get(context.Background(), req)

			fail := err != nil
			if tt.fail != fail {
				if fail {
					t.Fatalf("err: unexpected error %s", err.Error())
				}
				t.Fatalf("err: expecting to fail but it did not")
			}

			if !tt.fail {
				if result == nil {
					t.Fatalf("err: get result should not be nil")
				}
				if result.Dst != dst {
					t.Fatalf("err: expected destination: %s \n actual destination: %s", dst, result.Dst)
				}
				if tt.mounted && tt.mode == ModeDir {
					// Verify the destination folder is a symlink to the mounted one
					fi, err := os.Lstat(dst)
					if err != nil {
						t.Fatalf("err: %s", err)
					}
					if fi.Mode()&os.ModeSymlink == 0 {
						t.Fatal("destination is not a symlink")
					}
				}
				if tt.mode == ModeDir {
					dst = filepath.Join(dst, tt.file)
				}
				// Verify if the file was successfully downloaded
				// and exists at the destination folder
				assertContents(t, dst, "Hello\n")
			}
		})
	}
}
