# delPack - Directory Cleaner

A powerful command-line tool to find and delete directories based on a configurable list of target directory names. Originally designed for `node_modules` and `vendor` directories, now fully customizable.

## Features

- **Configurable targets**: Read directory names to delete from a targets file
- **Recursive scanning**: Searches through directories recursively from a specified root path
- **Size calculation**: Calculates and displays the size of each found directory
- **Dry run mode**: Preview what would be deleted without actually removing anything
- **Safety features**: Confirmation prompt before deletion (can be skipped with `-y` flag)
- **Verbose mode**: Detailed output for troubleshooting
- **Error handling**: Gracefully handles permission issues and inaccessible directories
- **Performance optimized**: Skips walking inside target directories to save time
- **Concurrent processing**: Uses multiple workers for faster size calculation and deletion

## Installation

### Prerequisites
- Go 1.16 or higher

### Build from source
```bash
git clone https://github.com/KevinTheSome/delPack.git
cd delPack
go build -o delpack main.go
```

## Usage

```bash
delpack [flags]
```

### Flags
```
  -path string
        Root directory to search from (default ".")
  -targets string
        File containing directory names to delete (default "targets.txt")
  -dry-run
        Only list directories, don't delete
  -y    Skip confirmation prompt
  -v    Verbose output
  -workers int
        Maximum number of concurrent workers (default 4)
```

## Configuration

Create a `targets.txt` file with the directory names you want to delete, one per line:

```txt
# Directory names to delete
node_modules
vendor
dist
build
.cache
.terraform
```

Lines starting with `#` are treated as comments and ignored.

## Examples

1. **Basic usage** (scan current directory with default targets.txt):
```bash
delpack
```

2. **Scan a specific directory**:
```bash
delpack -path /path/to/your/project
```

3. **Dry run** (preview without deleting):
```bash
delpack -dry-run
```

4. **Skip confirmation prompt**:
```bash
delpack -y
```

5. **Verbose output** (shows detailed information):
```bash
delpack -v
```

6. **Use custom targets file**:
```bash
delpack -targets my-custom-targets.txt
```

7. **Combine flags**:
```bash
delpack -path /projects -targets production-targets.txt -dry-run -v
```

8. **Use more workers for faster processing** (on fast storage):
```bash
delpack -workers 8 -v
```

## Output Explanation

- **🔍 Searching**: Shows the root directory being scanned
- **🎯 Target directories**: Lists the directory names being searched for
- **📁 Found**: Lists each discovered target directory with its size
- **📊 Summary**: Provides statistics about found directories and total size
- **🗑️ Deleting**: Shows deletion progress (only in non-dry-run mode)
- **📊 Deletion Results**: Final summary of deleted directories and freed space

## Safety Notes

1. **Always run with `-dry-run` first** to verify what will be deleted
2. The tool skips directories it can't access (shows warnings in verbose mode)
3. Deletion errors are reported but don't stop the entire process
4. Use `-y` flag with caution as it bypasses the confirmation prompt
5. **Carefully review your targets.txt file** to ensure you're not accidentally targeting important directories

## Performance Considerations

- Large directories may take time to scan and calculate sizes
- The tool skips walking inside target directories after finding them to improve performance
- Verbose mode adds overhead but provides useful debugging information
- The `-workers` flag controls concurrency level (default: 4 workers)
- Higher worker counts improve performance on SSDs and fast storage

## Default Targets

The default `targets.txt` includes common directories that are often safe to delete:
- `node_modules` - Node.js dependencies
- `vendor` - PHP dependencies
- `dist` - Build output directories
- `build` - Build directories
- `.cache` - Cache directories
- `.terraform` - Terraform cache

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Support

For issues or questions, please open an issue on the GitHub repository.