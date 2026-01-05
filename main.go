package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

var (
	rootPath    string
	dryRun      bool
	skipPrompt  bool
	verbose     bool
	maxWorkers  int
	targetsFile string
	skipWarning bool
)

func init() {
	flag.StringVar(&rootPath, "path", ".", "Root directory to search from")
	flag.BoolVar(&dryRun, "dry-run", false, "Only list directories, don't delete")
	flag.BoolVar(&skipPrompt, "y", false, "Skip confirmation prompt")
	flag.BoolVar(&skipWarning, "skip-warning", false, "Skip the targets file warning")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.IntVar(&maxWorkers, "workers", 4, "Maximum number of concurrent workers")
	flag.StringVar(&targetsFile, "targets", "targets.txt", "File containing directory names to delete")
	flag.Parse()
}

func readTargets(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open targets file: %v", err)
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			targets = append(targets, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading targets file: %v", err)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid targets found in %s", filename)
	}

	return targets, nil
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

	// Read targets from file
	targets, err := readTargets(targetsFile)
	if err != nil {
		log.Fatalf("Error reading targets: %v", err)
	}

	fmt.Printf("🔍 Searching for directories in: %s\n", absRoot)
	fmt.Printf("🎯 Target directories: %s\n", strings.Join(targets, ", "))
	if !skipWarning {
		fmt.Println("⚠️  WARNING: Please review the targets.txt file to ensure you're not accidentally targeting important directories.")
	}
	if dryRun {
		fmt.Println("📋 DRY RUN MODE: No directories will be deleted.")
	}
	if verbose {
		fmt.Println("📢 Verbose mode enabled")
		fmt.Printf("👷 Using %d concurrent workers\n", maxWorkers)
	}

	var dirsToDelete []string
	var scanErrors []string
	var mu sync.Mutex

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if verbose {
				mu.Lock()
				scanErrors = append(scanErrors, fmt.Sprintf("⚠️  Error accessing %s: %v", path, err))
				mu.Unlock()
			}
			return nil // Skip errors and continue
		}

		if !d.IsDir() {
			return nil
		}

		name := d.Name()
		for _, target := range targets {
			if name == target {
				mu.Lock()
				dirsToDelete = append(dirsToDelete, path)
				mu.Unlock()

				fmt.Printf("📁 Found: %s\n", path)

				// Skip walking inside this directory to save time
				return filepath.SkipDir
			}
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
		fmt.Println("✅ No target directories found.")
		return
	}

	// Calculate sizes concurrently
	fmt.Println("\n📊 Calculating directory sizes...")
	var totalSize int64
	var dirSizes []int64
	var sizeErrors []string

	// Worker pool for size calculation
	workChan := make(chan string, len(dirsToDelete))
	resultsChan := make(chan sizeResult, len(dirsToDelete))
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for path := range workChan {
				if verbose {
					fmt.Printf("👷 Worker %d: Calculating size for %s\n", workerID, path)
				}
				size, err := dirSize(path)
				if err != nil {
					resultsChan <- sizeResult{path: path, size: 0, err: err}
					continue
				}
				resultsChan <- sizeResult{path: path, size: size, err: nil}
			}
		}(i)
	}

	// Send work to workers
	for _, dir := range dirsToDelete {
		workChan <- dir
	}
	close(workChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		if result.err != nil {
			if verbose {
				sizeErrors = append(sizeErrors, fmt.Sprintf("⚠️  Could not calculate size of %s: %v", result.path, result.err))
			}
			result.size = 0
		}
		totalSize += result.size
		dirSizes = append(dirSizes, result.size)
	}

	// Report size calculation errors
	if len(sizeErrors) > 0 && verbose {
		fmt.Println("\n📋 Size Calculation Errors:")
		for _, err := range sizeErrors {
			fmt.Println(err)
		}
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

	// Initialize progress bar
	bar := progressbar.NewOptions(
		int(len(dirsToDelete)),
		progressbar.OptionSetDescription("Deleting directories..."),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)

	// Worker pool for deletion
	deleteWorkChan := make(chan int, len(dirsToDelete))
	deleteResultsChan := make(chan deleteResult, len(dirsToDelete))
	wg = sync.WaitGroup{}

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for index := range deleteWorkChan {
				dir := dirsToDelete[index]
				if verbose {
					fmt.Printf("👷 Worker %d: Deleting %s...\n", workerID, dir)
				}
				err := os.RemoveAll(dir)
				if err != nil {
					deleteResultsChan <- deleteResult{index: index, err: err}
					continue
				}
				deleteResultsChan <- deleteResult{index: index, err: nil}
			}
		}(i)
	}

	// Send work to workers
	for i := range dirsToDelete {
		deleteWorkChan <- i
	}
	close(deleteWorkChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(deleteResultsChan)
	}()

	// Collect results
	for result := range deleteResultsChan {
		if result.err != nil {
			errorMsg := fmt.Sprintf("❌ ERROR: %v", result.err)
			fmt.Printf("🗑︸  Deleting: %s %s\n", dirsToDelete[result.index], errorMsg)
			deleteErrors = append(deleteErrors, fmt.Sprintf("%s: %v", dirsToDelete[result.index], result.err))
		} else {
			fmt.Printf("🗑︸  Deleting: %s ✅ Done.\n", dirsToDelete[result.index])
			deletedCount++
			deletedSize += dirSizes[result.index]
		}
		bar.Add(1)
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

// Result types for worker communication
type sizeResult struct {
	path string
	size int64
	err  error
}

type deleteResult struct {
	index int
	err   error
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
