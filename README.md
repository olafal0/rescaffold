# rescaffold

Rescaffold is a project scaffolding generator and migration tool. Unlike other tools which can only create a template once, rescaffold can update scaffolding in-place without mangling any of your code.

## Status

- [x] Initial template generation
- [x] Interactive variable setting
- [ ] Escape sequences
- [x] Loading scaffolds from a local directory
- [x] Scaffold lockfile
- [x] Scaffold checksums and conflict detection on first generation
- [x] Basic template modifiers (only one per variable instance in a template)
- [x] Scaffold upgrades
- [x] Scaffold removal
- [x] Auto-clone scaffold sources from git
- [ ] Version tracking of git sources
- [x] Composable modifiers
- [ ] Method of fixing conflicts (interactively or not)

## Usage

From the root directory of a new or existing project, run

`rescaffold <git-template-url>`

This will clone the git repository and interactively run first template setup. Rescaffold will perform any template generation tasks (e.g. name directories according to your project name), and unpack new files into your project. If there are any conflicting files, rescaffold will gracefully back out and return your project to its original state.

Rescaffold automatically tracks the template source and version information of the scaffold(s) you use, so if you want to use the newest version of a scaffold, you can run:

`rescaffold -upgrade`

or

`rescaffold -upgrade <git-template-url>` to upgrade a specific scaffold.

You can add as many scaffolds as you want, simply by repeating the initial command. If you want to remove a scaffold (which only removes files created by rescaffold and are since untouched), you can run:

`rescaffold -remove <git-template-url>`

Scaffolds can be:

- URLs of git repositories
- Relative or absolute paths to locally stored directories

This means you can develop scaffolds without going through a git remote, and also that you can clone a repo yourself if your setup requires more than an unauthenticated `git clone`.

## `.rescaffold.toml`

`.rescaffold.toml` is a file that rescaffold will place in the working directory when you first run it. This toml file tracks which scaffolds are in place in your project, their versions, their sources, and the list of files that they have placed, along with their checksums. This file is used by rescaffold to avoid overwriting any files or directories that were not created by rescaffold, so it should be committed along with the rest of your code.

If `.rescaffold.toml` gets deleted, rescaffold will need to be run interactively to resolve any conflicts that arise, and any files that need to be updated will have to be checked manually.

Here's an example of `.rescaffold.toml` created when generating using the scaffold in `example/`:

```toml
[scaffolds]
[scaffolds.example]
source = "example"

[[scaffolds.example.file]]
path = "ext/go.mod"
checksum = "5dc2e5fcdd34653232d8b2efbfa1c050c49f9002cae1b6093ad2f53d8af8b4c9"

[[scaffolds.example.file]]
path = "ext/web/index.html"
checksum = "dd0a2bb1d85c01a876d5e766af7f7bd714c637527776ee7b93f3dcd6e6497945"

[[scaffolds.example.file]]
path = "ext/foo.go"
checksum = "b5399904673d9c9209f54163c9625d90ff02c16e886c0066e1eff0401683cae5"
[scaffolds.example.vars]
name = "foo"
port = "8000"
```

## Creating Scaffolds

Scaffolds are directories with a `.rescaffold-manifest.toml` file at the root. They can be stored in a VCS, like git, or live as a directory on your local filesystem. The manifest file looks like:

```toml
rescaffold_version = "0"

[meta]
title = "Example Scaffold"
author = "Firstname Lastname <me@example.com>"
description = "An example scaffold"
post_install = "Some text instructions to print after generation"

[config]
open_delim = "_"
close_delim = "_"
modifier_delim = "|"

[vars]
[vars.project_name]
type = "string"
description = "A short, descriptive name for your project"

[vars.postgres_version]
type = "enum"
enum_values = ["12", "13", "14", "15"]
default = "15"
description = "Postgres version to use in tools"
```

Directory names, file names, and file contents can all use values from `vars` as needed. For example, your directory structure could be:

```
_project_name_
├── _project_name__config.go
└── _project_name_.go
```

And `_project_name_.go` might contain

```go
package _project_name_

func PrintWelcome() {
  fmt.Println("Welcome to my project, _project_name_!")
}
```

If your project name was "foobar", this would be generated as the file `foobar.go`, containing

```go
package foobar

func PrintWelcome() {
  fmt.Println("Welcome to my project, foobar!")
}
```

## Delimiters

Rescaffold uses delimiters for template replacement by searching for instances of `open_delim + var_name + close_delim`, for all variable names, and replacing those substrings with the actual value the var is set to. Also, occurrences of `open_delim + "\" + var_name + close_delim` will be replaced by the same string with the backslash removed, to allow predictable escaping.

You can edit the delimiters used for template replacement to whatever makes life easier for your scaffold. For example, if a scaffold contains a lot of HTML, using `<>` delimiters for replacement might cause problems. In that case, you might prefer delimiters of `${` and `}` instead. Also, you can always follow opening delimiters with a backslash to escape replacement, e.g. `<\title>` will be replaced by the literal string `<title>`.

Delimiters, both opening and closing, are also optional. For example, you could set an opening delimiter of `MY_SCAFFOLD_`, and an empty closing delimiter. This means that template replacement would replace all instances of `MY_SCAFFOLD_title` with the string value of the `title` var. Delimiters won't be exposed to users of your scaffold—they will only interact with the end result.

## Modifiers

Replacement substrings can also contain modifiers, such as `_name|titleCase_`. These modifiers can change the var value before performing replacement. The modifiers that are available are:

- `titlecase`: "some string" -> "Some String"
- `lowercase`: "Foo" -> "foo"
- `uppercase`: "Foo" -> "FOO"
