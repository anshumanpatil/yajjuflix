# Movie Name Cleaner / Renamer

A utility script to clean buzzwords and metadata from movie filenames.

## Configuration
Edit [config.json](config.json) to set your movie directory, buzzwords, separators, and allowed extensions.

## Usage

### Go
```bash
# Preview changes (Dry run)
go run main.go

# Interactive rename
go run main.go rename
# or
go run main.go -rename

# Auto-confirm all renames
go run main.go rename -y
```

### Python
```bash
# Preview changes (Dry run)
python3 main.py

# Interactive rename
python3 main.py rename
# or
python3 main.py --rename

# Auto-confirm all renames
python3 main.py rename -y
```