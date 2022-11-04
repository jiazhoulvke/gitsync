A simple git auto sync application

# Installation

```
go install -v github.com/jiazhoulvke/gitsync@latest
```

# Config

```jsonc
{
  "repos": [
    {
      "path": "~/Documents/mynotes", // git repository path, required
      "username": "", // for http auth, default: ""
      "user": "git", // for ssh access, default: "git"
      "private_key_file": "~/.ssh/id_rsa", // for ssh access, default: "~/.ssh/id_rsa"
      "password": "", // default: ""
      "interval": 10, // default: 60
      "include_patterns": [".+\\.md"], // only include markdown files, default: []
      "exclude_patterns": [] // exclude matching files, default: []
    }
  ]
}
// vim: filetype=jsonc
```

# Usage

```
gitsync -c gitsync.json
```
