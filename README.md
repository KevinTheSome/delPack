# delPack - Node Modules and Vendor Directory Cleaner

A powerful command-line tool to find and delete `node_modules` and `vendor` directories, helping you reclaim valuable disk space.

## Features

- **Recursive scanning**: Searches through directories recursively from a specified root path
- **Size calculation**: Calculates and displays the size of each found directory
- **Dry run mode**: Preview what would be deleted without actually removing anything
- **Safety features**: Confirmation prompt before deletion (can be skipped with `-y` flag)
- **Verbose mode**: Detailed output for troubleshooting
- **Error handling**: Gracefully handles permission issues and inaccessible directories
- **Performance optimized**: Skips walking inside target directories to save time

## Installation

### Prerequisites
- Go 1.16 or higher

### Build from source
```bash
git clone https://github.com/yourusername/delPack.git
cd delPack
go build -o delpack main.go
```

### Install globally
```bash
go install github.com/yourusername/delPack@latest
```

## Usage

```bash
delpack [flags]
```

### Flags
```
  -path string
        Root directory to search from (default ".")
  -dry-run
        Only list directories, don't delete
  -y    Skip confirmation prompt
  -v    Verbose output
```

## Examples

1. **Basic usage** (scan current directory):
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

6. **Combine flags**:
```bash
delpack -path /projects -dry-run -v
```

## Output Explanation

- **🔍 Searching**: Shows the root directory being scanned
- **📁 Found**: Lists each discovered `node_modules` or `vendor` directory with its size
- **📊 Summary**: Provides statistics about found directories and total size
- **🗑️ Deleting**: Shows deletion progress (only in non-dry-run mode)
- **📊 Deletion Results**: Final summary of deleted directories and freed space

## Safety Notes

1. **Always run with `-dry-run` first** to verify what will be deleted
2. The tool skips directories it can't access (shows warnings in verbose mode)
3. Deletion errors are reported but don't stop the entire process
4. Use `-y` flag with caution as it bypasses the confirmation prompt

## Performance Considerations

- Large directories may take time to scan and calculate sizes
- The tool skips walking inside target directories after finding them to improve performance
- Verbose mode adds overhead but provides useful debugging information

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Support

For issues or questions, please open an issue on the GitHub repository.