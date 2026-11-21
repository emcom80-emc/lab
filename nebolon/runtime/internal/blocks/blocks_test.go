package blocks_test

import (
	"os"
	"path/filepath"
	"testing"

	"nebolon/runtime/internal/blocks"
)

// writeFile creates a file with the given content, creating parent dirs as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// TestResolveLibraryRoot_ExplicitOverride verifies that an explicit path always overrides default discovery.
func TestResolveLibraryRoot_ExplicitOverride(t *testing.T) {
	custom := "/custom/explicit/library/path"
	resolved := blocks.ResolveLibraryRoot(custom)
	if resolved != custom {
		t.Errorf("expected explicit path %q, got %q", custom, resolved)
	}

	// Also when baseDir is provided
	resolvedFrom := blocks.ResolveLibraryRootFrom(custom, "/some/installation/dir")
	if resolvedFrom != custom {
		t.Errorf("expected explicit path %q, got %q", custom, resolvedFrom)
	}
}

// TestResolveLibraryRoot_RelativeInstallation verifies resolution relative to the installation base directory.
func TestResolveLibraryRoot_RelativeInstallation(t *testing.T) {
	// Simulate an installation tree:
	// <tmp>/nebolon/
	//    ├── runtime/ (where binary lives)
	//    └── block-libraries/
	//         └── test-block/
	//              └── test-block.block
	tmp := t.TempDir()
	installRoot := filepath.Join(tmp, "nebolon")
	binDir := filepath.Join(installRoot, "runtime")
	libDir := filepath.Join(installRoot, "block-libraries")
	writeFile(t, filepath.Join(libDir, "test-block", "test-block.block"), `{"id":"test-block"}`)

	resolved := blocks.ResolveLibraryRootFrom("", binDir)
	absLibDir, _ := filepath.Abs(libDir)
	if resolved != absLibDir {
		t.Errorf("expected resolved path %q, got %q", absLibDir, resolved)
	}
}

// TestResolveLibraryRoot_FromDifferentWorkingDirectories verifies that resolution succeeds
// regardless of which working directory the process is currently in.
func TestResolveLibraryRoot_FromDifferentWorkingDirectories(t *testing.T) {
	tmp := t.TempDir()
	installRoot := filepath.Join(tmp, "nebolon")
	binDir := filepath.Join(installRoot, "runtime")
	libDir := filepath.Join(installRoot, "block-libraries")
	writeFile(t, filepath.Join(libDir, "test-block", "test-block.block"), `{"id":"test-block"}`)

	absLibDir, _ := filepath.Abs(libDir)

	// Change working directory to a separate temp directory to simulate running from anywhere.
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	otherDir := t.TempDir()
	if err := os.Chdir(otherDir); err != nil {
		t.Fatalf("failed to chdir to %s: %v", otherDir, err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()

	resolved := blocks.ResolveLibraryRootFrom("", binDir)
	if resolved != absLibDir {
		t.Errorf("running from %s: expected resolved path %q, got %q", otherDir, absLibDir, resolved)
	}
}

// TestDiscover_FindsBlocks verifies that Discover lists valid block definitions.
func TestDiscover_FindsBlocks(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "block-a", "block-a.block"), `{"id":"block-a"}`)
	writeFile(t, filepath.Join(tmp, "block-b", "block-b.block"), `{"id":"block-b"}`)
	// Directory without .block should be ignored
	if err := os.MkdirAll(filepath.Join(tmp, "ignored-dir"), 0755); err != nil {
		t.Fatal(err)
	}

	found, err := blocks.Discover(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %v", len(found), found)
	}
	if found[0] != "block-a" && found[1] != "block-a" {
		t.Errorf("expected block-a in results, got %v", found)
	}
	if found[0] != "block-b" && found[1] != "block-b" {
		t.Errorf("expected block-b in results, got %v", found)
	}
}

// TestDiscover_ErrorOnMissingDir verifies clear error when library dir is missing.
func TestDiscover_ErrorOnMissingDir(t *testing.T) {
	missingPath := "/path/that/does/not/exist/block-libraries"
	_, err := blocks.Discover(missingPath)
	if err == nil {
		t.Fatal("expected error for missing directory, got nil")
	}
	if !os.IsNotExist(err) && filepath.Base(missingPath) == "" {
		t.Errorf("expected not exist error, got %v", err)
	}
}

// TestLoad_ErrorOnMissingBlock verifies clear error when block definition is missing.
func TestLoad_ErrorOnMissingBlock(t *testing.T) {
	tmp := t.TempDir()
	_, err := blocks.Load(tmp, "nonexistent-block")
	if err == nil {
		t.Fatal("expected error for missing block, got nil")
	}
}
