package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	rootPath   string
	dryRun     bool
	skipPrompt bool
	verbose    bool
)

func init() {
	flag.StringVar(&rootPath, "path", ".", "Root directory to search from")
	flag.BoolVar(&dryRun, "dry-run", false, "Only list directories, don't delete")
	flag.BoolVar(&skipPrompt, "y", false, "Skip confirmation prompt")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()
}

func main() {
	startTime := time.Now()
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		log.Fatalf("Invalid path: %v", err)
	}

	// Verify the root path exists
	if _, err := os.Stat(absRoot); os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", absRoot)
	}

	fmt.Printf("🔍 Searching for node_modules and vendor directories in: %s\n", absRoot)
	if dryRun {
		fmt.Println("📋 DRY RUN MODE: No directories will be deleted.")
	}
	if verbose {
		fmt.Println("📢 Verbose mode enabled")
	}

	var totalSize int64
	var dirsToDelete []string
	var dirSizes []int64
	var scanErrors []string

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if verbose {
				scanErrors = append(scanErrors, fmt.Sprintf("⚠️  Error accessing %s: %v", path, err))
			}
			return nil // Skip errors and continue
		}

		if !d.IsDir() {
			return nil
		}

		name := d.Name()
		if name == "node_modules" || name == "vendor" {
			// Calculate directory size
			if verbose {
				fmt.Printf("📊 Calculating size for: %s\n", path)
			}

			size, err := dirSize(path)
			if err != nil {
				if verbose {
					fmt.Printf("⚠️  Could not calculate size of %s: %v\n", path, err)
				}
				size = 0
			}

			totalSize += size
			dirsToDelete = append(dirsToDelete, path)
			dirSizes = append(dirSizes, size)

			fmt.Printf("📁 Found: %s (%s)\n", path, formatBytes(size))

			// Skip walking inside this directory to save time
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		log.Fatalf("❌ Error walking directory: %v", err)
	}

	// Report any scan errors
	if len(scanErrors) > 0 && verbose {
		fmt.Println("\n📋 Scan Errors:")
		for _, err := range scanErrors {
			fmt.Println(err)
		}
	}

	if len(dirsToDelete) == 0 {
		fmt.Println("✅ No node_modules or vendor directories found.")
		return
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   • Directories found: %d\n", len(dirsToDelete))
	fmt.Printf("   • Total size: %s\n", formatBytes(totalSize))
	fmt.Printf("   • Scan duration: %v\n", time.Since(startTime))

	if dryRun {
		fmt.Println("🏁 Dry run completed successfully.")
		return
	}

	if !skipPrompt {
		fmt.Print("\n⚠️  Are you sure you want to delete these directories? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("🛑 Operation cancelled by user.")
			return
		}
	}

	fmt.Println("\n🗑️  Starting deletion process...")
	var deletedSize int64
	var deletedCount int
	var deleteErrors []string

	for i, dir := range dirsToDelete {
		fmt.Printf("🗑︸  Deleting: %s ... ", dir)
		err := os.RemoveAll(dir)
		if err != nil {
			errorMsg := fmt.Sprintf("❌ ERROR: %v", err)
			fmt.Println(errorMsg)
			deleteErrors = append(deleteErrors, fmt.Sprintf("%s: %v", dir, err))
		} else {
			fmt.Println("✅ Done.")
			deletedCount++
			deletedSize += dirSizes[i]
		}
	}

	// Report deletion errors if any
	if len(deleteErrors) > 0 {
		fmt.Println("\n⚠️  Deletion Errors:")
		for _, err := range deleteErrors {
			fmt.Println(err)
		}
	}

	fmt.Printf("\n📊 Deletion Results:\n")
	fmt.Printf("   • Successfully deleted: %d out of %d directories\n", deletedCount, len(dirsToDelete))
	fmt.Printf("   • Freed space: %s\n", formatBytes(deletedSize))
	fmt.Printf("   • Total operation time: %v\n", time.Since(startTime))

	if deletedCount > 0 {
		fmt.Println("🎉 Operation completed successfully!")
	} else {
		fmt.Println("❌ No directories were deleted.")
	}
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
