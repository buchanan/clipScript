# ClipScript

**ClipScript** is a cross-platform utility that plays through custom scripts line-by-line, copying each to your clipboard and optionally executing commands or prompting for input. Originally created to streamline HSM provisioning via serial consoles, it's useful for anyone who follows procedural scripts or frequently pastes text into remote terminals.

## ✨ Features

- ⌨️ Copy script lines to clipboard one-by-one
- 🧠 Prompt for and substitute variables
- 🖥️ Execute external processes or shell scripts
- 📝 Log variable values and output to `log.txt`
- 🔁 Pause after each line for manual advancement
- ⚙️ Cross-platform support (Windows, macOS, Linux)

## 📦 Installation

Pre-built binaries are available in the [GitHub Releases](#).  
Alternatively, clone and build using the included build script:

```bash
git clone https://github.com/youruser/clipscript.git
cd clipscript
./build.sh
```

## 🗂️ Script Format

Scripts are plain text files with the following format:

```text
# This is a description and will be printed but not copied

${VAR_NAME} = Prompt to show the user

!notepad C:\Windows\System32\drivers\etc\hosts  # Launch an external program

echo Hello $VAR_NAME  # This will be copied to clipboard and shown on screen

# Press any key to continue after each line
```

### Example

```text
${Name} = What is your name?
echo Hello $Name
```

## 🚀 Usage

Run ClipScript with a path to your script file:

```bash
clipscript myscript.txt
```

Flags:
- `--version` – Display version and build information

## 🔒 Notes

- Variables are inserted verbatim with no sanitization—use with caution if sensitive inputs are involved.
- Scripts pause for a keypress between each line to ensure you stay in control of command flow.

## 🧰 Dependencies

ClipScript is built in Go and uses:
- [`fatih/color`](https://github.com/fatih/color)
- [`golang.design/x/clipboard`](https://pkg.go.dev/golang.design/x/clipboard)
- [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term)

## 📁 Logs

All variable values and program output are logged to `log.txt`.

## 🧪 Extensibility

While not written as a library, the code is open source and can be easily modified for custom workflows.

## 📄 License

GPL v3.0

---

> Created with love for engineers who hate repeating themselves.
